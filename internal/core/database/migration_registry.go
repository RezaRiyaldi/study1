package database

import "sort"

// Migration represents a database migration
type Migration struct {
	Version string
	Name    string
	Up      string
	Down    string
}

var migrations []*Migration

// RegisterMigration registers a migration
func RegisterMigration(migration *Migration) {
	migrations = append(migrations, migration)
}

// GetMigrations returns all registered migrations sorted by version
func GetMigrations() []*Migration {
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations
}

// GetMigrationByVersion returns a migration by version
func GetMigrationByVersion(version string) *Migration {
	for _, migration := range migrations {
		if migration.Version == version {
			return migration
		}
	}
	return nil
}
