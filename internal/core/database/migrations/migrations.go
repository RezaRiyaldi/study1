package migrations

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"study1/internal/core/database"
	"study1/internal/modules/activity"
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
		&activity.ActivityLog{},
		// Add more models here as you create them
	}

	// Generate migrations for models that don't have tables yet.
	migrationsDir := "internal/core/database/migrations/generated"
	generator := database.NewMigrationGenerator(m.DB, migrationsDir)

	generatedAny := false
	for _, model := range models {
		tableName := tableNameOf(model)
		if migrator.TableExists(model) {
			// already migrated
			continue
		}

		migrationName := "create_" + tableName + "_table"
		// Check if migration file already exists in migrations directory
		pattern := filepath.Join(migrationsDir, "*_"+migrationName+".go")
		if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
			log.Printf("⚠️  Migration file for %s already exists, skipping generation", tableName)
			continue
		}

		if err := generator.GenerateForModel(model); err != nil {
			return err
		}
		generatedAny = true
	}

	// Apply any SQL files found in the generated migrations directory.
	// This covers the case when migration .go files exist on disk but are not
	// compiled into the running binary (so their init() didn't register them).
	if err := m.applySQLMigrationsFromDir(migrationsDir); err != nil {
		return err
	}

	// Apply any migrations registered in-memory (including those generated in this run).
	if err := m.ApplyRegistered(); err != nil {
		return err
	}

	if !generatedAny {
		// nothing newly generated; ApplyRegistered already ensured any pending registered migrations were applied
		return nil
	}

	return nil
}

// applySQLMigrationsFromDir scans the given directory for *.up.sql files and
// applies any that have not been recorded yet. Files are applied in filename
// order (which includes timestamp prefix), and after successful execution the
// migration is recorded.
func (m *Migration) applySQLMigrationsFromDir(dir string) error {
	pattern := filepath.Join(dir, "*.up.sql")
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		log.Printf("No .up.sql files found in %s", dir)
		return nil
	}

	sort.Strings(matches)
	migrator := database.NewMigrator(m.DB)

	log.Printf("Found %d .up.sql files in %s", len(matches), dir)
	for _, p := range matches {
		base := filepath.Base(p)
		parts := strings.SplitN(base, "_", 2)
		if len(parts) < 2 {
			log.Printf("Skipping file with unexpected name: %s", base)
			continue
		}
		version := parts[0]
		name := strings.TrimSuffix(parts[1], ".up.sql")

		// Check whether this exact file has already been applied. We include the
		// filename in the migration record to avoid collisions when multiple
		// migrations are generated in the same second (same version).
		applied, err := migrator.HasMigrationRecordWithFile(version, name, base)
		if err != nil {
			return err
		}
		if applied {
			log.Printf("Skipping already-applied migration file: %s", base)
			continue
		}

		log.Printf("Applying SQL migration: %s (%s)", version, name)
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		if len(content) > 0 {
			// Some SQL files contain multiple statements (CREATE TABLE then CREATE INDEX).
			// The MySQL driver disallows executing multiple statements in one Exec unless
			// `multiStatements=true` is enabled in the DSN. To avoid changing DSN and
			// improve portability, split the file by semicolons and execute each
			// non-empty statement individually.
			sqlText := string(content)
			stmts := strings.Split(sqlText, ";")
			for _, stmt := range stmts {
				stmt = strings.TrimSpace(stmt)
				if stmt == "" {
					continue
				}
				if err := m.DB.Exec(stmt).Error; err != nil {
					return err
				}
			}
			log.Printf("Applied SQL migration: %s", base)
		} else {
			log.Printf("Empty SQL file, skipping execution: %s", base)
		}

		if err := migrator.RecordMigrationWithFile(version, name, base); err != nil {
			return err
		}
		log.Printf("Recorded migration: %s (file=%s)", version, base)
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
		applied, err := migrator.HasMigrationRecordVersionName(reg.Version, reg.Name)
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

	if err := migrator.RemoveMigrationRecord(target, reg.Name); err != nil {
		return err
	}

	return nil
}
