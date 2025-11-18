package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"gorm.io/gorm"
)

type MigrationGenerator struct {
	db            *gorm.DB
	migrationsDir string
}

func NewMigrationGenerator(db *gorm.DB, migrationsDir string) *MigrationGenerator {
	return &MigrationGenerator{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// GenerateFromModels generates migration files from models
func (g *MigrationGenerator) GenerateFromModels(models ...interface{}) error {
	// Ensure migrations directory exists
	if err := os.MkdirAll(g.migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	for _, model := range models {
		tableName := getTableName(model)
		migrationName := fmt.Sprintf("create_%s_table", tableName)

		// Check if migration already exists
		if g.migrationExists(migrationName) {
			log.Printf("⚠️  Migration for %s already exists, skipping...", tableName)
			continue
		}

		// Generate migration SQL
		upSQL, downSQL, err := g.generateTableSQL(model)
		if err != nil {
			return fmt.Errorf("failed to generate SQL for %s: %w", tableName, err)
		}

		// Create migration file and register migration in memory
		version, err := g.createMigrationFile(migrationName, upSQL, downSQL)
		if err != nil {
			return fmt.Errorf("failed to create migration file for %s: %w", tableName, err)
		}

		// Register migration in runtime registry so it can be applied immediately
		RegisterMigration(&Migration{
			Version: version,
			Name:    migrationName,
			Up:      upSQL,
			Down:    downSQL,
		})

		log.Printf("✅ Generated migration for table: %s (version=%s)", tableName, version)
	}

	return nil
}

// Generate migration for a single model
func (g *MigrationGenerator) GenerateForModel(model interface{}) error {
	tableName := getTableName(model)
	migrationName := fmt.Sprintf("create_%s_table", tableName)

	// Generate migration SQL
	upSQL, downSQL, err := g.generateTableSQL(model)
	if err != nil {
		return err
	}

	// Create migration file and register
	version, err := g.createMigrationFile(migrationName, upSQL, downSQL)
	if err != nil {
		return err
	}
	RegisterMigration(&Migration{
		Version: version,
		Name:    migrationName,
		Up:      upSQL,
		Down:    downSQL,
	})
	return nil
}

// Generate SQL for creating table
func (g *MigrationGenerator) generateTableSQL(model interface{}) (upSQL, downSQL string, err error) {
	tableName := getTableName(model)

	// Get table definition from GORM
	stmt := &gorm.Statement{DB: g.db}
	if err := stmt.Parse(model); err != nil {
		return "", "", err
	}

	// Build CREATE TABLE SQL
	var columns []string
	var primaryKeys []string

	// Get model type and collect all exported fields, including embedded ones
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := collectFields(t)
	for _, field := range fields {
		columnSQL := g.generateColumnSQL(field, stmt)
		if columnSQL != "" {
			columns = append(columns, columnSQL)
		}

		// Check for primary key
		if gormTag := field.Tag.Get("gorm"); gormTag != "" {
			if strings.Contains(gormTag, "primaryKey") {
				columnName := g.getColumnName(field, stmt)
				primaryKeys = append(primaryKeys, columnName)
			}
		}
	}

	// Build final SQL
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s", tableName, strings.Join(columns, ",\n  "))

	if len(primaryKeys) > 0 {
		createSQL += fmt.Sprintf(",\n  PRIMARY KEY (%s)", strings.Join(primaryKeys, ", "))
	}

	createSQL += "\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;"

	// Add indexes
	indexes := g.generateIndexSQL(model, stmt)
	if indexes != "" {
		createSQL += "\n\n" + indexes
	}

	upSQL = createSQL
	downSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)

	return upSQL, downSQL, nil
}

// Generate column SQL definition
func (g *MigrationGenerator) generateColumnSQL(field reflect.StructField, stmt *gorm.Statement) string {
	columnName := g.getColumnName(field, stmt)
	if columnName == "" {
		return ""
	}

	gormTag := field.Tag.Get("gorm")
	fieldType := field.Type

	// Skip relations and ignored fields
	if strings.Contains(gormTag, "-") || strings.Contains(gormTag, "foreignKey") || strings.Contains(gormTag, "references") {
		return ""
	}

	// Determine SQL type
	sqlType := g.mapGoTypeToSQL(fieldType, gormTag)
	if sqlType == "" {
		return ""
	}

	// Build column definition
	var parts []string
	parts = append(parts, fmt.Sprintf("%s %s", columnName, sqlType))

	// Handle primary key
	isPrimaryKey := strings.Contains(gormTag, "primaryKey")

	// Add NOT NULL
	if strings.Contains(gormTag, "not null") || strings.Contains(gormTag, "NOT NULL") || isPrimaryKey {
		parts = append(parts, "NOT NULL")
	} else {
		parts = append(parts, "NULL")
	}

	// Add AUTO_INCREMENT (harus setelah NOT NULL)
	if strings.Contains(gormTag, "autoIncrement") || strings.Contains(gormTag, "AUTO_INCREMENT") {
		parts = append(parts, "AUTO_INCREMENT")
	}

	// Add DEFAULT value
	if defaultVal := g.extractDefaultValue(gormTag); defaultVal != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", defaultVal))
	}

	// Add COMMENT
	if comment := g.extractComment(gormTag); comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", comment))
	}

	return strings.Join(parts, " ")
}

