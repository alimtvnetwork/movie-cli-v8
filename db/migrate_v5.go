// migrate_v5.go — V5: adds the ReconciliationActionType lookup and the
// ReconciliationHistory audit table used by SmartRescan to record every
// hydrate / missing-removal / new-discovery / converged-skip event.
//
// Spec: spec/08-app/10-remove-move-rescan/rescan-reconciliation/01-spec.md
//
// New artefacts:
//   - ReconciliationActionType  (lookup, 4 seed rows)
//   - ReconciliationHistory     (audit rows; MediaId nullable for AddedNew)
package db

import "github.com/alimtvnetwork/movie-cli-v7/apperror"

// ReconciliationActionTypeId enum values, kept in sync with the seed order.
const (
	ReconActionHydratedFromJson = 1
	ReconActionRemovedMissing   = 2
	ReconActionAddedNew         = 3
	ReconActionConverged        = 4
)

func migrateV5(d *DB) error {
	if err := createReconActionTypeTable(d); err != nil {
		return apperror.Wrap("create ReconciliationActionType", err)
	}
	if err := seedReconActionType(d); err != nil {
		return apperror.Wrap("seed ReconciliationActionType", err)
	}
	return createReconHistoryTable(d)
}

func createReconActionTypeTable(d *DB) error {
	_, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS ReconciliationActionType (
			ReconciliationActionTypeId INTEGER PRIMARY KEY AUTOINCREMENT,
			Name                       TEXT    NOT NULL UNIQUE
		)`)
	return err
}

func seedReconActionType(d *DB) error {
	names := []string{"HydratedFromJson", "RemovedMissing", "AddedNew", "Converged"}
	for _, name := range names {
		if _, err := d.Exec(
			"INSERT OR IGNORE INTO ReconciliationActionType (Name) VALUES (?)", name,
		); err != nil {
			return apperror.Wrapf(err, "seed ReconciliationActionType %q", name)
		}
	}
	return nil
}

func createReconHistoryTable(d *DB) error {
	_, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS ReconciliationHistory (
			ReconciliationHistoryId    INTEGER PRIMARY KEY AUTOINCREMENT,
			MediaId                    INTEGER,
			ReconciliationActionTypeId INTEGER NOT NULL,
			Details                    TEXT    NOT NULL DEFAULT '',
			OccurredAt                 TEXT    NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (MediaId)                    REFERENCES Media(MediaId),
			FOREIGN KEY (ReconciliationActionTypeId) REFERENCES ReconciliationActionType(ReconciliationActionTypeId)
		)`)
	return err
}
