package migrations

import (
	"study1/internal/core/database"
)

func init() {
	database.RegisterMigration(&database.Migration{
		Version: "20251117195835",
		Name:    "create_users_table",
		Up: `CREATE TABLE users (
  id INT NOT NULL AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(100) NOT NULL,
  age int NULL DEFAULT '0',
  created_at DATETIME NULL,
  updated_at DATETIME NULL,
  deleted_at TEXT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE UNIQUE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);`,
		Down: `DROP TABLE IF EXISTS users;`,
	})
}
