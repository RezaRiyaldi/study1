package database

import (
	"fmt"
	"log"
	"reflect"

	"gorm.io/gorm"
)

type Migrator struct {
	db *gorm.DB
}

func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
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
