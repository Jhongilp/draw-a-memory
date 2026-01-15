package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HandleUpload handles photo upload requests
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with max size
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("ParseMultipartForm error: %v", err)
		SendError(w, "Failed to parse upload. Maximum total size is 50MB", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	files := r.MultipartForm.File["photos"]
	if len(files) == 0 {
		SendError(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	if len(files) > maxPhotoCount {
		SendError(w, fmt.Sprintf("Too many files. Maximum is %d photos per upload", maxPhotoCount), http.StatusBadRequest)
		return
	}

	var uploadedPhotos []Photo

	for _, fileHeader := range files {
		// Validate file type
		if !isValidImageType(fileHeader.Filename) {
			continue
		}

		// Validate individual file size
		if fileHeader.Size > maxFileSize {
			log.Printf("File %s exceeds max size (%d > %d bytes)", fileHeader.Filename, fileHeader.Size, maxFileSize)
			continue
		}

		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening file %s: %v", fileHeader.Filename, err)
			continue
		}
		defer file.Close()

		// Generate unique filename
		ext := filepath.Ext(fileHeader.Filename)
		photoID := uuid.New().String()
		newFilename := photoID + ext
		filePath := filepath.Join(uploadDir, newFilename)

		// Create destination file
		dst, err := os.Create(filePath)
		if err != nil {
			log.Printf("Error creating file %s: %v", filePath, err)
			continue
		}
		defer dst.Close()

		// Copy file content
		size, err := io.Copy(dst, file)
		if err != nil {
			log.Printf("Error saving file %s: %v", filePath, err)
			os.Remove(filePath)
			continue
		}

		photo := Photo{
			ID:         photoID,
			Filename:   fileHeader.Filename,
			Path:       "/uploads/" + newFilename,
			Size:       size,
			UploadedAt: time.Now(),
		}
		log.Printf("Saving photo %s, %s", photoID, fileHeader.Filename)

		uploadedPhotos = append(uploadedPhotos, photo)
		log.Printf("Uploaded: %s -> %s (%d bytes)", fileHeader.Filename, newFilename, size)
	}

	if len(uploadedPhotos) == 0 {
		SendError(w, "No valid images were uploaded", http.StatusBadRequest)
		return
	}

	response := UploadResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully uploaded %d photo(s)", len(uploadedPhotos)),
		Photos:  uploadedPhotos,
	}

	SendJSON(w, response)
}

// HandleGetPhotos returns all uploaded photos
func HandleGetPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		SendError(w, "Failed to read photos directory", http.StatusInternalServerError)
		return
	}

	var photos []Photo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !isValidImageType(file.Name()) {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Extract ID from filename (remove extension)
		id := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		photo := Photo{
			ID:         id,
			Filename:   file.Name(),
			Path:       "/uploads/" + file.Name(),
			Size:       info.Size(),
			UploadedAt: info.ModTime(),
		}
		photos = append(photos, photo)
	}

	SendJSON(w, photos)
}

// HandleServePhoto serves photo files
func HandleServePhoto(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/uploads/")
	filePath := filepath.Join(uploadDir, filename)

	// Security: prevent directory traversal
	if strings.Contains(filename, "..") {
		SendError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, filePath)
}

// HandleClusterPhotos analyzes photos using Gemini AI and groups them into clusters
func HandleClusterPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.PhotoIds) == 0 {
		SendError(w, "No photo IDs provided", http.StatusBadRequest)
		return
	}

	// Get photo file paths
	var photoPaths []string
	for _, photoID := range req.PhotoIds {
		files, err := os.ReadDir(uploadDir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if strings.HasPrefix(file.Name(), photoID) {
				photoPaths = append(photoPaths, filepath.Join(uploadDir, file.Name()))
				break
			}
		}
	}

	if len(photoPaths) == 0 {
		SendError(w, "No valid photos found", http.StatusBadRequest)
		return
	}

	// Use Gemini AI to analyze and cluster photos
	clusters, err := AnalyzeAndClusterPhotos(req.PhotoIds, photoPaths)
	if err != nil {
		log.Printf("Error clustering photos: %v", err)
		SendError(w, "Failed to analyze photos", http.StatusInternalServerError)
		return
	}

	// Generate background images for each cluster and create drafts
	var pageDrafts []PageDraft
	for i, cluster := range clusters {
		// Generate themed background image
		backgroundPath, err := GenerateBackgroundImage(cluster.Theme, cluster.Title, cluster.Description)
		if err != nil {
			log.Printf("Failed to generate background for cluster %s: %v", cluster.ID, err)
			// Continue without background - it's optional
		} else {
			clusters[i].BackgroundPath = backgroundPath
		}

		draft := PageDraft{
			ID:             uuid.New().String(),
			ClusterID:      cluster.ID,
			PhotoIds:       cluster.PhotoIds,
			Title:          cluster.Title,
			Description:    cluster.Description,
			Theme:          cluster.Theme,
			BackgroundPath: backgroundPath,
			Status:         "draft",
			CreatedAt:      time.Now().Format(time.RFC3339),
		}
		drafts[draft.ID] = draft
		pageDrafts = append(pageDrafts, draft)
	}

	response := ClusterResponse{
		Clusters: clusters,
		Drafts:   pageDrafts,
	}

	SendJSON(w, response)
}

// isValidImageType checks if the file has a valid image extension
func isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".heic", ".heif"}

	for _, valid := range validExtensions {
		if ext == valid {
			return true
		}
	}
	return false
}
