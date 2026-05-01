// reconciliation_history.go — Insert + query helpers for the
// ReconciliationHistory audit table introduced in migrate_v5.
//
// SHARED: used by movie scan (SmartRescan reconciliation orchestrator) and
// by movie history (display). Do NOT INSERT into ReconciliationHistory
// directly from command files — always go through these helpers so the
// enum mapping and timestamp formatting stay consistent.
package db

import (
	"database/sql"

	"github.com/alimtvnetwork/movie-cli-v7/apperror"
)

// ReconRecord represents one ReconciliationHistory row (read side).
type ReconRecord struct {
	Details                    string
	OccurredAt                 string
	ReconciliationHistoryId    int64
	MediaId                    sql.NullInt64
	ReconciliationActionTypeId int
}

// ReconInput holds fields for inserting a reconciliation row.
type ReconInput struct {
	Details    string
	MediaID    sql.NullInt64
	ActionType int // one of ReconAction* constants
}

// InsertReconciliation appends one ReconciliationHistory row.
func (d *DB) InsertReconciliation(input ReconInput) (int64, error) {
	res, err := d.Exec(`
		INSERT INTO ReconciliationHistory
			(MediaId, ReconciliationActionTypeId, Details)
		VALUES (?, ?, ?)`,
		input.MediaID, input.ActionType, input.Details,
	)
	if err != nil {
		return 0, apperror.Wrap("insert ReconciliationHistory", err)
	}
	return res.LastInsertId()
}

// CountReconciliationByType returns counts grouped by action type for one
// SmartRescan run. Caller passes a since-timestamp (RFC3339) to scope it.
func (d *DB) CountReconciliationByType(since string) (map[int]int, error) {
	rows, err := d.Query(`
		SELECT ReconciliationActionTypeId, COUNT(*)
		FROM ReconciliationHistory
		WHERE OccurredAt >= ?
		GROUP BY ReconciliationActionTypeId`, since)
	if err != nil {
		return nil, apperror.Wrap("query ReconciliationHistory", err)
	}
	defer rows.Close()

	out := make(map[int]int)
	for rows.Next() {
		var actionType, count int
		if scanErr := rows.Scan(&actionType, &count); scanErr != nil {
			return nil, apperror.Wrap("scan ReconciliationHistory row", scanErr)
		}
		out[actionType] = count
	}
	return out, nil
}

// ListReconciliation returns the most recent ReconciliationHistory rows,
// joined with the action-type name. Limit caps the row count.
func (d *DB) ListReconciliation(limit int) ([]ReconRecord, error) {
	rows, err := d.Query(`
		SELECT rh.ReconciliationHistoryId, rh.MediaId,
		       rh.ReconciliationActionTypeId, rh.OccurredAt,
		       COALESCE(rh.Details, '')
		FROM   ReconciliationHistory rh
		ORDER  BY rh.OccurredAt DESC, rh.ReconciliationHistoryId DESC
		LIMIT  ?`, limit)
	if err != nil {
		return nil, apperror.Wrap("list ReconciliationHistory", err)
	}
	defer rows.Close()
	var out []ReconRecord
	for rows.Next() {
		var r ReconRecord
		if scanErr := rows.Scan(&r.ReconciliationHistoryId, &r.MediaId,
			&r.ReconciliationActionTypeId, &r.OccurredAt, &r.Details); scanErr != nil {
			return nil, apperror.Wrap("scan ReconciliationHistory row", scanErr)
		}
		out = append(out, r)
	}
	return out, nil
}

// ReconActionName maps a ReconciliationActionTypeId to its display name.
func ReconActionName(id int) string {
	switch id {
	case ReconActionHydratedFromJson:
		return "HydratedFromJson"
	case ReconActionRemovedMissing:
		return "RemovedMissing"
	case ReconActionAddedNew:
		return "AddedNew"
	case ReconActionConverged:
		return "Converged"
	}
	return "Unknown"
}
