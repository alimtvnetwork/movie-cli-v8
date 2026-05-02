// migrate_v4.go — V4: introduces MediaStatus lookup and adds two columns to
// Media to support the `movie rm` / `movie remove` / `movie delete` family.
//
// Spec: spec/08-app/10-remove-move-rescan/00-overview.md
//
// New artifacts:
//   - MediaStatus            (lookup, 4 seed rows: Active|Removed|Moved|Missing)
//   - Media.IsDeleted        (BOOL; soft-delete flag, default 0 = false)
//   - Media.MediaStatusId    (INT FK → MediaStatus, default 1 = Active)
//
// SQLite ALTER TABLE only supports ADD COLUMN — both new columns use safe
// defaults so existing rows remain valid Active media.
package db

import "github.com/alimtvnetwork/movie-cli-v8/apperror"

// MediaStatusId enum values, kept in sync with the seed order below.
const (
	MediaStatusActive  = 1
	MediaStatusRemoved = 2
	MediaStatusMoved   = 3
	MediaStatusMissing = 4
)

func migrateV4(d *DB) error {
	if err := createMediaStatusTable(d); err != nil {
		return apperror.Wrap("create MediaStatus", err)
	}
	if err := seedMediaStatus(d); err != nil {
		return apperror.Wrap("seed MediaStatus", err)
	}
	return addMediaStatusColumns(d)
}

func createMediaStatusTable(d *DB) error {
	_, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS MediaStatus (
			MediaStatusId INTEGER PRIMARY KEY AUTOINCREMENT,
			Name          TEXT    NOT NULL UNIQUE
		)`)
	return err
}

func seedMediaStatus(d *DB) error {
	names := []string{"Active", "Removed", "Moved", "Missing"}
	for _, name := range names {
		if _, err := d.Exec(
			"INSERT OR IGNORE INTO MediaStatus (Name) VALUES (?)", name,
		); err != nil {
			return apperror.Wrapf(err, "seed MediaStatus %q", name)
		}
	}
	return nil
}

func addMediaStatusColumns(d *DB) error {
	stmts := []string{
		`ALTER TABLE Media ADD COLUMN IsDeleted     INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE Media ADD COLUMN MediaStatusId INTEGER NOT NULL DEFAULT 1`,
	}
	for _, stmt := range stmts {
		if _, err := d.Exec(stmt); err != nil {
			return apperror.Wrap("alter Media", err)
		}
	}
	return nil
}
