package main

import (
	"context"
	"encoding/json"
	"fmt"
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

// themeToPromptStyle maps themes to artistic style descriptions for background generation
var themeToPromptStyle = map[string]string{
	"adventure":   "adventurous outdoor scenery with mountains, forests, soft watercolor style",
	"cozy":        "warm cozy interior, soft blankets, warm lighting, gentle pastel watercolor",
	"celebration": "festive confetti, balloons, sparkles, joyful pastel watercolor style",
	"nature":      "gentle nature scene, flowers, leaves, butterflies, soft botanical watercolor",
	"family":      "warm family home atmosphere, soft hearts, gentle embrace motifs, watercolor",
	"milestone":   "celebratory stars, achievement ribbons, gentle golden accents, watercolor",
	"playful":     "fun toys, colorful blocks, playful patterns, cheerful watercolor style",
	"love":        "soft hearts, gentle pink and red tones, romantic watercolor florals",
	"growth":      "growing plants, seedlings, gentle green sprouts, nature watercolor",
	"serene":      "calm clouds, peaceful sky, soft blue tones, dreamy watercolor style",
}

// GenerateBackgroundImage generates a themed background image using Gemini 2.0 Flash
func GenerateBackgroundImage(theme, title, description string) (string, error) {
	ctx := context.Background()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("No GEMINI_API_KEY set, skipping background generation")
		return "", nil
	}

	// Create Gemini client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Printf("Failed to create Gemini client for image generation: %v", err)
		return "", err
	}

	// Get style based on theme
	style, ok := themeToPromptStyle[theme]
	if !ok {
		style = themeToPromptStyle["love"] // default
	}

	// Create a prompt for a subtle, artistic background
	prompt := fmt.Sprintf(`Generate an image: A beautiful, soft, and subtle background for a baby memory book page. 
Theme: %s
Page title: %s
Style: %s

Requirements:
- Very soft, muted pastel colors
- Dreamy, ethereal watercolor or soft gradient style
- Abstract or semi-abstract design
- NO text, NO words, NO letters anywhere in the image
- Should work as a background (not too busy)
- Light enough that text and photos can be placed on top
- Gentle, soothing, and appropriate for a baby memory book
- Landscape orientation, suitable for a book page`, theme, title, style)

	// Configure for image generation with Gemini 2.0 Flash
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"image", "text"},
	}

	contents := []*genai.Content{
		genai.NewContentFromText(prompt, "user"),
	}

	// Generate using gemini-2.0-flash-exp which supports native image generation
	resp, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash-exp", contents, config)
	if err != nil {
		log.Printf("Gemini image generation error: %v", err)
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Println("No response from Gemini for image generation")
		return "", fmt.Errorf("no response from Gemini")
	}

	// Find the image part in the response
	var imageData []byte
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "image/") {
			imageData = part.InlineData.Data
			break
		}
	}

	if len(imageData) == 0 {
		log.Println("No image data in Gemini response")
		return "", fmt.Errorf("no image generated")
	}

	// Create backgrounds directory if it doesn't exist
	backgroundDir := "./uploads/backgrounds"
	if err := os.MkdirAll(backgroundDir, 0755); err != nil {
		log.Printf("Failed to create backgrounds directory: %v", err)
		return "", err
	}

	// Save the image
	filename := fmt.Sprintf("bg_%s_%s.png", theme, uuid.New().String()[:8])
	filePath := filepath.Join(backgroundDir, filename)

	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		log.Printf("Failed to save background image: %v", err)
		return "", err
	}

	// Return the URL path
	urlPath := fmt.Sprintf("/uploads/backgrounds/%s", filename)
	log.Printf("Generated background image: %s", urlPath)

	return urlPath, nil
}
