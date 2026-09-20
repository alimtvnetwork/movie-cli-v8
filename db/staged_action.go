// staged_action.go — CRUD operations and domain models for staged actions.
package db

import (
	"database/sql"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

// StagedActionType defines the nature of the staged operation.
type StagedActionType string

const (
	StagedDelete StagedActionType = "delete"
	StagedMove   StagedActionType = "move"
	StagedRename StagedActionType = "rename"
)

// StagedStatus defines the lifecycle state of a staged operation.
type StagedStatus string

const (
	StatusPending   StagedStatus = "pending"
	StatusApplied   StagedStatus = "applied"
	StatusDiscarded StagedStatus = "discarded"
	StatusFailed    StagedStatus = "failed"
)

// StagedActionRecord models a row in the StagedAction table.
type StagedActionRecord struct {
	ActionType      StagedActionType `json:"action_type"`
	SourcePath      string           `json:"source_path"`
	DestinationPath string           `json:"destination_path,omitempty"`
	MediaSnapshot   string           `json:"media_snapshot,omitempty"`
	Payload         string           `json:"payload,omitempty"`
	Status          StagedStatus     `json:"status"`
	BatchId         string           `json:"batch_id,omitempty"`
	ErrorMessage    string           `json:"error_message,omitempty"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
	StagedActionId  int64            `json:"staged_action_id"`
	MediaId         sql.NullInt64    `json:"media_id"`
}

// InsertStagedAction inserts a new staged action into the queue.
func (d *DB) InsertStagedAction(record *StagedActionRecord) (int64, error) {
	if record == nil {
		return 0, appfault.New("staged action record cannot be nil")
	}

	var mediaID interface{}
	if record.MediaId.Valid {
		mediaID = record.MediaId.Int64
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := d.Exec(`
		INSERT INTO StagedAction (
			ActionType, MediaId, SourcePath, DestinationPath,
			MediaSnapshot, Payload, Status, BatchId, ErrorMessage,
			CreatedAt, UpdatedAt
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ActionType, mediaID, record.SourcePath, record.DestinationPath,
		record.MediaSnapshot, record.Payload, StatusPending, record.BatchId,
		record.ErrorMessage, now, now,
	)
	if err != nil {
		return 0, appfault.Wrap("insert staged action", err)
	}

	return res.LastInsertId()
}

// ListPendingStagedActions returns all staged actions with status 'pending'.
func (d *DB) ListPendingStagedActions() ([]StagedActionRecord, error) {
	rows, err := d.Query(`
		SELECT StagedActionId, ActionType, MediaId, SourcePath,
		       COALESCE(DestinationPath, ''), COALESCE(MediaSnapshot, ''),
		       COALESCE(Payload, ''), Status, COALESCE(BatchId, ''),
		       COALESCE(ErrorMessage, ''), CreatedAt, UpdatedAt
		FROM StagedAction
		WHERE Status = ?
		ORDER BY StagedActionId ASC`,
		StatusPending,
	)
	if err != nil {
		return nil, appfault.Wrap("query pending staged actions", err)
	}
	defer rows.Close()

	var records []StagedActionRecord
	for rows.Next() {
		var rec StagedActionRecord
		if scanErr := rows.Scan(
			&rec.StagedActionId, &rec.ActionType, &rec.MediaId, &rec.SourcePath,
			&rec.DestinationPath, &rec.MediaSnapshot, &rec.Payload, &rec.Status,
			&rec.BatchId, &rec.ErrorMessage, &rec.CreatedAt, &rec.UpdatedAt,
		); scanErr != nil {
			return nil, appfault.Wrap("scan staged action row", scanErr)
		}

		records = append(records, rec)
	}

	return records, rows.Err()
}

// GetStagedActionByID retrieves a single staged action by its primary key.
func (d *DB) GetStagedActionByID(id int64) (*StagedActionRecord, error) {
	var rec StagedActionRecord
	row := d.QueryRow(`
		SELECT StagedActionId, ActionType, MediaId, SourcePath,
		       COALESCE(DestinationPath, ''), COALESCE(MediaSnapshot, ''),
		       COALESCE(Payload, ''), Status, COALESCE(BatchId, ''),
		       COALESCE(ErrorMessage, ''), CreatedAt, UpdatedAt
		FROM StagedAction
		WHERE StagedActionId = ?`,
		id,
	)

	if err := row.Scan(
		&rec.StagedActionId, &rec.ActionType, &rec.MediaId, &rec.SourcePath,
		&rec.DestinationPath, &rec.MediaSnapshot, &rec.Payload, &rec.Status,
		&rec.BatchId, &rec.ErrorMessage, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return nil, appfault.Wrapf(err, "get staged action #%d", id)
	}

	return &rec, nil
}

// UpdateStagedStatus updates the lifecycle status of a staged action.
func (d *DB) UpdateStagedStatus(id int64, status StagedStatus, errMsg string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.Exec(`
		UPDATE StagedAction
		SET Status = ?, ErrorMessage = ?, UpdatedAt = ?
		WHERE StagedActionId = ?`,
		status, errMsg, now, id,
	)
	if err != nil {
		return appfault.Wrapf(err, "update staged action #%d status to %s", id, status)
	}

	return nil
}

// DiscardStagedAction marks a pending staged action as discarded.
func (d *DB) DiscardStagedAction(id int64) error {
	return d.UpdateStagedStatus(id, StatusDiscarded, "")
}

// DiscardAllPendingStagedActions marks all pending staged actions as discarded.
func (d *DB) DiscardAllPendingStagedActions() error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.Exec(`
		UPDATE StagedAction
		SET Status = ?, UpdatedAt = ?
		WHERE Status = ?`,
		StatusDiscarded, now, StatusPending,
	)
	if err != nil {
		return appfault.Wrap("discard all pending staged actions", err)
	}

	return nil
}
