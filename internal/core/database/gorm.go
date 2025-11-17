package database

import (
	"reflect"
	"strings"
	"study1/internal/core/types"
	"unicode"

	"gorm.io/gorm"
)

type QueryBuilder[T any] struct {
	DB     *gorm.DB
	Model  T
	Params types.QueryParams
}

func (b *QueryBuilder[T]) WithParams(params types.QueryParams) *QueryBuilder[T] {
	b.Params = params
	return b
}

func (b *QueryBuilder[T]) Build() *gorm.DB {
	b.applySearch()
	b.applyFilters()
	b.applySorting()
	b.applyPagination()
	b.applyRelation()
	b.applyFieldSelection()
	return b.DB
}

func (b *QueryBuilder[T]) applySearch() *QueryBuilder[T] {
	if b.Params.Search == "" {
		return b
	}

	searchableFields := b.detectSearchableFields()

	if len(searchableFields) > 0 {
		query := b.DB

		for i, field := range searchableFields {
			if i == 0 {
				// MySQL menggunakan LIKE
				query = query.Where(field+" LIKE ?", "%"+b.Params.Search+"%")
			} else {
				query = query.Or(field+" LIKE ?", "%"+b.Params.Search+"%")
			}
		}

		b.DB = query
	}

	return b
}

func (b *QueryBuilder[T]) applyFilters() *QueryBuilder[T] {
	if b.Params.Filter == nil {
		return b
	}

	for field, value := range b.Params.Filter {
		b.DB = b.DB.Where(field+" = ?", value)
	}

	return b
}

func (b *QueryBuilder[T]) applySorting() *QueryBuilder[T] {
	if b.Params.Sort == "" {
		b.DB = b.DB.Order("created_at DESC")
		return b
	}

	b.DB = b.DB.Order(b.Params.Sort)
	return b
}

func (b *QueryBuilder[T]) applyPagination() *QueryBuilder[T] {
	if b.Params.Page == 0 {
		b.Params.Page = 1
	}
	if b.Params.PageSize == 0 {
		b.Params.PageSize = 10
	}

	offset := (b.Params.Page - 1) * b.Params.PageSize
	b.DB = b.DB.Offset(offset).Limit(b.Params.PageSize)
	return b
}

func (b *QueryBuilder[T]) applyRelation() *QueryBuilder[T] {
	if b.Params.Include == "" {
		return b
	}

	relations := strings.Split(b.Params.Include, ",")
	for _, relation := range relations {
		b.DB = b.DB.Preload(strings.TrimSpace(relation))
	}

	return b
}

func (b *QueryBuilder[T]) applyFieldSelection() *QueryBuilder[T] {
	if b.Params.Fields == "" {
		return b
	}

	fields := strings.Split(b.Params.Fields, ",")
	b.DB = b.DB.Select(fields)
	return b
}

// toSnakeCase mengubah string CamelCase menjadi snake_case
func toSnakeCase(s string) string {
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

func (b *QueryBuilder[T]) detectSearchableFields() []string {
	var fields []string

	t := reflect.TypeOf(b.Model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if search := field.Tag.Get("searchable"); search == "true" {
			// Prioritaskan gorm tag untuk column name
			if gormTag := field.Tag.Get("gorm"); gormTag != "" {
				// Extract column name dari gorm tag
				parts := strings.Split(gormTag, ";")
				for _, part := range parts {
					if strings.HasPrefix(part, "column:") {
						columnName := strings.TrimPrefix(part, "column:")
						fields = append(fields, columnName)
						break
					}
				}
			} else if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				// Fallback ke json tag
				fields = append(fields, strings.Split(jsonTag, ",")[0])
			} else {
				// Fallback terakhir: convert field name ke snake_case
				fields = append(fields, toSnakeCase(field.Name))
			}
		}
	}

	return fields
}
