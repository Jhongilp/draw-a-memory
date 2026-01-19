package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

// AnalyzeAndClusterPhotosWithData analyzes photos from byte data (for GCS-stored photos)
func AnalyzeAndClusterPhotosWithData(photoIds []string, photoData [][]byte) ([]PhotoCluster, error) {
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

	// Add images from byte data
	for _, data := range photoData {
		// Detect mime type from magic bytes
		mimeType := detectMimeType(data)
		imagePart := genai.NewPartFromBytes(data, mimeType)
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

	// Find JSON in response
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

// GenerateBackgroundImageData generates a themed background and returns the image bytes
func GenerateBackgroundImageData(theme, title, description string) ([]byte, error) {
	ctx := context.Background()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("No GEMINI_API_KEY set, skipping background generation")
		return nil, fmt.Errorf("no API key")
	}

	// Create Gemini client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Printf("Failed to create Gemini client for image generation: %v", err)
		return nil, err
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

	// Configure for image generation
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
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Println("No response from Gemini for image generation")
		return nil, fmt.Errorf("no response from Gemini")
	}

	// Find the image part in the response
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "image/") {
			return part.InlineData.Data, nil
		}
	}

	log.Println("No image data in Gemini response")
	return nil, fmt.Errorf("no image generated")
}

// detectMimeType detects the MIME type from image magic bytes
func detectMimeType(data []byte) string {
	if len(data) < 4 {
		return "image/jpeg"
	}

	// Check magic bytes
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		if data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
			return "image/webp"
		}
	}

	return "image/jpeg" // default
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
