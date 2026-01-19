package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Initialize database
	var db *Database
	if config.DatabaseURL != "" {
		db, err = NewDatabase(config.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()
		log.Println("Connected to Cloud SQL database")
	} else {
		log.Println("Warning: DATABASE_URL not set, running without database")
	}

	// Initialize storage
	var storage *Storage
	if config.GCSBucket != "" && config.GCSProjectID != "" {
		storage, err = NewStorage(ctx, config.GCSProjectID, config.GCSBucket)
		if err != nil {
			log.Fatalf("Failed to initialize storage: %v", err)
		}
		defer storage.Close()
		log.Printf("Connected to GCS bucket: %s", config.GCSBucket)
	} else {
		log.Println("Warning: GCS_BUCKET or GCS_PROJECT_ID not set, running without cloud storage")
	}

	// Initialize auth middleware
	auth := NewAuthMiddleware(config)

	// Create app with dependencies
	app := &App{
		config:  config,
		db:      db,
		storage: storage,
		auth:    auth,
	}

	// Create CORS middleware
	cors := CorsMiddleware(config)

	// Set up routes
	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Photo endpoints (auth required)
	mux.HandleFunc("/api/photos/upload", cors(auth.Middleware(app.HandleUpload)))
	mux.HandleFunc("/api/photos/cluster", cors(auth.Middleware(app.HandleClusterPhotos)))
	mux.HandleFunc("/api/photos", cors(auth.Middleware(app.HandleGetPhotos)))
	mux.HandleFunc("/api/photos/", cors(auth.Middleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.HandleGetPhotoURL(w, r)
		case http.MethodDelete:
			app.HandleDeletePhoto(w, r)
		default:
			SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Draft endpoints (auth required)
	mux.HandleFunc("/api/drafts/", cors(auth.Middleware(app.HandleDrafts)))

	// Create server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", config.Port)
		log.Printf("Environment: %s", config.Environment)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