// Get column name from field
func (g *MigrationGenerator) getColumnName(field reflect.StructField, stmt *gorm.Statement) string {
	gormTag := field.Tag.Get("gorm")

	// Extract column name from gorm tag
	if gormTag != "" {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "column:") {
				columnName := strings.TrimPrefix(part, "column:")
				// Validasi nama column
				if columnName != "" {
					return columnName
				}
			}
		}
	}

	name := field.Name

	return toSnakeCase(name)
}

// Map Go type to SQL type
func (g *MigrationGenerator) mapGoTypeToSQL(fieldType reflect.Type, gormTag string) string {
	// Check if type is specified in gorm tag
	if strings.Contains(gormTag, "type:") {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "type:") {
				sqlType := strings.TrimPrefix(part, "type:")
				// Handle size specification
				if strings.Contains(gormTag, "size:") {
					sizeParts := strings.Split(gormTag, ";")
					for _, sizePart := range sizeParts {
						if strings.HasPrefix(sizePart, "size:") {
							size := strings.TrimPrefix(sizePart, "size:")
							sqlType = strings.Replace(sqlType, ")", ","+size+")", 1)
							break
						}
					}
				}
				return sqlType
			}
		}
	}

	// Map based on Go type
	switch fieldType.Kind() {
	case reflect.String:
		// Check for size in gorm tag
		size := "255"
		if strings.Contains(gormTag, "size:") {
			parts := strings.Split(gormTag, ";")
			for _, part := range parts {
				if strings.HasPrefix(part, "size:") {
					size = strings.TrimPrefix(part, "size:")
					break
				}
			}
		}
		return fmt.Sprintf("VARCHAR(%s)", size)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "INT"

	case reflect.Int64, reflect.Uint64:
		return "BIGINT"

	case reflect.Bool:
		return "TINYINT(1)"

	case reflect.Float32:
		return "FLOAT"

	case reflect.Float64:
		return "DOUBLE"

	case reflect.Struct:
		// time.Time -> DATETIME
		if fieldType.String() == "time.Time" {
			return "DATETIME"
		}
		// gorm.DeletedAt type should be mapped to DATETIME so it can be indexed
		// (gorm.DeletedAt is a struct type in the gorm package)
		if strings.HasSuffix(fieldType.String(), ".DeletedAt") || fieldType.String() == "gorm.DeletedAt" {
			return "DATETIME"
		}
	case reflect.Ptr:
		// Handle pointer types: map common pointer-to-primitive types to SQL
		// (e.g. *uint -> INT, *int64 -> BIGINT, *time.Time -> DATETIME).
		elem := fieldType.Elem()
		switch elem.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			return "INT"
		case reflect.Int64, reflect.Uint64:
			return "BIGINT"
		case reflect.Struct:
			if elem.String() == "time.Time" {
				return "DATETIME"
			}
		}
	}

	return "TEXT"
}

// Extract default value from gorm tag
func (g *MigrationGenerator) extractDefaultValue(gormTag string) string {
	if strings.Contains(gormTag, "default:") {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "default:") {
				value := strings.TrimPrefix(part, "default:")
				// Handle special cases
				if value == "CURRENT_TIMESTAMP" {
					return value
				}
				return "'" + value + "'"
			}
		}
	}
	return ""
}

