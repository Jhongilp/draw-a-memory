package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

// AnalyzeAndClusterPhotos uses Gemini AI to analyze photos and create clusters
func AnalyzeAndClusterPhotos(photoIds []string, photoPaths []string) ([]PhotoCluster, error) {
	ctx := context.Background()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("No GEMINI_API_KEY set, using mock clusters")
		return CreateMockClusters(photoIds), nil
	}

	// Create Gemini client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Printf("Failed to create Gemini client: %v", err)
		return CreateMockClusters(photoIds), nil
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
		return CreateMockClusters(photoIds), nil
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Println("No response from Gemini")
		return CreateMockClusters(photoIds), nil
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			responseText += part.Text
		}
	}

	if responseText == "" {
		log.Println("Empty response from Gemini")
		return CreateMockClusters(photoIds), nil
	}

	log.Printf("Gemini response: %s", responseText)

	// Find JSON in response (it might be wrapped in markdown code blocks)
	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		log.Printf("No JSON found in Gemini response")
		return CreateMockClusters(photoIds), nil
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
		return CreateMockClusters(photoIds), nil
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
		return CreateMockClusters(photoIds), nil
	}

	return clusters, nil
}

// CreateMockClusters creates sample clusters when AI is not available
func CreateMockClusters(photoIds []string) []PhotoCluster {
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
