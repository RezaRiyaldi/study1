package database

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"gorm.io/gorm"
)

type Migrator struct {
	db *gorm.DB
}

func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// MigrationRecord represents a record in the migrations table.
type MigrationRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Version   string    `gorm:"uniqueIndex;size:64"`
	Name      string    `gorm:"size:255"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}

// ensureMigrationsTable ensures the migrations table exists.
func (m *Migrator) ensureMigrationsTable() error {
	return m.db.AutoMigrate(&MigrationRecord{})
}

// RecordMigration inserts a migration record if it does not already exist.
func (m *Migrator) RecordMigration(version, name string) error {
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}

	var rec MigrationRecord
	if err := m.db.Where("version = ?", version).First(&rec).Error; err == nil {
		// already recorded
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("check migration record: %w", err)
	}

	rec = MigrationRecord{
		Version:   version,
		Name:      name,
		AppliedAt: time.Now(),
	}

	if err := m.db.Create(&rec).Error; err != nil {
		return fmt.Errorf("create migration record: %w", err)
	}
	return nil
}

// HasMigrationRecord checks if a migration version has been recorded.
func (m *Migrator) HasMigrationRecord(version string) (bool, error) {
	if err := m.ensureMigrationsTable(); err != nil {
		return false, err
	}
	var rec MigrationRecord
	if err := m.db.Where("version = ?", version).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// RemoveMigrationRecord deletes a migration record by version.
func (m *Migrator) RemoveMigrationRecord(version string) error {
	if err := m.ensureMigrationsTable(); err != nil {
		return err
	}
	if err := m.db.Where("version = ?", version).Delete(&MigrationRecord{}).Error; err != nil {
		return err
	}
	return nil
}

// AutoMigrate automatically migrates all models
func (m *Migrator) AutoMigrate(models ...interface{}) error {
	log.Println("🔄 Starting database migration...")

	for _, model := range models {
		tableName := getTableName(model)
		log.Printf("Migrating table: %s", tableName)

		if err := m.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate table %s: %w", tableName, err)
		}
	}

	log.Println("✅ Database migration completed successfully")
	return nil
}

// DropTables drops all tables (for development only)
func (m *Migrator) DropTables(models ...interface{}) error {
	log.Println("🗑️  Dropping all tables...")

	// Disable foreign key checks
	m.db.Exec("SET FOREIGN_KEY_CHECKS = 0")

	for _, model := range models {
		tableName := getTableName(model)
		log.Printf("Dropping table: %s", tableName)

		if err := m.db.Migrator().DropTable(model); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", tableName, err)
		}
	}

	// Enable foreign key checks
	m.db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	log.Println("✅ All tables dropped successfully")
	return nil
}

// Get table name from model
func getTableName(model interface{}) string {
	if tableNamer, ok := model.(interface{ TableName() string }); ok {
		return tableNamer.TableName()
	}

	// Fallback: get type name and convert to snake_case plural
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()
	snakeCase := toSnakeCase(name)
	return snakeCase + "s" // Simple pluralization
}

// Refresh drops and recreates all tables
func (m *Migrator) Refresh(models ...interface{}) error {
	if err := m.DropTables(models...); err != nil {
		return err
	}
	return m.AutoMigrate(models...)
}

// Check if table exists
func (m *Migrator) TableExists(model interface{}) bool {
	tableName := getTableName(model)
	return m.db.Migrator().HasTable(tableName)
}
