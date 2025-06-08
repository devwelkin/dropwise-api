package handlers

import (
	"log"
	"net/http"

	"github.com/nouvadev/dropwise/internal/config"
	db "github.com/nouvadev/dropwise/internal/database/sqlc"
	"github.com/nouvadev/dropwise/internal/server/httputils"
)

// TagsHandler handles HTTP requests for tags.
type TagsHandler struct {
	APIConfig *config.APIConfig
}

// NewTagsHandler creates a new TagsHandler.
func NewTagsHandler(apiCfg *config.APIConfig) *TagsHandler {
	return &TagsHandler{APIConfig: apiCfg}
}

// ListTagsHandler handles fetching all unique tags.
// GET /api/v1/tags
func (h *TagsHandler) ListTagsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only GET method is allowed")
		return
	}

	log.Println("Attempting to list all tags")

	tags, err := h.APIConfig.DB.ListTags(r.Context())
	if err != nil {
		log.Printf("Error fetching tags from database: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch tags: "+err.Error())
		return
	}

	// Ensure a non-nil slice for JSON marshaling as [] if no tags are found.
	// sqlc typically returns an empty slice, but this is a good safeguard.
	if tags == nil {
		tags = []db.Tag{}
	}

	log.Printf("Successfully fetched %d tags", len(tags))
	httputils.RespondWithJSON(w, http.StatusOK, tags)
}
