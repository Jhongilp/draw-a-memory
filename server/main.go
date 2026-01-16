package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const (
	uploadDir     = "./uploads"
	maxFileSize   = 5 << 20  // 5 MB per file
	maxTotalSize  = 50 << 20 // 50 MB total (10 files * 5 MB)
	maxPhotoCount = 10
	serverPort    = ":8080"
)

// generateMissingThumbnails creates thumbnails for any existing photos that don't have them
func generateMissingThumbnails() {
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		log.Printf("Warning: could not read upload directory for thumbnail generation: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		ext := strings.ToLower(filepath.Ext(name))

		// Skip non-image files and existing thumbnails
		if !isValidImageType(name) || strings.Contains(name, "_thumb") {
			continue
		}

		// Check if thumbnail already exists
		baseName := strings.TrimSuffix(name, ext)
		thumbName := baseName + "_thumb.jpg"
		thumbPath := filepath.Join(uploadDir, thumbName)

		if _, err := os.Stat(thumbPath); err == nil {
			continue // Thumbnail already exists
		}

		// Generate thumbnail
		srcPath := filepath.Join(uploadDir, name)
		log.Printf("Generating missing thumbnail for: %s", name)
		if err := generateThumbnail(srcPath, thumbPath); err != nil {
			log.Printf("Warning: failed to generate thumbnail for %s: %v", name, err)
		}
	}
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Generate thumbnails for any existing photos that don't have them
	log.Println("Checking for missing thumbnails...")
	generateMissingThumbnails()
	log.Println("Thumbnail check complete")

	// Set up routes
	http.HandleFunc("/api/photos/upload", CorsMiddleware(HandleUpload))
	http.HandleFunc("/api/photos/cluster", CorsMiddleware(HandleClusterPhotos))
	http.HandleFunc("/api/photos", CorsMiddleware(HandleGetPhotos))
	http.HandleFunc("/api/drafts/", CorsMiddleware(HandleDrafts))
	http.HandleFunc("/uploads/", CorsMiddleware(HandleServePhoto))

	log.Printf("Server starting on port %s", serverPort)
	log.Printf("Upload directory: %s", uploadDir)

	if err := http.ListenAndServe(serverPort, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
