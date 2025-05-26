// filepath: cmd/api/main.go
package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/twomotive/dropwise/internal/config"
	"github.com/twomotive/dropwise/internal/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	mux := server.NewRouter(cfg)
	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Your frontend origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		Debug:            true, // Enable for development debugging
	})
	handler := c.Handler(mux)

	log.Printf("Starting server on port %s", cfg.Port)

	// Start the HTTP server
	serverAddr := ":" + cfg.Port
	log.Printf("API server listening on %s", serverAddr)
	err = http.ListenAndServe(serverAddr, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
