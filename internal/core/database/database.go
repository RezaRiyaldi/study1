package database

import (
	"study1/internal/core/config"
	"study1/internal/core/types"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(cfg config.DatabaseConfig) (*DB, error) {
	dsn := cfg.GetDSNMySQL()

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}
	return &DB{DB: db}, nil
}

// NewQueryBuilder creates a new QueryBuilder for the specified model type T.
func NewQueryBuilder[T any](db *gorm.DB, params types.QueryParams) *QueryBuilder[T] {
	var model T
	return &QueryBuilder[T]{
		DB:     db.Model(&model),
		Model:  model,
		Params: params,
	}
}
