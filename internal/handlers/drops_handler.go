package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/twomotive/dropwise/internal/config"
	db "github.com/twomotive/dropwise/internal/database/sqlc"
	"github.com/twomotive/dropwise/internal/middleware" // Ensure middleware is imported
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
	Topic     string   `json:"topic"`
	URL       string   `json:"url"`
	UserNotes string   `json:"user_notes,omitempty"`
	Priority  *int32   `json:"priority,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// UpdateDropRequest defines the expected request body for updating a drop.
type UpdateDropRequest struct {
	Topic     *string   `json:"topic,omitempty"`
	URL       *string   `json:"url,omitempty"`
	UserNotes *string   `json:"user_notes,omitempty"`
	Priority  *int32    `json:"priority,omitempty"`
	Status    *string   `json:"status,omitempty"` // e.g., "new", "sent", "archived"
	Tags      *[]string `json:"tags,omitempty"`
}

// DropResponse defines the structure for drop responses.
type DropResponse struct {
	ID           uuid.UUID      `json:"id"`
	UserUuid     uuid.NullUUID  `json:"user_uuid"`
	Topic        string         `json:"topic"`
	Url          string         `json:"url"`
	UserNotes    sql.NullString `json:"user_notes"`
	AddedDate    time.Time      `json:"added_date"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Status       string         `json:"status"`
	LastSentDate sql.NullTime   `json:"last_sent_date"`
	SendCount    int32          `json:"send_count"`
	Priority     sql.NullInt32  `json:"priority"`
	Tags         []string       `json:"tags"`
}

// toDropResponse converts a db.Drop and its tags to a DropResponse.
func toDropResponse(drop db.Drop, tags []db.Tag) DropResponse {
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	return DropResponse{
		ID:           drop.ID,
		UserUuid:     drop.UserUuid,
		Topic:        drop.Topic,
		Url:          drop.Url,
		UserNotes:    drop.UserNotes,
		AddedDate:    drop.AddedDate,
		UpdatedAt:    drop.UpdatedAt,
		Status:       drop.Status,
		LastSentDate: drop.LastSentDate,
		SendCount:    drop.SendCount,
		Priority:     drop.Priority,
		Tags:         tagNames,
	}
}

// CreateDropHandler handles the creation of a new drop.
// POST /api/v1/drops
func (h *DropsHandler) CreateDropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		log.Printf("CreateDropHandler: UserID not found in context or not a UUID for path %s", r.URL.Path)
		httputils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateDropRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(req.Topic) == "" {
		httputils.RespondWithError(w, http.StatusBadRequest, "Topic cannot be empty")
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		httputils.RespondWithError(w, http.StatusBadRequest, "URL cannot be empty")
		return
	}

	params := db.CreateDropParams{
		UserUuid: uuid.NullUUID{UUID: userUUID, Valid: true},
		Topic:    req.Topic,
		Url:      req.URL,
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

	log.Printf("Attempting to create drop for UserUUID: %s, Topic: %s", userUUID, params.Topic)

	createdDrop, err := h.APIConfig.DB.CreateDrop(r.Context(), params)
	if err != nil {
		log.Printf("Error creating drop in database: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to create drop: "+err.Error())
		return
	}

	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			trimmedTagName := strings.TrimSpace(tagName)
			if trimmedTagName == "" {
				continue
			}
			tag, err := h.APIConfig.DB.CreateTag(r.Context(), trimmedTagName)
			if err != nil {
				log.Printf("Error creating/getting tag '%s': %v", trimmedTagName, err)
				continue
			}
			err = h.APIConfig.DB.AddTagToDrop(r.Context(), db.AddTagToDropParams{
				DropsID: createdDrop.ID,
				TagID:   tag.ID,
			})
			if err != nil {
				log.Printf("Error associating tag '%s' (ID: %d) with drop '%s': %v", trimmedTagName, tag.ID, createdDrop.ID, err)
				continue
			}
		}
	}

	finalTags, err := h.APIConfig.DB.GetTagsForDrop(r.Context(), createdDrop.ID)
	if err != nil {
		log.Printf("Error fetching tags for drop %s after creation: %v", createdDrop.ID, err)
		finalTags = []db.Tag{}
	}

	log.Printf("Successfully created drop with ID: %s", createdDrop.ID.String())
	response := toDropResponse(createdDrop, finalTags)
	httputils.RespondWithJSON(w, http.StatusCreated, response)
}