// Extract comment from gorm tag
func (g *MigrationGenerator) extractComment(gormTag string) string {
	if strings.Contains(gormTag, "comment:") {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "comment:") {
				return strings.TrimPrefix(part, "comment:")
			}
		}
	}
	return ""
}

// Generate index SQL
func (g *MigrationGenerator) generateIndexSQL(model interface{}, stmt *gorm.Statement) string {
	var indexes []string
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	tableName := getTableName(model)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		gormTag := field.Tag.Get("gorm")
		columnName := g.getColumnName(field, stmt)

		if strings.Contains(gormTag, "uniqueIndex") {
			// Extract index name from gorm tag
			indexName := ""
			parts := strings.Split(gormTag, ";")
			for _, part := range parts {
				if strings.HasPrefix(part, "uniqueIndex:") {
					indexName = strings.TrimPrefix(part, "uniqueIndex:")
					break
				}
			}
			if indexName == "" {
				indexName = fmt.Sprintf("uidx_%s_%s", tableName, columnName)
			}
			indexes = append(indexes, fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s);",
				indexName, tableName, columnName))

		} else if strings.Contains(gormTag, "index") {
			// Extract index name from gorm tag
			indexName := ""
			parts := strings.Split(gormTag, ";")
			for _, part := range parts {
				if strings.HasPrefix(part, "index:") {
					indexName = strings.TrimPrefix(part, "index:")
					break
				}
			}
			if indexName == "" {
				indexName = fmt.Sprintf("idx_%s_%s", tableName, columnName)
			}
			indexes = append(indexes, fmt.Sprintf("CREATE INDEX %s ON %s (%s);",
				indexName, tableName, columnName))
		}
	}

	return strings.Join(indexes, "\n")
}

// Check if migration already exists
func (g *MigrationGenerator) migrationExists(migrationName string) bool {
	pattern := filepath.Join(g.migrationsDir, "*_"+migrationName+".go")
	matches, _ := filepath.Glob(pattern)
	return len(matches) > 0
}

// Create migration file
func (g *MigrationGenerator) createMigrationFile(migrationName, upSQL, downSQL string) (string, error) {
	version := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.go", version, migrationName)
	filePath := filepath.Join(g.migrationsDir, filename)

	// Create migration file template
	tmpl := `package migrations

import (
	"study1/internal/core/database"
)

func init() {
	database.RegisterMigration(&database.Migration{
		Version: "{{.Version}}",
		Name:    "{{.Name}}",
		Up: ` + "`" + `{{.UpSQL}}` + "`" + `,
		Down: ` + "`" + `{{.DownSQL}}` + "`" + `,
	})
}
`

	data := struct {
		Version string
		Name    string
		UpSQL   string
		DownSQL string
	}{
		Version: version,
		Name:    migrationName,
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	template := template.Must(template.New("migration").Parse(tmpl))
	if err := template.Execute(file, data); err != nil {
		return "", err
	}

	// Also write plain SQL files so runtime can apply migrations even if
	// generated Go files are not compiled into the running binary.
	upPath := filepath.Join(g.migrationsDir, fmt.Sprintf("%s_%s.up.sql", version, migrationName))
	downPath := filepath.Join(g.migrationsDir, fmt.Sprintf("%s_%s.down.sql", version, migrationName))

	if err := os.WriteFile(upPath, []byte(upSQL), 0644); err != nil {
		return "", err
	}
	if err := os.WriteFile(downPath, []byte(downSQL), 0644); err != nil {
		return "", err
	}

	return version, nil
}

// collectFields returns all exported struct fields for the provided type,
// recursively descending into anonymous (embedded) struct fields so that
// embedded fields are presented as a flat list. Unexported fields are skipped.
func collectFields(t reflect.Type) []reflect.StructField {
	var out []reflect.StructField
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// skip unexported
		if f.PkgPath != "" {
			continue
		}
		if f.Anonymous {
			// embedded struct -> recurse
			ft := f.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				inner := collectFields(ft)
				out = append(out, inner...)
				continue
			}
		}
		out = append(out, f)
	}
	return out
}
