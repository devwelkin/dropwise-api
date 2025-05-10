// filepath: cmd/api/main.go
package main

import (
	"log"
	"net/http"

	"github.com/twomotive/dropwise/internal/config"
	"github.com/twomotive/dropwise/internal/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	mux := server.NewRouter(cfg)

	log.Printf("Starting server on port %s", cfg.Port)

	// Start the HTTP server
	serverAddr := ":" + cfg.Port
	log.Printf("API server listening on %s", serverAddr)
	err = http.ListenAndServe(serverAddr, mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
