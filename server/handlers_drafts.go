package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// In-memory storage for drafts (in production, use a database)
var drafts = make(map[string]PageDraft)

// HandleDrafts handles CRUD operations for page drafts
func HandleDrafts(w http.ResponseWriter, r *http.Request) {
	// Extract draft ID from path if present
	path := strings.TrimPrefix(r.URL.Path, "/api/drafts/")
	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodGet:
		handleGetDrafts(w, r, parts)
	case http.MethodPut:
		handleUpdateDraft(w, r, parts)
	case http.MethodDelete:
		handleDeleteDraft(w, r, parts)
	default:
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetDrafts(w http.ResponseWriter, r *http.Request, parts []string) {
	// Get single draft by ID
	if len(parts) == 1 && parts[0] != "" {
		draftID := parts[0]
		if draft, ok := drafts[draftID]; ok {
			SendJSON(w, draft)
			return
		}
		SendError(w, "Draft not found", http.StatusNotFound)
		return
	}

	// Return all drafts
	var allDrafts []PageDraft
	for _, draft := range drafts {
		allDrafts = append(allDrafts, draft)
	}
	SendJSON(w, allDrafts)
}

func handleUpdateDraft(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) < 1 || parts[0] == "" {
		SendError(w, "Draft ID required", http.StatusBadRequest)
		return
	}

	draftID := parts[0]

	// Check if this is an approve action
	if len(parts) == 2 && parts[1] == "approve" {
		if draft, ok := drafts[draftID]; ok {
			draft.Status = "approved"
			drafts[draftID] = draft
			SendJSON(w, draft)
			return
		}
		SendError(w, "Draft not found", http.StatusNotFound)
		return
	}

	// Regular update
	var updatedDraft PageDraft
	if err := json.NewDecoder(r.Body).Decode(&updatedDraft); err != nil {
		SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if _, ok := drafts[draftID]; ok {
		updatedDraft.ID = draftID
		drafts[draftID] = updatedDraft
		SendJSON(w, updatedDraft)
		return
	}
	SendError(w, "Draft not found", http.StatusNotFound)
}

func handleDeleteDraft(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) < 1 || parts[0] == "" {
		SendError(w, "Draft ID required", http.StatusBadRequest)
		return
	}

	draftID := parts[0]
	if _, ok := drafts[draftID]; ok {
		delete(drafts, draftID)
		SendJSON(w, map[string]bool{"success": true})
		return
	}
	SendError(w, "Draft not found", http.StatusNotFound)
}
