package migrations

import (
	"reflect"
	"time"

	"study1/internal/core/database"
	"study1/internal/modules/user"

	"gorm.io/gorm"
)

// tableNameOf returns the inferred table name for a model (simple plural snake_case).
func tableNameOf(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	// simple snake_case conversion + plural
	// reuse basic heuristic: lowercased + "s"
	// For more accurate column names, models can implement TableName().
	return toSnakeSimple(name) + "s"
}

func toSnakeSimple(s string) string {
	var result []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			if r >= 'A' && r <= 'Z' {
				r = r + ('a' - 'A')
			}
			result = append(result, r)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

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
	// Migrate only models that don't have tables yet and record per-model migration.
	migratedAny := false
	for _, model := range models {
		tableName := tableNameOf(model)
		if migrator.TableExists(model) {
			// already migrated
			continue
		}

		// Migrate this single model
		if err := migrator.AutoMigrate(model); err != nil {
			return err
		}

		// Record migration for this model
		version := time.Now().Format("20060102150405")
		name := "migrate_" + tableName
		if err := migrator.RecordMigration(version, name); err != nil {
			return err
		}
		migratedAny = true
	}

	if !migratedAny {
		// nothing to do
		return nil
	}

	return nil
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

// ApplyRegistered applies SQL migrations registered in the central registry.
// Only migrations that are not yet recorded will be executed.
func (m *Migration) ApplyRegistered() error {
	migrator := database.NewMigrator(m.DB)

	regs := database.GetMigrations()
	for _, reg := range regs {
		applied, err := migrator.HasMigrationRecord(reg.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		if reg.Up != "" {
			if err := m.DB.Exec(reg.Up).Error; err != nil {
				return err
			}
		}

		if err := migrator.RecordMigration(reg.Version, reg.Name); err != nil {
			return err
		}
	}

	return nil
}

// RollbackRegistered rolls back the specified registered migration (by version).
// If version is empty, it will roll back the last applied migration.
func (m *Migration) RollbackRegistered(version string) error {
	migrator := database.NewMigrator(m.DB)

	// Determine target version
	target := version
	if target == "" {
		var rec database.MigrationRecord
		if err := m.DB.Order("applied_at desc").First(&rec).Error; err != nil {
			return err
		}
		target = rec.Version
	}

	reg := database.GetMigrationByVersion(target)
	if reg == nil {
		return gorm.ErrRecordNotFound
	}

	if reg.Down != "" {
		if err := m.DB.Exec(reg.Down).Error; err != nil {
			return err
		}
	}

	if err := migrator.RemoveMigrationRecord(target); err != nil {
		return err
	}

	return nil
}
