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
	"unicode"

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

		// Create migration file
		if err := g.createMigrationFile(migrationName, upSQL, downSQL); err != nil {
			return fmt.Errorf("failed to create migration file for %s: %w", tableName, err)
		}

		log.Printf("✅ Generated migration for table: %s", tableName)
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

	// Create migration file
	return g.createMigrationFile(migrationName, upSQL, downSQL)
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

	// Get model type
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
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
	createSQL := fmt.Sprintf("CREATE TABLE %s (\n  %s", tableName, strings.Join(columns, ",\n  "))

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

	// Fallback to field name in snake_case
	name := field.Name
	// Handle common abbreviations
	if name == "ID" {
		return "id"
	}
	return g.toSnakeCase(name)
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
		if fieldType.String() == "time.Time" {
			return "DATETIME"
		}
	case reflect.Ptr:
		// Handle pointer types (for soft delete)
		if fieldType.String() == "*time.Time" {
			return "DATETIME"
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
func (g *MigrationGenerator) createMigrationFile(migrationName, upSQL, downSQL string) error {
	version := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.go", version, migrationName)
	filepath := filepath.Join(g.migrationsDir, filename)

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

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	template := template.Must(template.New("migration").Parse(tmpl))
	return template.Execute(file, data)
}

// toSnakeCase mengubah string CamelCase menjadi snake_case
func (g *MigrationGenerator) toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
