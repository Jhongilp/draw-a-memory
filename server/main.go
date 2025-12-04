package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const (
	uploadDir     = "./uploads"
	maxFileSize   = 5 << 20  // 5 MB per file
	maxTotalSize  = 50 << 20 // 50 MB total (10 files * 5 MB)
	maxPhotoCount = 10
	serverPort    = ":8080"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

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
