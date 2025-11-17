package main

import (
	"log"
	"study1/internal/core/app"
	"study1/internal/core/config"
)

func main() {
	// Load Config
	cfg := config.LoadConfig()

	// Initialize application
	application, err := app.New(cfg)

	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Start application
	log.Printf("Starting application on port %s... in %s mode",
		cfg.Server.Port, cfg.Server.Environtment,
	)

	if err := application.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