// GetDropHandler handles fetching a specific drop.
// GET /api/v1/drops/{id}
func (h *DropsHandler) GetDropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only GET method is allowed")
		return
	}

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		log.Printf("GetDropHandler: UserID not found in context or not a UUID for path %s", r.URL.Path)
		httputils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
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

	log.Printf("Attempting to fetch drop with ID: %s for UserUUID: %s", dropID.String(), userUUID.String())

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

	if !drop.UserUuid.Valid || drop.UserUuid.UUID != userUUID {
		log.Printf("Authorization failed: User %s attempted to access drop %s owned by %s",
			userUUID.String(), drop.ID.String(), drop.UserUuid.UUID.String())
		httputils.RespondWithError(w, http.StatusForbidden, "Access to this drop is forbidden")
		return
	}

	tags, err := h.APIConfig.DB.GetTagsForDrop(r.Context(), drop.ID)
	if err != nil {
		log.Printf("Error fetching tags for drop %s: %v", drop.ID, err)
		tags = []db.Tag{}
	}

	log.Printf("Successfully fetched drop with ID: %s and %d tags", drop.ID.String(), len(tags))
	response := toDropResponse(drop, tags)
	httputils.RespondWithJSON(w, http.StatusOK, response)
}

// ListDropsHandler handles fetching all drops for the authenticated user.
// GET /api/v1/drops
func (h *DropsHandler) ListDropsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only GET method is allowed")
		return
	}

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		log.Printf("ListDropsHandler: UserID not found in context or not a UUID for path %s", r.URL.Path)
		httputils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("Attempting to list drops for UserUUID: %s", userUUID.String())

	drops, err := h.APIConfig.DB.ListDropsByUserUUID(r.Context(), uuid.NullUUID{UUID: userUUID, Valid: true})
	if err != nil {
		log.Printf("Error fetching drops from database for UserUUID %s: %v", userUUID.String(), err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch drops: "+err.Error())
		return
	}

	if drops == nil {
		drops = []db.Drop{}
	}

	dropResponses := make([]DropResponse, 0, len(drops))
	for _, drop := range drops {
		tags, err := h.APIConfig.DB.GetTagsForDrop(r.Context(), drop.ID)
		if err != nil {
			log.Printf("Error fetching tags for drop %s during list operation: %v. Proceeding with empty tags for this drop.", drop.ID, err)
			tags = []db.Tag{}
		}
		dropResponses = append(dropResponses, toDropResponse(drop, tags))
	}

	log.Printf("Successfully fetched %d drops for UserUUID: %s", len(dropResponses), userUUID.String())
	httputils.RespondWithJSON(w, http.StatusOK, dropResponses)
}

