// migrate_v6.go — V6: seed 4 reverse-sync action types into
// ReconciliationActionType (idempotent INSERT OR IGNORE).
//
// Spec: spec/08-app/10-remove-move-rescan/rescan-reconciliation/02-reverse-sync-spec.md
package db

import "github.com/alimtvnetwork/movie-cli-v7/apperror"

func migrateV6(d *DB) error {
	names := []string{
		"ReverseSyncedSidecar", "RemovedOrphanSidecar",
		"RemovedDeletedSidecar", "ReverseDetectedMissing",
	}
	for _, name := range names {
		if _, err := d.Exec(
			"INSERT OR IGNORE INTO ReconciliationActionType (Name) VALUES (?)", name,
		); err != nil {
			return apperror.Wrapf(err, "seed ReconciliationActionType %q", name)
		}
	}
	return nil
}
