// filepath: cmd/api/main.go
package main

import (
	"log"
	"net/http"

	"github.com/nouvadev/dropwise/internal/config"
	"github.com/nouvadev/dropwise/internal/server"
	"github.com/rs/cors"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	mux := server.NewRouter(cfg)
	// Configure CORS
	c := cors.New(cors.Options{
		// İzin verilen frontend adresleri. KENDİ VERCEL URL'Nİ YAZMALISIN.
		AllowedOrigins: []string{"https://dropwise.vercel.app", "http://localhost:5173"},

		// İzin verilen HTTP metodları
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},

		// İzin verilen HTTP header'ları
		AllowedHeaders: []string{"Authorization", "Content-Type"},

		// Tarayıcının preflight (OPTIONS) cevabını cache'lemesi için süre (saniye)
		MaxAge: 86400,
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
