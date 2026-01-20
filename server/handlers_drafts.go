package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// HandleDrafts handles CRUD operations for page drafts
func (app *App) HandleDrafts(w http.ResponseWriter, r *http.Request) {
	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		SendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract draft ID from path if present
	path := strings.TrimPrefix(r.URL.Path, "/api/drafts/")
	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodGet:
		app.handleGetDrafts(w, r, parts, authUser)
	case http.MethodPut:
		app.handleUpdateDraft(w, r, parts, authUser)
	case http.MethodDelete:
		app.handleDeleteDraft(w, r, parts, authUser)
	default:
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleGetDrafts(w http.ResponseWriter, r *http.Request, parts []string, authUser *AuthUser) {
	ctx := r.Context()

	// Get database user
	dbUser, err := app.db.GetOrCreateUser(ctx, authUser.ClerkID, authUser.Email, authUser.Name)
	if err != nil {
		SendError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Get single draft by ID
	if len(parts) == 1 && parts[0] != "" {
		draftID := parts[0]
		draft, err := app.db.GetDraftByID(ctx, draftID)
		if err != nil {
			SendError(w, "Draft not found", http.StatusNotFound)
			return
		}

		// Verify ownership
		if draft.UserID != dbUser.ID {
			SendError(w, "Unauthorized", http.StatusForbidden)
			return
		}

		// Get photo IDs
		photoIDs, _ := app.db.GetDraftPhotos(ctx, draftID)

		// Generate signed URL for background if exists
		var backgroundURL string
		if draft.BackgroundGCSPath.Valid {
			backgroundURL, _ = app.storage.GetSignedURL(ctx, draft.BackgroundGCSPath.String)
		}

		apiDraft := draft.ToAPIPageDraft(photoIDs, backgroundURL)
		SendJSON(w, apiDraft)
		return
	}

	// Return all drafts for user
	dbDrafts, err := app.db.GetDraftsByUser(ctx, dbUser.ID)
	if err != nil {
		log.Printf("Failed to get drafts: %v", err)
		SendError(w, "Failed to retrieve drafts", http.StatusInternalServerError)
		return
	}

	log.Printf("[DEBUG] Found %d drafts for user %s", len(dbDrafts), dbUser.ID)

	var allDrafts []PageDraft
	for _, draft := range dbDrafts {
		photoIDs, _ := app.db.GetDraftPhotos(ctx, draft.ID)

		var backgroundURL string
		if draft.BackgroundGCSPath.Valid {
			backgroundURL, _ = app.storage.GetSignedURL(ctx, draft.BackgroundGCSPath.String)
		}

		allDrafts = append(allDrafts, draft.ToAPIPageDraft(photoIDs, backgroundURL))
	}

	if allDrafts == nil {
		allDrafts = []PageDraft{}
	}

	SendJSON(w, allDrafts)
}

func (app *App) handleUpdateDraft(w http.ResponseWriter, r *http.Request, parts []string, authUser *AuthUser) {
	if len(parts) < 1 || parts[0] == "" {
		SendError(w, "Draft ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	draftID := parts[0]

	// Get database user
	dbUser, err := app.db.GetOrCreateUser(ctx, authUser.ClerkID, authUser.Email, authUser.Name)
	if err != nil {
		SendError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Get existing draft
	existingDraft, err := app.db.GetDraftByID(ctx, draftID)
	if err != nil {
		SendError(w, "Draft not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if existingDraft.UserID != dbUser.ID {
		SendError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Check if this is an approve action
	log.Printf("[DEBUG] handleUpdateDraft: parts=%v, len=%d", parts, len(parts))
	if len(parts) == 2 && parts[1] == "approve" {
		log.Printf("[DEBUG] Approving draft ID=%s", draftID)
		existingDraft.Status = "approved"
		if err := app.db.UpdateDraft(ctx, existingDraft); err != nil {
			log.Printf("[ERROR] Failed to approve draft %s: %v", draftID, err)
			SendError(w, "Failed to approve draft", http.StatusInternalServerError)
			return
		}
		log.Printf("[DEBUG] Successfully approved draft ID=%s", draftID)

		photoIDs, _ := app.db.GetDraftPhotos(ctx, draftID)
		var backgroundURL string
		if existingDraft.BackgroundGCSPath.Valid {
			backgroundURL, _ = app.storage.GetSignedURL(ctx, existingDraft.BackgroundGCSPath.String)
		}

		SendJSON(w, existingDraft.ToAPIPageDraft(photoIDs, backgroundURL))
		return
	}

	// Regular update
	var updateReq PageDraft
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	existingDraft.Title = sql.NullString{String: updateReq.Title, Valid: updateReq.Title != ""}
	existingDraft.Description = sql.NullString{String: updateReq.Description, Valid: updateReq.Description != ""}
	existingDraft.Theme = sql.NullString{String: updateReq.Theme, Valid: updateReq.Theme != ""}
	if updateReq.Status != "" {
		existingDraft.Status = updateReq.Status
	}

	if err := app.db.UpdateDraft(ctx, existingDraft); err != nil {
		SendError(w, "Failed to update draft", http.StatusInternalServerError)
		return
	}

	photoIDs, _ := app.db.GetDraftPhotos(ctx, draftID)
	var backgroundURL string
	if existingDraft.BackgroundGCSPath.Valid {
		backgroundURL, _ = app.storage.GetSignedURL(ctx, existingDraft.BackgroundGCSPath.String)
	}

	SendJSON(w, existingDraft.ToAPIPageDraft(photoIDs, backgroundURL))
}

func (app *App) handleDeleteDraft(w http.ResponseWriter, r *http.Request, parts []string, authUser *AuthUser) {
	if len(parts) < 1 || parts[0] == "" {
		SendError(w, "Draft ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	draftID := parts[0]

	// Get database user
	dbUser, err := app.db.GetOrCreateUser(ctx, authUser.ClerkID, authUser.Email, authUser.Name)
	if err != nil {
		SendError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	if err := app.db.DeleteDraft(ctx, dbUser.ID, draftID); err != nil {
		SendError(w, "Draft not found", http.StatusNotFound)
		return
	}

	SendJSON(w, map[string]bool{"success": true})
}
