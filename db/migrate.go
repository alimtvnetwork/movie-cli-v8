// migrate.go — Versioned migration runner with SchemaVersion tracking.
package db

import (
	"fmt"
	"time"

	"github.com/alimtvnetwork/movie-cli-v7/apperror"
)

// Migration represents a single versioned schema migration.
type Migration struct {
	Apply       func(d *DB) error
	Description string
	Version     int
}

// allMigrations returns all registered migrations in order.
func allMigrations() []Migration {
	return []Migration{
		{Version: 1, Description: "Initial PascalCase schema", Apply: migrateV1},
		{Version: 2, Description: "ImdbLookupCache table for cached DuckDuckGo→IMDb lookups", Apply: migrateV2},
		{Version: 3, Description: "ImdbLookupCache: add TmdbId + MediaType to skip /find on hit", Apply: migrateV3},
		{Version: 4, Description: "MediaStatus lookup + Media.IsDeleted/MediaStatusId for soft-delete", Apply: migrateV4},
		{Version: 5, Description: "ReconciliationActionType + ReconciliationHistory for SmartRescan", Apply: migrateV5},
	}
}

// ensureSchemaVersionTable creates the SchemaVersion tracking table.
func (d *DB) ensureSchemaVersionTable() error {
	_, err := d.Exec(`
	CREATE TABLE IF NOT EXISTS SchemaVersion (
		SchemaVersionId INTEGER PRIMARY KEY AUTOINCREMENT,
		Version         INTEGER NOT NULL UNIQUE,
		Description     TEXT NOT NULL,
		AppliedAt       TEXT NOT NULL DEFAULT (datetime('now'))
	);
	`)
	return err
}

// currentSchemaVersion returns the highest applied migration version.
func (d *DB) currentSchemaVersion() (int, error) {
	var ver int
	err := d.QueryRow("SELECT COALESCE(MAX(Version), 0) FROM SchemaVersion").Scan(&ver)
	return ver, err
}

// recordMigration stamps a migration as applied.
func (d *DB) recordMigration(m Migration) error {
	_, err := d.Exec(
		"INSERT INTO SchemaVersion (Version, Description, AppliedAt) VALUES (?, ?, ?)",
		m.Version, m.Description, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// runMigrations applies all pending migrations in order.
func (d *DB) runMigrations() error {
	if err := d.ensureSchemaVersionTable(); err != nil {
		return apperror.Wrap("create SchemaVersion table", err)
	}

	current, err := d.currentSchemaVersion()
	if err != nil {
		return apperror.Wrap("read schema version", err)
	}

	for _, m := range allMigrations() {
		if m.Version <= current {
			continue
		}
		if err := m.Apply(d); err != nil {
			return apperror.Wrapf(err, "migration v%d (%s)", m.Version, m.Description)
		}
		if err := d.recordMigration(m); err != nil {
			return apperror.Wrapf(err, "record migration v%d", m.Version)
		}
		fmt.Printf("  ✅ Migration v%d: %s\n", m.Version, m.Description)
	}

	return nil
}
