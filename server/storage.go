package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

const (
	thumbWidth      = 800 // Width for thumbnails
	thumbHeight     = 600 // Max height for thumbnails
	signedURLExpiry = 15 * time.Minute
)

// Storage handles Google Cloud Storage operations
type Storage struct {
	client    *storage.Client
	bucket    string
	projectID string
}

// NewStorage creates a new GCS storage client
func NewStorage(ctx context.Context, projectID, bucket string) (*Storage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &Storage{
		client:    client,
		bucket:    bucket,
		projectID: projectID,
	}, nil
}

// Close closes the storage client
func (s *Storage) Close() error {
	return s.client.Close()
}

// UploadPhoto uploads a photo to GCS and returns the object path
func (s *Storage) UploadPhoto(ctx context.Context, userID string, file io.Reader, filename string, contentType string) (gcsPath string, thumbPath string, sizeBytes int64, err error) {
	// Generate unique filename
	ext := filepath.Ext(filename)
	photoID := uuid.New().String()
	objectName := fmt.Sprintf("photos/%s/%s%s", userID, photoID, ext)

	// Read file into buffer for processing
	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to read file: %w", err)
	}
	sizeBytes = int64(len(data))

	// Upload original photo
	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(objectName)

	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType
	writer.CacheControl = "private, max-age=31536000" // Private caching for sensitive content

	if _, err := writer.Write(data); err != nil {
		return "", "", 0, fmt.Errorf("failed to write to GCS: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", "", 0, fmt.Errorf("failed to close GCS writer: %w", err)
	}

	log.Printf("Uploaded photo to GCS: %s", objectName)

	// Generate and upload thumbnail
	thumbObjectName, err := s.generateAndUploadThumbnail(ctx, userID, photoID, data)
	if err != nil {
		log.Printf("Warning: failed to generate thumbnail: %v", err)
		// Continue without thumbnail
	} else {
		thumbPath = thumbObjectName
	}

	return objectName, thumbPath, sizeBytes, nil
}

// generateAndUploadThumbnail creates a thumbnail and uploads it to GCS
func (s *Storage) generateAndUploadThumbnail(ctx context.Context, userID, photoID string, imageData []byte) (string, error) {
	// Decode image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Generate thumbnail
	thumb := imaging.Fit(img, thumbWidth, thumbHeight, imaging.Lanczos)

	// Encode thumbnail as JPEG
	var thumbBuf bytes.Buffer
	if err := jpeg.Encode(&thumbBuf, thumb, &jpeg.Options{Quality: 85}); err != nil {
		return "", fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	// Upload thumbnail
	thumbObjectName := fmt.Sprintf("photos/%s/%s_thumb.jpg", userID, photoID)
	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(thumbObjectName)

	writer := obj.NewWriter(ctx)
	writer.ContentType = "image/jpeg"
	writer.CacheControl = "private, max-age=31536000"

	if _, err := writer.Write(thumbBuf.Bytes()); err != nil {
		return "", fmt.Errorf("failed to write thumbnail to GCS: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close thumbnail writer: %w", err)
	}

	log.Printf("Uploaded thumbnail to GCS: %s", thumbObjectName)
	return thumbObjectName, nil
}

// UploadBackground uploads a generated background image to GCS
func (s *Storage) UploadBackground(ctx context.Context, userID string, imageData []byte, theme string) (string, error) {
	objectName := fmt.Sprintf("backgrounds/%s/%s_%s.png", userID, theme, uuid.New().String()[:8])

	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(objectName)

	writer := obj.NewWriter(ctx)
	writer.ContentType = "image/png"
	writer.CacheControl = "private, max-age=31536000"

	if _, err := writer.Write(imageData); err != nil {
		return "", fmt.Errorf("failed to write background to GCS: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close background writer: %w", err)
	}

	log.Printf("Uploaded background to GCS: %s", objectName)
	return objectName, nil
}

// GetSignedURL generates a signed URL for private GCS access
func (s *Storage) GetSignedURL(ctx context.Context, objectPath string) (string, error) {
	if objectPath == "" {
		return "", fmt.Errorf("empty object path")
	}

	// Clean the path (remove leading slash if present)
	objectPath = strings.TrimPrefix(objectPath, "/")

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(signedURLExpiry),
	}

	url, err := s.client.Bucket(s.bucket).SignedURL(objectPath, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// DeletePhoto deletes a photo and its thumbnail from GCS
func (s *Storage) DeletePhoto(ctx context.Context, gcsPath, thumbPath string) error {
	bucket := s.client.Bucket(s.bucket)

	// Delete original
	if gcsPath != "" {
		if err := bucket.Object(gcsPath).Delete(ctx); err != nil && err != storage.ErrObjectNotExist {
			log.Printf("Warning: failed to delete original photo %s: %v", gcsPath, err)
		}
	}

	// Delete thumbnail
	if thumbPath != "" {
		if err := bucket.Object(thumbPath).Delete(ctx); err != nil && err != storage.ErrObjectNotExist {
			log.Printf("Warning: failed to delete thumbnail %s: %v", thumbPath, err)
		}
	}

	return nil
}

// DownloadToBuffer downloads a GCS object to a buffer (for AI analysis)
func (s *Storage) DownloadToBuffer(ctx context.Context, objectPath string) ([]byte, error) {
	bucket := s.client.Bucket(s.bucket)
	reader, err := bucket.Object(objectPath).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return data, nil
}

// GetImageDimensions returns the width and height of an image
func GetImageDimensions(data []byte) (width, height int, err error) {
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}
	return img.Width, img.Height, nil
}

// ValidateContentType checks if the content type is a valid image type
func ValidateContentType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"image/heic": true,
		"image/heif": true,
	}
	return validTypes[contentType]
}

// ValidateFileExtension checks if the file extension is valid
func ValidateFileExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".heic": true,
		".heif": true,
	}
	return validExtensions[ext]
}
