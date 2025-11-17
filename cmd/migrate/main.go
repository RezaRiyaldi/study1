package main

import (
	"log"
	"os"
	"path/filepath"
	"study1/internal/core/config"
	"study1/internal/core/database"
	"study1/internal/core/database/migrations"
	"study1/internal/modules/user"

	"gorm.io/gorm"
)

func generateMigrations(db *gorm.DB) {
	// Define models to generate migrations for
	models := []interface{}{
		&user.User{},
		// Add more models here as needed
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	migrationsDir := filepath.Join(cwd, "internal", "core", "database", "migrations", "generated")

	generator := database.NewMigrationGenerator(db, migrationsDir)

	log.Println("🔄 Generating migrations from models...")

	if err := generator.GenerateFromModels(models...); err != nil {
		log.Fatalf("Failed to generate migrations: %v", err)
	}

	log.Printf("✅ Migrations generated successfully in: %s", migrationsDir)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate <command> \nCommands: generate, up, down, refresh, fresh")
	}

	command := os.Args[1]

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize migrator
	migration := migrations.NewMigration(db.DB)

	switch command {
	case "generate", "gen":
		generateMigrations(db.DB)

	case "up", "migrate":
		log.Println("Running migrations...")
		if err := migration.RunAll(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("✅ Migrations completed successfully")

	case "down", "drop":
		log.Println("Dropping all tables...")
		if err := migration.DropAll(); err != nil {
			log.Fatalf("Drop tables failed: %v", err)
		}
		log.Println("✅ All tables dropped successfully")

	case "refresh":
		log.Println("Refreshing database...")
		if err := migration.Refresh(); err != nil {
			log.Fatalf("Refresh failed: %v", err)
		}
		log.Println("✅ Database refreshed successfully")

	case "fresh":
		log.Println("Dropping all tables...")
		if err := migration.DropAll(); err != nil {
			log.Fatalf("Drop tables failed: %v", err)
		}
		log.Println("Running migrations...")
		if err := migration.RunAll(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("✅ Database recreated successfully")

	default:
		log.Fatal("Invalid command. Available commands: generate, up, down, refresh, fresh")
	}
}
