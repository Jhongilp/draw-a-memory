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

const (
	uploadDir     = "./uploads"
	maxFileSize   = 5 << 20  // 5 MB per file
	maxTotalSize  = 50 << 20 // 50 MB total (10 files * 5 MB)
	maxPhotoCount = 10
	serverPort    = ":8080"
)

type Photo struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploadedAt"`
}

type UploadResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Photos  []Photo `json:"photos,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Set up routes
	http.HandleFunc("/api/photos/upload", corsMiddleware(handleUpload))
	http.HandleFunc("/api/photos", corsMiddleware(handleGetPhotos))
	http.HandleFunc("/uploads/", corsMiddleware(handleServePhoto))

	log.Printf("Server starting on port %s", serverPort)
	log.Printf("Upload directory: %s", uploadDir)

	if err := http.ListenAndServe(serverPort, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with max size - use memory limit, not body limit
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("ParseMultipartForm error: %v", err)
		sendError(w, "Failed to parse upload. Maximum total size is 50MB", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	files := r.MultipartForm.File["photos"]
	if len(files) == 0 {
		sendError(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	if len(files) > maxPhotoCount {
		sendError(w, fmt.Sprintf("Too many files. Maximum is %d photos per upload", maxPhotoCount), http.StatusBadRequest)
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
		log.Printf("saving photo %s, %s", photoID, fileHeader.Filename)

		uploadedPhotos = append(uploadedPhotos, photo)

		log.Printf("Uploaded: %s -> %s (%d bytes)", fileHeader.Filename, newFilename, size)
	}

	if len(uploadedPhotos) == 0 {
		sendError(w, "No valid images were uploaded", http.StatusBadRequest)
		return
	}

	response := UploadResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully uploaded %d photo(s)", len(uploadedPhotos)),
		Photos:  uploadedPhotos,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		sendError(w, "Failed to read photos directory", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photos)
}

func handleServePhoto(w http.ResponseWriter, r *http.Request) {
	// Serve static files from uploads directory
	filename := strings.TrimPrefix(r.URL.Path, "/uploads/")
	filePath := filepath.Join(uploadDir, filename)

	// Security: prevent directory traversal
	if strings.Contains(filename, "..") {
		sendError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, filePath)
}

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

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error:   message,
	})
}
