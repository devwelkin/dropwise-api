package main

import (
	"log"

	"github.com/twomotive/dropwise/internal/config"
	"github.com/twomotive/dropwise/internal/worker"
)

func main() {
	log.Println("Starting Dropwise Worker Process (Simulation)...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration for worker: %v", err)
	}

	// Call the worker function to process due drops
	worker.ProcessDueDrops(cfg)

	log.Println("Dropwise Worker Process (Simulation) finished.")
}
