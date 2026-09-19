// migrate_v7.go — V7: create StagedAction table and add BackdropPath to Media.
package db

import (
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

func migrateV7(d *DB) error {
	createStagedTableSQL := `
	CREATE TABLE IF NOT EXISTS StagedAction (
		StagedActionId   INTEGER PRIMARY KEY AUTOINCREMENT,
		ActionType       TEXT NOT NULL CHECK(ActionType IN ('delete', 'move', 'rename')),
		MediaId          INTEGER,
		SourcePath       TEXT NOT NULL,
		DestinationPath  TEXT,
		MediaSnapshot    TEXT,
		Payload          TEXT,
		Status           TEXT NOT NULL CHECK(Status IN ('pending', 'applied', 'discarded', 'failed')) DEFAULT 'pending',
		BatchId          TEXT,
		ErrorMessage     TEXT,
		CreatedAt        TEXT NOT NULL DEFAULT (datetime('now')),
		UpdatedAt        TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (MediaId) REFERENCES Media(MediaId) ON DELETE SET NULL
	);
	CREATE INDEX IF NOT EXISTS Idx_StagedAction_Status ON StagedAction(Status);
	CREATE INDEX IF NOT EXISTS Idx_StagedAction_MediaId ON StagedAction(MediaId);
	`

	if _, err := d.Exec(createStagedTableSQL); err != nil {
		return appfault.Wrap("create StagedAction table", err)
	}

	hasBackdrop := false
	rows, err := d.Query("PRAGMA table_info(Media)")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var dfltValue interface{}
			if scanErr := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); scanErr == nil {
				if strings.EqualFold(name, "BackdropPath") {
					hasBackdrop = true
					break
				}
			}
		}
	}

	if !hasBackdrop {
		if _, alterErr := d.Exec("ALTER TABLE Media ADD COLUMN BackdropPath TEXT"); alterErr != nil {
			return appfault.Wrap("add BackdropPath column to Media", alterErr)
		}
	}

	return nil
}
