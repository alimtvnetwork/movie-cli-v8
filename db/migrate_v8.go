// migrate_v8.go — V8: create Task table for task tracking, quarantine history, and safe undo.
package db

import (
	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

func migrateV8(d *DB) error {
	createTaskTableSQL := `
	CREATE TABLE IF NOT EXISTS Task (
		TaskId         INTEGER PRIMARY KEY AUTOINCREMENT,
		TaskType       TEXT NOT NULL,
		MediaId        INTEGER,
		SourcePath     TEXT NOT NULL,
		QuarantinePath TEXT,
		TargetPath     TEXT,
		Status         TEXT NOT NULL CHECK(Status IN ('pending', 'quarantined', 'purged', 'undone', 'failed')) DEFAULT 'pending',
		BatchId        TEXT,
		ItemTitle      TEXT,
		Payload        TEXT,
		ErrorMessage   TEXT,
		CreatedAt      TEXT NOT NULL DEFAULT (datetime('now')),
		UpdatedAt      TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (MediaId) REFERENCES Media(MediaId) ON DELETE SET NULL
	);
	CREATE INDEX IF NOT EXISTS Idx_Task_Status ON Task(Status);
	CREATE INDEX IF NOT EXISTS Idx_Task_BatchId ON Task(BatchId);
	CREATE INDEX IF NOT EXISTS Idx_Task_MediaId ON Task(MediaId);
	`

	if _, err := d.Exec(createTaskTableSQL); err != nil {
		return appfault.Wrap("create Task table", err)
	}

	return nil
}
