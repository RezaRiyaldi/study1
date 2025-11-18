package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"study1/internal/core/config"
	"study1/internal/core/database"
	"study1/internal/core/database/migrations"
	"study1/internal/modules/activity"
	"study1/internal/modules/user"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func generateMigrations(db *gorm.DB) {
	// Define models to generate migrations for
	models := []interface{}{
		&activity.ActivityLog{},
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

	// Before initializing GORM, check whether the database exists and offer to create it.
	ok, err := checkAndOfferCreateDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to check database existence: %v", err)
	}
	if !ok {
		log.Println("Database does not exist and was not created. Skipping migrations.")
		os.Exit(0)
	}

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

// checkAndOfferCreateDB checks if the configured database exists. If not, it prompts
// the user whether to create it. Returns true if the DB exists (or was created),
// false if the DB does not exist and user chose not to create it.
func checkAndOfferCreateDB(cfg config.DatabaseConfig) (bool, error) {
	// connect without specifying database
	dsn := cfg.GetDSNNoDB()

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, err
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return false, fmt.Errorf("cannot connect to DB server: %w", err)
	}

	// check existence
	var name string
	row := sqlDB.QueryRow("SELECT SCHEMA_NAME FROM information_schema.schemata WHERE schema_name = ?", cfg.Name)
	if err := row.Scan(&name); err == nil {
		// exists
		return true, nil
	} else if err != sql.ErrNoRows {
		return false, err
	}

	// Not exists: prompt
	fmt.Printf("Database '%s' does not exist. Create it now? (y/N): ", cfg.Name)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "y" || input == "yes" {
		createStmt := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", cfg.Name)
		if _, err := sqlDB.Exec(createStmt); err != nil {
			return false, fmt.Errorf("failed to create database: %w", err)
		}
		fmt.Printf("Database '%s' created successfully.\n", cfg.Name)
		return true, nil
	}

	// user chose not to create
	return false, nil
}
