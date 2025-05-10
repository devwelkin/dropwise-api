package server

import (
	"net/http"

	"github.com/twomotive/dropwise/internal/config"
	"github.com/twomotive/dropwise/internal/handlers"
	"github.com/twomotive/dropwise/internal/server/httputils" // Though not directly used here, good to keep for consistency if middleware is added
)

// NewRouter creates and configures a new HTTP ServeMux with all application routes.
func NewRouter(apiCfg *config.APIConfig) *http.ServeMux {
	mux := http.NewServeMux()

	// Initialize handlers
	dropsHandler := handlers.NewDropsHandler(apiCfg)

	// --- Route Definitions ---

	// Health check / Root path
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		httputils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "API is running"})
	})

	// POST /api/v1/drops - Create a new drop
	mux.HandleFunc("POST /api/v1/drops", dropsHandler.CreateDropHandler)

	// GET /api/v1/drops/{id} - Get a specific drop
	mux.HandleFunc("GET /api/v1/drops/{id}", dropsHandler.GetDropHandler)

	return mux
}
