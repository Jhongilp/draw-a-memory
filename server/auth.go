package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	userContextKey contextKey = "user"
)

// AuthUser represents the authenticated user from the JWT
type AuthUser struct {
	ClerkID string
	Email   string
	Name    string
}

// AuthMiddleware validates Clerk JWTs and extracts user info
type AuthMiddleware struct {
	jwksURL   string
	jwksCache jwk.Set
	cacheMu   sync.RWMutex
	cacheTime time.Time
	cacheTTL  time.Duration
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(config *Config) *AuthMiddleware {
	// Use Clerk's JWKS URL - format: https://<your-clerk-instance>.clerk.accounts.dev/.well-known/jwks.json
	jwksURL := config.ClerkJWKSURL
	if jwksURL == "" {
		// Default Clerk JWKS endpoint pattern
		jwksURL = "https://clerk.your-domain.com/.well-known/jwks.json"
	}

	return &AuthMiddleware{
		jwksURL:  jwksURL,
		cacheTTL: 1 * time.Hour,
	}
}

// getJWKS fetches and caches the JWKS
func (am *AuthMiddleware) getJWKS(ctx context.Context) (jwk.Set, error) {
	am.cacheMu.RLock()
	if am.jwksCache != nil && time.Since(am.cacheTime) < am.cacheTTL {
		cache := am.jwksCache
		am.cacheMu.RUnlock()
		return cache, nil
	}
	am.cacheMu.RUnlock()

	am.cacheMu.Lock()
	defer am.cacheMu.Unlock()

	// Double-check after acquiring write lock
	if am.jwksCache != nil && time.Since(am.cacheTime) < am.cacheTTL {
		return am.jwksCache, nil
	}

	set, err := jwk.Fetch(ctx, am.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	am.jwksCache = set
	am.cacheTime = time.Now()
	return set, nil
}

// ValidateToken validates a Clerk JWT and returns the claims
func (am *AuthMiddleware) ValidateToken(ctx context.Context, tokenString string) (*AuthUser, error) {
	// Fetch JWKS
	keySet, err := am.getJWKS(ctx)
	if err != nil {
		return nil, err
	}

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		// Find the key in the JWKS
		key, found := keySet.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("key %s not found in JWKS", kid)
		}

		var rawKey interface{}
		if err := key.Raw(&rawKey); err != nil {
			return nil, fmt.Errorf("failed to get raw key: %w", err)
		}

		return rawKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	// Extract user info from Clerk JWT claims
	user := &AuthUser{
		ClerkID: claims["sub"].(string),
	}

	// Clerk puts email in different places depending on configuration
	if email, ok := claims["email"].(string); ok {
		user.Email = email
	}

	if name, ok := claims["name"].(string); ok {
		user.Name = name
	}

	return user, nil
}

// Middleware wraps an http.HandlerFunc with authentication
func (am *AuthMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			SendError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			SendError(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate token
		user, err := am.ValidateToken(r.Context(), tokenString)
		if err != nil {
			log.Printf("Auth error: %v", err)
			SendError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(ctx context.Context) *AuthUser {
	user, ok := ctx.Value(userContextKey).(*AuthUser)
	if !ok {
		return nil
	}
	return user
}

// OptionalAuthMiddleware allows requests without auth but enriches context if present
func (am *AuthMiddleware) OptionalAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				user, err := am.ValidateToken(r.Context(), parts[1])
				if err == nil {
					ctx := context.WithValue(r.Context(), userContextKey, user)
					r = r.WithContext(ctx)
				}
			}
		}
		next(w, r)
	}
}

// CorsMiddleware adds CORS headers to responses
func CorsMiddleware(config *Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}
}

// SendError sends a JSON error response
func SendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error:   message,
	})
}

// SendJSON sends a JSON response
func SendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
