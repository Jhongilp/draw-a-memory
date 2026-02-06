package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// HandleSettings handles user settings CRUD operations
func (app *App) HandleSettings(w http.ResponseWriter, r *http.Request) {
	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		SendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	dbUser, err := app.db.GetOrCreateUser(ctx, authUser.ClerkID, authUser.Email, authUser.Name)
	if err != nil {
		SendError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.handleGetSettings(w, dbUser)
	case http.MethodPut:
		app.handleUpdateSettings(w, r, dbUser)
	default:
		SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleGetSettings(w http.ResponseWriter, user *DBUser) {
	settings := UserSettings{}

	if user.ChildName.Valid {
		settings.ChildName = user.ChildName.String
	}

	if user.ChildBirthday.Valid {
		dateStr := user.ChildBirthday.Time.Format("2006-01-02")
		settings.ChildBirthday = &dateStr
	}

	SendJSON(w, settings)
}

func (app *App) handleUpdateSettings(w http.ResponseWriter, r *http.Request, user *DBUser) {
	var req UserSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var birthday *time.Time
	if req.ChildBirthday != nil && *req.ChildBirthday != "" {
		parsed, err := time.Parse("2006-01-02", *req.ChildBirthday)
		if err != nil {
			log.Printf("[ERROR] Failed to parse birthday: %v", err)
			SendError(w, "Invalid birthday format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		birthday = &parsed
	}

	if err := app.db.UpdateUserSettings(r.Context(), user.ID, req.ChildName, birthday); err != nil {
		log.Printf("[ERROR] Failed to update settings: %v", err)
		SendError(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	// Return updated settings
	settings := UserSettings{
		ChildName:     req.ChildName,
		ChildBirthday: req.ChildBirthday,
	}

	SendJSON(w, settings)
}
