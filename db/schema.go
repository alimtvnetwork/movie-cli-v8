// schema.go — Schema orchestration via versioned migrations.
package db

import (
	"github.com/alimtvnetwork/movie-cli-v8/apperror"
	"github.com/alimtvnetwork/movie-cli-v8/version"
)

// migrateSchema runs all pending migrations and stamps the app version.
func (d *DB) migrateSchema() error {
	if err := d.runMigrations(); err != nil {
		return apperror.Wrap("run migrations", err)
	}
	if err := d.SetConfig("AppVersion", version.Short()); err != nil {
		return apperror.Wrap("stamp app version", err)
	}
	return nil
}

// createTables creates all PascalCase tables.
func (d *DB) createTables() error {
	if err := d.createLookupTables(); err != nil {
		return err
	}
	if err := d.createCoreTables(); err != nil {
		return err
	}
	if err := d.createJoinTables(); err != nil {
		return err
	}
	if err := d.createHistoryTables(); err != nil {
		return err
	}
	if err := d.createSystemTables(); err != nil {
		return err
	}
	return d.createIndexes()
}
