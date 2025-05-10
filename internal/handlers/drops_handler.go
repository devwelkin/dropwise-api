package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/twomotive/dropwise/internal/config"
	db "github.com/twomotive/dropwise/internal/database/sqlc"
	"github.com/twomotive/dropwise/internal/server/httputils"
)

// DropsHandler handles HTTP requests for drops.
type DropsHandler struct {
	APIConfig *config.APIConfig
}

// NewDropsHandler creates a new DropsHandler.
func NewDropsHandler(apiCfg *config.APIConfig) *DropsHandler {
	return &DropsHandler{APIConfig: apiCfg}
}

// CreateDropRequest defines the expected request body for creating a drop.
type CreateDropRequest struct {
	UserID    string `json:"user_id"` // Will be made more robust later
	Topic     string `json:"topic"`
	URL       string `json:"url"`
	UserNotes string `json:"user_notes,omitempty"`
	Priority  *int32 `json:"priority,omitempty"`
}

type UpdateDropRequest struct {
	Topic     *string `json:"topic,omitempty"`
	URL       *string `json:"url,omitempty"`
	UserNotes *string `json:"user_notes,omitempty"`
	Priority  *int32  `json:"priority,omitempty"`
	Status    *string `json:"status,omitempty"` // e.g., "new", "sent", "archived"
}

// CreateDropHandler handles the creation of a new drop.
// POST /api/v1/drops
func (h *DropsHandler) CreateDropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	var req CreateDropRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Basic Validation
	if strings.TrimSpace(req.Topic) == "" {
		httputils.RespondWithError(w, http.StatusBadRequest, "Topic cannot be empty")
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		httputils.RespondWithError(w, http.StatusBadRequest, "URL cannot be empty")
		return
	}
	// A placeholder for UserID until proper auth is implemented.
	// For now, if not provided in request, use a default.
	// In a real app, this would come from auth context.
	userID := req.UserID
	if strings.TrimSpace(userID) == "" {
		userID = "default-user" // Placeholder
		log.Printf("UserID not provided, using default: %s", userID)
	}

	params := db.CreateDropParams{
		UserID: sql.NullString{String: userID, Valid: true},
		Topic:  req.Topic,
		Url:    req.URL,
	}

	if req.UserNotes != "" {
		params.UserNotes = sql.NullString{String: req.UserNotes, Valid: true}
	} else {
		params.UserNotes = sql.NullString{Valid: false}
	}

	if req.Priority != nil {
		params.Priority = sql.NullInt32{Int32: *req.Priority, Valid: true}
	} else {
		params.Priority = sql.NullInt32{Valid: false}
	}

	log.Printf("Attempting to create drop with UserID: %s, Topic: %s", params.UserID.String, params.Topic)

	createdDrop, err := h.APIConfig.DB.CreateDrop(r.Context(), params)
	if err != nil {
		log.Printf("Error creating drop in database: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to create drop: "+err.Error())
		return
	}

	log.Printf("Successfully created drop with ID: %s", createdDrop.ID.String())
	httputils.RespondWithJSON(w, http.StatusCreated, createdDrop)
}

func (h *DropsHandler) GetDropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only GET method is allowed")
		return
	}

	dropIDStr := r.PathValue("id")
	if dropIDStr == "" {
		httputils.RespondWithError(w, http.StatusBadRequest, "Drop ID is required in the path")
		return
	}

	dropID, err := uuid.Parse(dropIDStr)
	if err != nil {
		httputils.RespondWithError(w, http.StatusBadRequest, "Invalid Drop ID format: "+err.Error())
		return
	}

	log.Printf("Attempting to fetch drop with ID: %s", dropID.String())

	drop, err := h.APIConfig.DB.GetDrop(r.Context(), dropID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Drop with ID %s not found", dropID.String())
			httputils.RespondWithError(w, http.StatusNotFound, "Drop not found")
		} else {
			log.Printf("Error fetching drop from database: %v", err)
			httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch drop: "+err.Error())
		}
		return
	}

	log.Printf("Successfully fetched drop with ID: %s", drop.ID.String())
	httputils.RespondWithJSON(w, http.StatusOK, drop)
}

// ListDropsHandler handles fetching all drops for a user.
// GET /api/v1/drops
func (h *DropsHandler) ListDropsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only GET method is allowed")
		return
	}

	// Placeholder for UserID until proper auth is implemented.
	// In a real app, this would come from auth context.
	// For now, we'll use a default user ID to list drops.
	// This should be consistent with the UserID used in CreateDropHandler.
	userID := "default-user" // Placeholder
	log.Printf("Attempting to list drops for UserID: %s", userID)

	drops, err := h.APIConfig.DB.ListDrops(r.Context(), sql.NullString{String: userID, Valid: true})
	if err != nil {
		log.Printf("Error fetching drops from database: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch drops: "+err.Error())
		return
	}

	// If no drops are found, drops will be an empty slice (nil or zero-length).
	// Responding with an empty JSON array [] is the correct behavior for this case.
	if drops == nil {
		drops = []db.Drop{} // Ensure a non-nil slice for JSON marshaling as []
	}

	log.Printf("Successfully fetched %d drops for UserID: %s", len(drops), userID)
	httputils.RespondWithJSON(w, http.StatusOK, drops)
}

