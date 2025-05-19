package server

import (
	"net/http"

	"github.com/twomotive/dropwise/internal/config"
	"github.com/twomotive/dropwise/internal/handlers"
	"github.com/twomotive/dropwise/internal/server/httputils"
)

// NewRouter creates and newServeMux with all application routes.
func NewRouter(apiCfg *config.APIConfig) *http.ServeMux {
	mux := http.NewServeMux()

	// Initialize handlers
	dropsHandler := handlers.NewDropsHandler(apiCfg)
	tagsHandler := handlers.NewTagsHandler(apiCfg)
	authHandler := handlers.NewAuthHandler(apiCfg) // New Auth Handler

	// --- Route Definitions ---

	// Health check / Root path
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		httputils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "API is running"})
	})

	// --- Authentication Endpoints ---
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.RegisterHandler)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.LoginHandler)

	// --- Drop Endpoints ---
	// POST /api/v1/drops - Create a new drop
	mux.HandleFunc("POST /api/v1/drops", dropsHandler.CreateDropHandler)

	// GET /api/v1/drops/{id} - Get a specific drop
	mux.HandleFunc("GET /api/v1/drops/{id}", dropsHandler.GetDropHandler)

	// GET /api/v1/drops - List all drops for a user
	mux.HandleFunc("GET /api/v1/drops", dropsHandler.ListDropsHandler)

	// PUT /api/v1/drops/{id} - Update a specific drop
	mux.HandleFunc("PUT /api/v1/drops/{id}", dropsHandler.UpdateDropHandler)

	// DELETE /api/v1/drops/{id} - Delete a specific drop
	mux.HandleFunc("DELETE /api/v1/drops/{id}", dropsHandler.DeleteDropHandler)

	// --- Tag Endpoints ---
	// GET /api/v1/tags - List all unique tags
	mux.HandleFunc("GET /api/v1/tags", tagsHandler.ListTagsHandler)

	return mux
}
