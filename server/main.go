package main

import (
	"context"
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
	"github.com/joho/godotenv"
	"google.golang.org/genai"
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

type PhotoCluster struct {
	ID          string   `json:"id"`
	PhotoIds    []string `json:"photoIds"`
	Theme       string   `json:"theme"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Date        string   `json:"date"`
}

type PageDraft struct {
	ID          string   `json:"id"`
	ClusterID   string   `json:"clusterId"`
	PhotoIds    []string `json:"photoIds"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Theme       string   `json:"theme"`
	Status      string   `json:"status"` // "draft" | "approved" | "rejected"
	CreatedAt   string   `json:"createdAt"`
}

type ClusterRequest struct {
	PhotoIds []string `json:"photoIds"`
}

type ClusterResponse struct {
	Clusters []PhotoCluster `json:"clusters"`
	Drafts   []PageDraft    `json:"drafts"`
}

// In-memory storage for drafts (in production, use a database)
var drafts = make(map[string]PageDraft)

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
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Set up routes
	http.HandleFunc("/api/photos/upload", corsMiddleware(handleUpload))
	http.HandleFunc("/api/photos", corsMiddleware(handleGetPhotos))
	http.HandleFunc("/api/photos/cluster", corsMiddleware(handleClusterPhotos))
	http.HandleFunc("/api/drafts/", corsMiddleware(handleDrafts))
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

// handleClusterPhotos analyzes photos using Gemini AI and groups them into clusters
func handleClusterPhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.PhotoIds) == 0 {
		sendError(w, "No photo IDs provided", http.StatusBadRequest)
		return
	}

	// Get photo file paths
	var photoPaths []string
	for _, photoID := range req.PhotoIds {
		// Find the file with this ID
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
		sendError(w, "No valid photos found", http.StatusBadRequest)
		return
	}

	// Use Gemini AI to analyze and cluster photos
	clusters, err := analyzeAndClusterPhotos(req.PhotoIds, photoPaths)
	if err != nil {
		log.Printf("Error clustering photos: %v", err)
		sendError(w, "Failed to analyze photos", http.StatusInternalServerError)
		return
	}

	// Create drafts from clusters
	var pageDrafts []PageDraft
	for _, cluster := range clusters {
		draft := PageDraft{
			ID:          uuid.New().String(),
			ClusterID:   cluster.ID,
			PhotoIds:    cluster.PhotoIds,
			Title:       cluster.Title,
			Description: cluster.Description,
			Theme:       cluster.Theme,
			Status:      "draft",
			CreatedAt:   time.Now().Format(time.RFC3339),
		}
		drafts[draft.ID] = draft
		pageDrafts = append(pageDrafts, draft)
	}

	response := ClusterResponse{
		Clusters: clusters,
		Drafts:   pageDrafts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDrafts handles CRUD operations for page drafts
func handleDrafts(w http.ResponseWriter, r *http.Request) {
	// Extract draft ID from path if present
	path := strings.TrimPrefix(r.URL.Path, "/api/drafts/")
	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodGet:
		// Get all drafts or single draft
		if len(parts) == 1 && parts[0] != "" {
			draftID := parts[0]
			if draft, ok := drafts[draftID]; ok {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(draft)
				return
			}
			sendError(w, "Draft not found", http.StatusNotFound)
			return
		}
		// Return all drafts
		var allDrafts []PageDraft
		for _, draft := range drafts {
			allDrafts = append(allDrafts, draft)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allDrafts)

	case http.MethodPut:
		// Update draft (including approval)
		if len(parts) >= 1 && parts[0] != "" {
			draftID := parts[0]

			// Check if this is an approve action
			if len(parts) == 2 && parts[1] == "approve" {
				if draft, ok := drafts[draftID]; ok {
					draft.Status = "approved"
					drafts[draftID] = draft
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(draft)
					return
				}
				sendError(w, "Draft not found", http.StatusNotFound)
				return
			}

			// Regular update
			var updatedDraft PageDraft
			if err := json.NewDecoder(r.Body).Decode(&updatedDraft); err != nil {
				sendError(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if _, ok := drafts[draftID]; ok {
				updatedDraft.ID = draftID
				drafts[draftID] = updatedDraft
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(updatedDraft)
				return
			}
			sendError(w, "Draft not found", http.StatusNotFound)
		}

	case http.MethodDelete:
		if len(parts) >= 1 && parts[0] != "" {
			draftID := parts[0]
			if _, ok := drafts[draftID]; ok {
				delete(drafts, draftID)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]bool{"success": true})
				return
			}
			sendError(w, "Draft not found", http.StatusNotFound)
		}

	default:
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// analyzeAndClusterPhotos uses Gemini AI to analyze photos and create clusters
func analyzeAndClusterPhotos(photoIds []string, photoPaths []string) ([]PhotoCluster, error) {
	ctx := context.Background()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("No GEMINI_API_KEY set, using mock clusters")
		return createMockClusters(photoIds), nil
	}

	// Create Gemini client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Printf("Failed to create Gemini client: %v", err)
		return createMockClusters(photoIds), nil
	}

	// Build the prompt parts
	var parts []*genai.Part

	// Add instruction text
	promptText := `Analyze these baby photos and group them into meaningful clusters based on activity, setting, or moment type. 
For each cluster, provide:
- A short, sweet title (e.g., "First Steps", "Bath Time Fun", "Sleepy Moments")
- A heartfelt description that a parent would love to read (2-3 sentences)
- A theme from: "milestone", "playful", "cozy", "adventure", "love", "growth"

Respond in this exact JSON format:
{
  "clusters": [
    {
      "photoIndexes": [0, 2],
      "title": "Title Here",
      "description": "Description here",
      "theme": "milestone"
    }
  ]
}

Make sure every photo is included in exactly one cluster.`

	textPart := genai.NewPartFromText(promptText)
	parts = append(parts, textPart)

	// Add images
	for _, photoPath := range photoPaths {
		imageData, err := os.ReadFile(photoPath)
		if err != nil {
			log.Printf("Error reading photo %s: %v", photoPath, err)
			continue
		}

		mimeType := "image/jpeg"
		ext := strings.ToLower(filepath.Ext(photoPath))
		switch ext {
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		}

		imagePart := genai.NewPartFromBytes(imageData, mimeType)
		parts = append(parts, imagePart)
	}

	// Create the content
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, "user"),
	}

	// Configure generation
	config := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(float32(0.7)),
		TopP:            genai.Ptr(float32(0.95)),
		MaxOutputTokens: 2048,
	}

	// Generate content using gemini-2.5-flash
	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
	if err != nil {
		log.Printf("Gemini API error: %v", err)
		return createMockClusters(photoIds), nil
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Println("No response from Gemini")
		return createMockClusters(photoIds), nil
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			responseText += part.Text
		}
	}

	if responseText == "" {
		log.Println("Empty response from Gemini")
		return createMockClusters(photoIds), nil
	}

	log.Printf("Gemini response: %s", responseText)

	// Find JSON in response (it might be wrapped in markdown code blocks)
	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		log.Printf("No JSON found in Gemini response")
		return createMockClusters(photoIds), nil
	}
	jsonStr := responseText[jsonStart : jsonEnd+1]

	var clusterResp struct {
		Clusters []struct {
			PhotoIndexes []int  `json:"photoIndexes"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			Theme        string `json:"theme"`
		} `json:"clusters"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &clusterResp); err != nil {
		log.Printf("Failed to parse cluster JSON: %v", err)
		return createMockClusters(photoIds), nil
	}

	// Convert to PhotoCluster with actual photo IDs
	var clusters []PhotoCluster
	for _, c := range clusterResp.Clusters {
		var clusterPhotoIds []string
		for _, idx := range c.PhotoIndexes {
			if idx >= 0 && idx < len(photoIds) {
				clusterPhotoIds = append(clusterPhotoIds, photoIds[idx])
			}
		}
		if len(clusterPhotoIds) == 0 {
			continue
		}

		cluster := PhotoCluster{
			ID:          uuid.New().String(),
			PhotoIds:    clusterPhotoIds,
			Theme:       c.Theme,
			Title:       c.Title,
			Description: c.Description,
			Date:        time.Now().Format("January 2006"),
		}
		clusters = append(clusters, cluster)
	}

	if len(clusters) == 0 {
		return createMockClusters(photoIds), nil
	}

	return clusters, nil
}

// createMockClusters creates sample clusters when AI is not available
func createMockClusters(photoIds []string) []PhotoCluster {
	if len(photoIds) == 0 {
		return nil
	}

	// Create a single cluster with all photos
	return []PhotoCluster{
		{
			ID:          uuid.New().String(),
			PhotoIds:    photoIds,
			Theme:       "love",
			Title:       "Precious Moments",
			Description: "A beautiful collection of memories capturing the joy and wonder of these special moments. Each photo tells a story of love and growth.",
			Date:        time.Now().Format("January 2006"),
		},
	}
}
