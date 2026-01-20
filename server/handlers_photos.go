package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxFileSize   = 5 << 20  // 5 MB per file
	maxTotalSize  = 50 << 20 // 50 MB total
	maxPhotoCount = 10
)

// App holds the application dependencies
type App struct {
	config  *Config
	db      *Database
	storage *Storage
	auth    *AuthMiddleware
}

// HandleUpload handles photo upload requests
func (app *App) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user
	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		SendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get or create database user
	ctx := r.Context()
	dbUser, err := app.db.GetOrCreateUser(ctx, authUser.ClerkID, authUser.Email, authUser.Name)
	if err != nil {
		log.Printf("Failed to get/create user: %v", err)
		SendError(w, "Failed to process user", http.StatusInternalServerError)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(maxTotalSize); err != nil {
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
		// Validate file extension
		if !ValidateFileExtension(fileHeader.Filename) {
			log.Printf("Invalid file type: %s", fileHeader.Filename)
			continue
		}

		// Validate file size
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

		// Determine content type
		contentType := fileHeader.Header.Get("Content-Type")
		if !ValidateContentType(contentType) {
			contentType = "image/jpeg" // Default fallback
		}

		// Upload to GCS
		gcsPath, thumbPath, sizeBytes, err := app.storage.UploadPhoto(ctx, dbUser.ID, file, fileHeader.Filename, contentType)
		if err != nil {
			log.Printf("Error uploading file %s: %v", fileHeader.Filename, err)
			continue
		}

		// Generate unique ID
		photoID := uuid.New().String()

		// Create database record
		dbPhoto := &DBPhoto{
			ID:               photoID,
			UserID:           dbUser.ID,
			Filename:         gcsPath,
			OriginalFilename: fileHeader.Filename,
			GCSPath:          gcsPath,
			ThumbGCSPath:     sql.NullString{String: thumbPath, Valid: thumbPath != ""},
			SizeBytes:        sizeBytes,
			ContentType:      contentType,
		}

		if err := app.db.CreatePhoto(ctx, dbPhoto); err != nil {
			log.Printf("Error saving photo to database: %v", err)
			// Try to clean up GCS files
			app.storage.DeletePhoto(ctx, gcsPath, thumbPath)
			continue
		}

		// Generate signed URL for response
		signedURL, _ := app.storage.GetSignedURL(ctx, gcsPath)

		photo := Photo{
			ID:         photoID,
			Filename:   fileHeader.Filename,
			Path:       signedURL,
			Size:       sizeBytes,
			UploadedAt: time.Now(),
		}
		uploadedPhotos = append(uploadedPhotos, photo)
		log.Printf("Uploaded: %s -> %s (%d bytes)", fileHeader.Filename, gcsPath, sizeBytes)
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

// HandleGetPhotos returns all photos for the authenticated user
func (app *App) HandleGetPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user
	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		SendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Get database user
	dbUser, err := app.db.GetOrCreateUser(ctx, authUser.ClerkID, authUser.Email, authUser.Name)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		SendError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Get photos from database
	dbPhotos, err := app.db.GetPhotosByUser(ctx, dbUser.ID)
	if err != nil {
		log.Printf("Failed to get photos: %v", err)
		SendError(w, "Failed to retrieve photos", http.StatusInternalServerError)
		return
	}

	// Convert to API format with signed URLs
	var photos []Photo
	for _, dbPhoto := range dbPhotos {
		signedURL, err := app.storage.GetSignedURL(ctx, dbPhoto.GCSPath)
		if err != nil {
			log.Printf("Failed to generate signed URL for %s: %v", dbPhoto.ID, err)
			continue
		}

		var thumbURL string
		if dbPhoto.ThumbGCSPath.Valid {
			thumbURL, _ = app.storage.GetSignedURL(ctx, dbPhoto.ThumbGCSPath.String)
		}

		// Use thumb URL as default path for gallery view
		path := signedURL
		if thumbURL != "" {
			path = thumbURL
		}

		photos = append(photos, Photo{
			ID:         dbPhoto.ID,
			Filename:   dbPhoto.OriginalFilename,
			Path:       path,
			Size:       dbPhoto.SizeBytes,
			UploadedAt: dbPhoto.CreatedAt,
		})
	}

	SendJSON(w, photos)
}

// HandleGetPhotoURL returns a signed URL for a specific photo
func (app *App) HandleGetPhotoURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user
	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		SendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Extract photo ID from path
	photoID := strings.TrimPrefix(r.URL.Path, "/api/photos/")
	photoID = strings.TrimSuffix(photoID, "/url")
	if photoID == "" {
		SendError(w, "Photo ID required", http.StatusBadRequest)
		return
	}

	// Get photo from database
	dbPhoto, err := app.db.GetPhotoByID(ctx, photoID)
	if err != nil {
		SendError(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Get database user to verify ownership
	dbUser, err := app.db.GetOrCreateUser(ctx, authUser.ClerkID, authUser.Email, authUser.Name)
	if err != nil || dbPhoto.UserID != dbUser.ID {
		SendError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Check if thumbnail is requested
	useThumb := r.URL.Query().Get("thumb") == "1"

	var signedURL string
	if useThumb && dbPhoto.ThumbGCSPath.Valid {
		signedURL, err = app.storage.GetSignedURL(ctx, dbPhoto.ThumbGCSPath.String)
	} else {
		signedURL, err = app.storage.GetSignedURL(ctx, dbPhoto.GCSPath)
	}

	if err != nil {
		log.Printf("Failed to generate signed URL: %v", err)
		SendError(w, "Failed to generate URL", http.StatusInternalServerError)
		return
	}

	SendJSON(w, map[string]string{"url": signedURL})
}

// HandleDeletePhoto handles photo deletion
func (app *App) HandleDeletePhoto(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		SendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Extract photo ID
	photoID := strings.TrimPrefix(r.URL.Path, "/api/photos/")
	if photoID == "" {
		SendError(w, "Photo ID required", http.StatusBadRequest)
		return
	}

	// Get database user
	dbUser, err := app.db.GetOrCreateUser(ctx, authUser.ClerkID, authUser.Email, authUser.Name)
	if err != nil {
		SendError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Get photo for GCS cleanup
	dbPhoto, err := app.db.GetPhotoByID(ctx, photoID)
	if err != nil {
		SendError(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if dbPhoto.UserID != dbUser.ID {
		SendError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Soft delete from database
	if err := app.db.SoftDeletePhoto(ctx, dbUser.ID, photoID); err != nil {
		SendError(w, "Failed to delete photo", http.StatusInternalServerError)
		return
	}

	// Note: We do NOT delete from GCS immediately to allow recovery
	// A background job should clean up soft-deleted photos after a retention period

	SendJSON(w, map[string]bool{"success": true})
}

// HandleClusterPhotos analyzes photos using Gemini AI and groups them into clusters
func (app *App) HandleClusterPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		SendError(w, "Unauthorized", http.StatusUnauthorized)
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

	ctx := r.Context()

	// Get database user
	dbUser, err := app.db.GetOrCreateUser(ctx, authUser.ClerkID, authUser.Email, authUser.Name)
	if err != nil {
		SendError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Get photos from database (verifies ownership)
	dbPhotos, err := app.db.GetPhotosByIDs(ctx, dbUser.ID, req.PhotoIds)
	if err != nil || len(dbPhotos) == 0 {
		SendError(w, "No valid photos found", http.StatusBadRequest)
		return
	}

	// Download photos from GCS for AI analysis
	var photoPaths []string
	var photoData [][]byte
	for _, photo := range dbPhotos {
		data, err := app.storage.DownloadToBuffer(ctx, photo.GCSPath)
		if err != nil {
			log.Printf("Failed to download photo %s: %v", photo.ID, err)
			continue
		}
		photoPaths = append(photoPaths, photo.GCSPath)
		photoData = append(photoData, data)
	}

	if len(photoData) == 0 {
		SendError(w, "Failed to load photos for analysis", http.StatusInternalServerError)
		return
	}

	// Get the photo IDs that were successfully loaded
	var validPhotoIds []string
	for _, photo := range dbPhotos {
		validPhotoIds = append(validPhotoIds, photo.ID)
	}

	// Use Gemini AI to analyze and cluster photos
	clusters, err := AnalyzeAndClusterPhotosWithData(validPhotoIds, photoData)
	if err != nil {
		log.Printf("Error clustering photos: %v", err)
		SendError(w, "Failed to analyze photos", http.StatusInternalServerError)
		return
	}

	// Generate background images and create drafts
	var pageDrafts []PageDraft
	for i, cluster := range clusters {
		// Save cluster to database
		dbCluster := &DBCluster{
			ID:          cluster.ID,
			UserID:      dbUser.ID,
			Title:       sql.NullString{String: cluster.Title, Valid: cluster.Title != ""},
			Description: sql.NullString{String: cluster.Description, Valid: cluster.Description != ""},
			Theme:       sql.NullString{String: cluster.Theme, Valid: cluster.Theme != ""},
			Date:        sql.NullString{String: cluster.Date, Valid: cluster.Date != ""},
		}
		if err := app.db.CreateCluster(ctx, dbCluster); err != nil {
			log.Printf("Failed to save cluster: %v", err)
		}
		if err := app.db.AddPhotosToCluster(ctx, cluster.ID, cluster.PhotoIds); err != nil {
			log.Printf("Failed to add photos to cluster: %v", err)
		}

		// Generate themed background image
		var backgroundURL string
		var bgGCSPath string
		backgroundData, err := GenerateBackgroundImageData(cluster.Theme, cluster.Title, cluster.Description)
		if err != nil {
			log.Printf("Failed to generate background for cluster %s: %v", cluster.ID, err)
		} else {
			bgPath, err := app.storage.UploadBackground(ctx, dbUser.ID, backgroundData, cluster.Theme)
			if err != nil {
				log.Printf("Failed to upload background: %v", err)
			} else {
				bgGCSPath = bgPath
				backgroundURL, _ = app.storage.GetSignedURL(ctx, bgPath)
				clusters[i].BackgroundPath = backgroundURL
			}
		}

		// Create draft
		draftID := uuid.New().String()
		dbDraft := &DBPageDraft{
			ID:                draftID,
			UserID:            dbUser.ID,
			ClusterID:         sql.NullString{String: cluster.ID, Valid: true},
			Title:             sql.NullString{String: cluster.Title, Valid: cluster.Title != ""},
			Description:       sql.NullString{String: cluster.Description, Valid: cluster.Description != ""},
			Theme:             sql.NullString{String: cluster.Theme, Valid: cluster.Theme != ""},
			BackgroundGCSPath: sql.NullString{String: bgGCSPath, Valid: bgGCSPath != ""},
			Status:            "draft",
		}

		if err := app.db.CreateDraft(ctx, dbDraft); err != nil {
			log.Printf("Failed to create draft: %v", err)
			continue
		}
		if err := app.db.AddPhotosToDraft(ctx, draftID, cluster.PhotoIds); err != nil {
			log.Printf("Failed to add photos to draft: %v", err)
		}

		draft := PageDraft{
			ID:             draftID,
			ClusterID:      cluster.ID,
			PhotoIds:       cluster.PhotoIds,
			Title:          cluster.Title,
			Description:    cluster.Description,
			Theme:          cluster.Theme,
			BackgroundPath: backgroundURL,
			Status:         "draft",
			CreatedAt:      time.Now().Format(time.RFC3339),
		}
		pageDrafts = append(pageDrafts, draft)
	}

	response := ClusterResponse{
		Clusters: clusters,
		Drafts:   pageDrafts,
	}

	SendJSON(w, response)
}
