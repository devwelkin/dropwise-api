package server

import (
	"net/http"

	"github.com/twomotive/dropwise/internal/config"
	"github.com/twomotive/dropwise/internal/handlers"
	"github.com/twomotive/dropwise/internal/middleware"
	"github.com/twomotive/dropwise/internal/server/httputils"
)

// NewRouter creates and newServeMux with all application routes.
func NewRouter(apiCfg *config.APIConfig) *http.ServeMux {
	mux := http.NewServeMux()

	// Initialize handlers
	dropsHandler := handlers.NewDropsHandler(apiCfg)
	tagsHandler := handlers.NewTagsHandler(apiCfg)
	authHandler := handlers.NewAuthHandler(apiCfg) // New Auth Handler

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(apiCfg.JWTSecret)
	loggingMiddleware := middleware.LoggingMiddleware

	// --- Route Definitions ---

	// Health check / Root path
	mux.HandleFunc("GET /", middleware.ApplyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		httputils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "API is running"})
	}, loggingMiddleware))

	// --- Authentication Endpoints ---
	// These endpoints don't need authentication but should be logged
	mux.HandleFunc("POST /api/v1/auth/register", middleware.ApplyMiddleware(authHandler.RegisterHandler, loggingMiddleware))
	mux.HandleFunc("POST /api/v1/auth/login", middleware.ApplyMiddleware(authHandler.LoginHandler, loggingMiddleware))

	// --- Drop Endpoints ---
	// POST /api/v1/drops - Create a new drop (protected)
	mux.HandleFunc("POST /api/v1/drops", middleware.Chain(dropsHandler.CreateDropHandler,
		loggingMiddleware, authMiddleware))

	// GET /api/v1/drops/{id} - Get a specific drop (protected)
	mux.HandleFunc("GET /api/v1/drops/{id}", middleware.Chain(dropsHandler.GetDropHandler,
		loggingMiddleware, authMiddleware))

	// GET /api/v1/drops - List all drops for a user (protected)
	mux.HandleFunc("GET /api/v1/drops", middleware.Chain(dropsHandler.ListDropsHandler,
		loggingMiddleware, authMiddleware))

	// PUT /api/v1/drops/{id} - Update a specific drop (protected)
	mux.HandleFunc("PUT /api/v1/drops/{id}", middleware.Chain(dropsHandler.UpdateDropHandler,
		loggingMiddleware, authMiddleware))

	// DELETE /api/v1/drops/{id} - Delete a specific drop (protected)
	mux.HandleFunc("DELETE /api/v1/drops/{id}", middleware.Chain(dropsHandler.DeleteDropHandler,
		loggingMiddleware, authMiddleware))

	// --- Tag Endpoints ---
	// GET /api/v1/tags - List all unique tags (protected)
	mux.HandleFunc("GET /api/v1/tags", middleware.Chain(tagsHandler.ListTagsHandler,
		loggingMiddleware, authMiddleware))

	return mux
}
