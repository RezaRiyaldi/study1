package main

import (
	"log"
	_ "study1/docs"
	"study1/internal/core/app"
	"study1/internal/core/config"
)

// @title Study1 API
// @version 1.0
// @description This is a sample server for Study1.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
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
