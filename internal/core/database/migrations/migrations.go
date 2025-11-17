package migrations

import (
	"study1/internal/core/database"
	"study1/internal/modules/user"

	"gorm.io/gorm"
)

type Migration struct {
	DB *gorm.DB
}

func NewMigration(db *gorm.DB) *Migration {
	return &Migration{DB: db}
}

// RunAll runs all migrations
func (m *Migration) RunAll() error {
	migrator := database.NewMigrator(m.DB)

	// List all models to migrate
	models := []interface{}{
		&user.User{},
		// Add more models here as you create them
	}

	return migrator.AutoMigrate(models...)
}

// DropAll drops all tables
func (m *Migration) DropAll() error {
	migrator := database.NewMigrator(m.DB)

	models := []interface{}{
		&user.User{},
		// Add more models here as you create them
	}

	return migrator.DropTables(models...)
}

// Refresh drops and recreates all tables
func (m *Migration) Refresh() error {
	migrator := database.NewMigrator(m.DB)

	models := []interface{}{
		&user.User{},
		// Add more models here as you create them
	}

	return migrator.Refresh(models...)
}
