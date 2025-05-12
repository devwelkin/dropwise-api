package main

import (
	"context"
	"log"

	"github.com/twomotive/dropwise/internal/config"
	"github.com/twomotive/dropwise/worker"
)

func main() {
	log.Println("Starting Dropwise Worker Process (Simulation)...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration for worker: %v", err)
	}

	// Call the core worker logic directly for command-line simulation
	// Pass a background context
	processedCount, err := worker.ProcessDropsLogic(context.Background(), cfg)
	if err != nil {
		log.Printf("Worker simulation finished with error: %v", err)
	} else {
		log.Printf("Worker simulation finished. Drops processed: %d", processedCount)
	}

	log.Println("Dropwise Worker Process (Simulation) finished.")
}
