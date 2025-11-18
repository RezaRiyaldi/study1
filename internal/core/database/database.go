package database

import (
	"database/sql"
	"fmt"
	"study1/internal/core/config"
	"study1/internal/core/types"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(cfg config.DatabaseConfig) (*DB, error) {
	// Ensure the database exists (create if missing). This is idempotent.
	if err := ensureDatabase(cfg); err != nil {
		return nil, fmt.Errorf("ensure database: %w", err)
	}

	dsn := cfg.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &DB{DB: db}, nil
}

// ensureDatabase connects to the MySQL server without selecting a database
// and creates the configured database if it does not exist. This mirrors
// behavior in frameworks that offer to create the DB automatically.
func ensureDatabase(cfg config.DatabaseConfig) error {
	dsn := cfg.GetDSNNoDB()
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	// Try pinging to ensure connection
	if err := sqlDB.Ping(); err != nil {
		return err
	}

	// Create database if it does not exist
	createStmt := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", cfg.Name)
	if _, err := sqlDB.Exec(createStmt); err != nil {
		return fmt.Errorf("create database %s: %w", cfg.Name, err)
	}

	return nil
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
