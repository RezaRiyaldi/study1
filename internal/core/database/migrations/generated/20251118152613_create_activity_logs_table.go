package migrations

import (
	"study1/internal/core/database"
)

func init() {
	database.RegisterMigration(&database.Migration{
		Version: "20251118152613",
		Name:    "create_activity_logs_table",
		Up: `CREATE TABLE IF NOT EXISTS activity_logs (
  id INT NOT NULL AUTO_INCREMENT,
  uuid VARCHAR(36) NOT NULL,
  method VARCHAR(16) NULL,
  path VARCHAR(1024) NULL,
  status INT NULL,
  latency_ms BIGINT NULL,
  ip VARCHAR(64) NULL,
  user_agent VARCHAR(512) NULL,
  user_id INT NULL,
  created_at DATETIME NULL,
  created_by INT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
		Down: `DROP TABLE IF EXISTS activity_logs;`,
	})
}
