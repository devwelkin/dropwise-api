// filepath: cmd/api/main.go
package main

import (
	"log"
	"net/http"

	"github.com/twomotive/dropwise/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	mux := http.NewServeMux()

	// Define a handler for the root path "/"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := w.Write([]byte("Hello, Dropwise! API is running."))
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	log.Printf("Starting server on port %s", cfg.Port)

	// Start the HTTP server
	err = http.ListenAndServe(":"+cfg.Port, mux)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