// UpdateDropHandler handles updating an existing drop.
// PUT /api/v1/drops/{id}
func (h *DropsHandler) UpdateDropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only PUT method is allowed")
		return
	}

	dropIDStr := r.PathValue("id") // Available in Go 1.22+
	if dropIDStr == "" {
		httputils.RespondWithError(w, http.StatusBadRequest, "Drop ID is required in the path")
		return
	}

	dropID, err := uuid.Parse(dropIDStr)
	if err != nil {
		httputils.RespondWithError(w, http.StatusBadRequest, "Invalid Drop ID format: "+err.Error())
		return
	}

	var req UpdateDropRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Placeholder for UserID until proper auth is implemented.
	// This should be consistent with the UserID used in CreateDropHandler.
	userID := "default-user" // Placeholder
	log.Printf("Attempting to update drop with ID: %s for UserID: %s", dropID.String(), userID)

	params := db.UpdateDropParams{
		ID:     dropID,
		UserID: sql.NullString{String: userID, Valid: true},
	}

	if req.Topic != nil {
		if strings.TrimSpace(*req.Topic) == "" {
			httputils.RespondWithError(w, http.StatusBadRequest, "Topic cannot be empty if provided")
			return
		}
		params.Topic = sql.NullString{String: *req.Topic, Valid: true}
	}
	if req.URL != nil {
		if strings.TrimSpace(*req.URL) == "" {
			httputils.RespondWithError(w, http.StatusBadRequest, "URL cannot be empty if provided")
			return
		}
		params.Url = sql.NullString{String: *req.URL, Valid: true} // Note: DB schema field is 'url'
	}
	if req.UserNotes != nil {
		params.UserNotes = sql.NullString{String: *req.UserNotes, Valid: true}
	}
	if req.Priority != nil {
		params.Priority = sql.NullInt32{Int32: *req.Priority, Valid: true}
	}
	if req.Status != nil {
		// Basic validation for status enum, can be expanded
		validStatuses := map[string]bool{"new": true, "sent": true, "archived": true, "snoozed": true}
		if !validStatuses[*req.Status] {
			httputils.RespondWithError(w, http.StatusBadRequest, "Invalid status value. Allowed: new, sent, archived, snoozed.")
			return
		}
		params.Status = sql.NullString{String: *req.Status, Valid: true}
	}

	updatedDrop, err := h.APIConfig.DB.UpdateDrop(r.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Drop with ID %s not found or user %s not authorized to update", dropID.String(), userID)
			httputils.RespondWithError(w, http.StatusNotFound, "Drop not found or not authorized to update")
		} else {
			log.Printf("Error updating drop in database: %v", err)
			httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to update drop: "+err.Error())
		}
		return
	}

	log.Printf("Successfully updated drop with ID: %s", updatedDrop.ID.String())
	httputils.RespondWithJSON(w, http.StatusOK, updatedDrop)
}

// DeleteDropHandler handles deleting an existing drop.
// DELETE /api/v1/drops/{id}
func (h *DropsHandler) DeleteDropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only DELETE method is allowed")
		return
	}

	dropIDStr := r.PathValue("id") // Available in Go 1.22+
	if dropIDStr == "" {
		httputils.RespondWithError(w, http.StatusBadRequest, "Drop ID is required in the path")
		return
	}

	dropID, err := uuid.Parse(dropIDStr)
	if err != nil {
		httputils.RespondWithError(w, http.StatusBadRequest, "Invalid Drop ID format: "+err.Error())
		return
	}

	// Placeholder for UserID until proper auth is implemented.
	// This ensures that a user can only delete their own drops.
	userID := "default-user" // Placeholder
	log.Printf("Attempting to delete drop with ID: %s for UserID: %s", dropID.String(), userID)

	params := db.DeleteDropParams{
		ID:     dropID,
		UserID: sql.NullString{String: userID, Valid: true},
	}

	err = h.APIConfig.DB.DeleteDrop(r.Context(), params)
	if err != nil {
		// Note: A simple DELETE without RETURNING won't return sql.ErrNoRows if the item didn't exist.
		// It will only return an error for actual execution problems.
		// If we needed to confirm a row was deleted, we'd need a different approach (e.g., checking rows affected, or using RETURNING).
		// For a DELETE operation, if no error occurs, it's generally considered successful from the client's perspective
		// (the resource is gone or was already gone).
		log.Printf("Error deleting drop from database: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete drop: "+err.Error())
		return
	}

	log.Printf("Successfully deleted drop with ID: %s (or it did not exist for user %s)", dropID.String(), userID)
	w.WriteHeader(http.StatusNoContent) // 204 No Content is standard for successful DELETE
}