// UpdateDropHandler handles updating an existing drop.
// PUT /api/v1/drops/{id}
func (h *DropsHandler) UpdateDropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only PUT method is allowed")
		return
	}

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		log.Printf("UpdateDropHandler: UserID not found in context or not a UUID for path %s", r.URL.Path)
		httputils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
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

	var req UpdateDropRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	log.Printf("Attempting to update drop with ID: %s for UserUUID: %s", dropID.String(), userUUID.String())

	// First, verify the drop exists and belongs to the user.
	// This is important for UpdateDrop to ensure the user owns the drop they are trying to update.
	// The UpdateDrop SQL query itself also checks user_uuid, but this provides a clearer error.
	existingDrop, err := h.APIConfig.DB.GetDrop(r.Context(), dropID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Update failed: Drop with ID %s not found for UserUUID %s", dropID.String(), userUUID.String())
			httputils.RespondWithError(w, http.StatusNotFound, "Drop not found")
		} else {
			log.Printf("Error checking drop existence before update: %v", err)
			httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to update drop: "+err.Error())
		}
		return
	}

	if !existingDrop.UserUuid.Valid || existingDrop.UserUuid.UUID != userUUID {
		log.Printf("Authorization failed: User %s attempted to update drop %s owned by %s",
			userUUID.String(), existingDrop.ID.String(), existingDrop.UserUuid.UUID.String())
		httputils.RespondWithError(w, http.StatusForbidden, "Not authorized to update this drop")
		return
	}

	params := db.UpdateDropParams{
		ID:       dropID,
		UserUuid: uuid.NullUUID{UUID: userUUID, Valid: true},
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
		params.Url = sql.NullString{String: *req.URL, Valid: true}
	}
	if req.UserNotes != nil {
		params.UserNotes = sql.NullString{String: *req.UserNotes, Valid: true}
	}
	if req.Priority != nil {
		params.Priority = sql.NullInt32{Int32: *req.Priority, Valid: true}
	}
	if req.Status != nil {
		validStatuses := map[string]bool{"new": true, "sent": true, "archived": true, "snoozed": true}
		if !validStatuses[*req.Status] {
			httputils.RespondWithError(w, http.StatusBadRequest, "Invalid status value. Allowed: new, sent, archived, snoozed.")
			return
		}
		params.Status = sql.NullString{String: *req.Status, Valid: true}
	}

	updatedDrop, err := h.APIConfig.DB.UpdateDrop(r.Context(), params)
	if err != nil {
		// sql.ErrNoRows might occur if the record was deleted between the GetDrop check and UpdateDrop,
		// or if the user_uuid check in the UPDATE query fails (though our GetDrop check should prevent this).
		if err == sql.ErrNoRows {
			log.Printf("Drop with ID %s not found or user %s not authorized to update (during DB.UpdateDrop)", dropID.String(), userUUID.String())
			httputils.RespondWithError(w, http.StatusNotFound, "Drop not found or not authorized to update")
		} else {
			log.Printf("Error updating drop in database: %v", err)
			httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to update drop: "+err.Error())
		}
		return
	}

	if req.Tags != nil {
		log.Printf("Updating tags for drop ID: %s", dropID.String())
		err = h.APIConfig.DB.RemoveAllTagsFromDrop(r.Context(), dropID)
		if err != nil {
			log.Printf("Error removing existing tags for drop %s: %v", dropID, err)
			// Continue to add new tags even if removal failed, though this might lead to duplicates if not handled.
		}

		if len(*req.Tags) > 0 {
			for _, tagName := range *req.Tags {
				trimmedTagName := strings.TrimSpace(tagName)
				if trimmedTagName == "" {
					continue
				}
				tag, err := h.APIConfig.DB.CreateTag(r.Context(), trimmedTagName)
				if err != nil {
					log.Printf("Error creating/getting tag '%s' for drop %s: %v", trimmedTagName, dropID, err)
					continue
				}
				err = h.APIConfig.DB.AddTagToDrop(r.Context(), db.AddTagToDropParams{
					DropsID: dropID,
					TagID:   tag.ID,
				})
				if err != nil {
					log.Printf("Error associating tag '%s' (ID: %d) with drop '%s': %v", trimmedTagName, tag.ID, dropID, err)
				}
			}
		}
		log.Printf("Finished updating tags for drop ID: %s", dropID.String())
	}

	finalTags, err := h.APIConfig.DB.GetTagsForDrop(r.Context(), updatedDrop.ID)
	if err != nil {
		log.Printf("Error fetching tags for drop %s after update: %v", updatedDrop.ID, err)
		finalTags = []db.Tag{}
	}

	log.Printf("Successfully updated drop with ID: %s and its tags", updatedDrop.ID.String())
	response := toDropResponse(updatedDrop, finalTags)
	httputils.RespondWithJSON(w, http.StatusOK, response)
}

// DeleteDropHandler handles deleting an existing drop.
// DELETE /api/v1/drops/{id}
func (h *DropsHandler) DeleteDropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only DELETE method is allowed")
		return
	}

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		log.Printf("DeleteDropHandler: UserID not found in context or not a UUID for path %s", r.URL.Path)
		httputils.RespondWithError(w, http.StatusUnauthorized, "User not authenticated")
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

	log.Printf("Attempting to delete drop with ID: %s for UserUUID: %s", dropID.String(), userUUID.String())

	existingDrop, err := h.APIConfig.DB.GetDrop(r.Context(), dropID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Delete failed: Drop with ID %s not found for UserUUID %s", dropID.String(), userUUID.String())
			httputils.RespondWithError(w, http.StatusNotFound, "Drop not found")
		} else {
			log.Printf("Error checking drop existence before delete: %v", err)
			httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete drop: "+err.Error())
		}
		return
	}

	if !existingDrop.UserUuid.Valid || existingDrop.UserUuid.UUID != userUUID {
		log.Printf("Authorization failed: User %s attempted to delete drop %s owned by %s",
			userUUID.String(), existingDrop.ID.String(), existingDrop.UserUuid.UUID.String())
		httputils.RespondWithError(w, http.StatusForbidden, "Not authorized to delete this drop")
		return
	}

	params := db.DeleteDropParams{
		ID:       dropID,
		UserUuid: uuid.NullUUID{UUID: userUUID, Valid: true},
	}

	err = h.APIConfig.DB.DeleteDrop(r.Context(), params)
	if err != nil {
		// This error should ideally not be sql.ErrNoRows if the check above passed and the DB.DeleteDrop also checks user_uuid.
		// It would typically indicate other execution errors.
		log.Printf("Error deleting drop from database: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete drop: "+err.Error())
		return
	}

	log.Printf("Successfully deleted drop with ID: %s for UserUUID: %s", dropID.String(), userUUID.String())
	w.WriteHeader(http.StatusNoContent)
}
