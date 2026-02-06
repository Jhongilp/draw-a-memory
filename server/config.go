package main

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port        string
	Environment string // "development" | "production"

	// Database (Cloud SQL)
	DatabaseURL string

	// Google Cloud Storage
	GCSBucket    string
	GCSProjectID string

	// Clerk Authentication
	ClerkSecretKey string
	ClerkJWKSURL   string

	// CORS
	AllowedOrigins []string

	// Gemini AI
	GeminiAPIKey string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	env := getEnv("ENVIRONMENT", "development")

	config := &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    env,
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		GCSBucket:      os.Getenv("GCS_BUCKET"),
		GCSProjectID:   os.Getenv("GCS_PROJECT_ID"),
		ClerkSecretKey: os.Getenv("CLERK_SECRET_KEY"),
		ClerkJWKSURL:   getEnv("CLERK_JWKS_URL", ""),
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
	}

	// Set allowed origins based on environment
	if env == "production" {
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			return nil, fmt.Errorf("FRONTEND_URL required in production")
		}
		// Support comma-separated origins for multiple domains
		config.AllowedOrigins = strings.Split(frontendURL, ",")
		for i, origin := range config.AllowedOrigins {
			config.AllowedOrigins[i] = strings.TrimSpace(origin)
		}
	} else {
		config.AllowedOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	}

	// Validate required config in production
	if env == "production" {
		if config.DatabaseURL == "" {
			return nil, fmt.Errorf("DATABASE_URL is required")
		}
		if config.GCSBucket == "" {
			return nil, fmt.Errorf("GCS_BUCKET is required")
		}
		if config.ClerkSecretKey == "" {
			return nil, fmt.Errorf("CLERK_SECRET_KEY is required")
		}
	}

	return config, nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
