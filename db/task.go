// task.go — SQLite task tracking, quarantine history, and audit log.
package db

import (
	"database/sql"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

// TaskStatus values.
const (
	TaskPending     = "pending"
	TaskQuarantined = "quarantined"
	TaskPurged      = "purged"
	TaskUndone      = "undone"
	TaskFailed      = "failed"
)

// TaskType values.
const (
	TaskTypeDeleteStage    = "delete_stage"
	TaskTypeQuarantineMove = "quarantine_move"
	TaskTypePurge          = "purge"
	TaskTypeUndo           = "undo"
	TaskTypeMove           = "move"
	TaskTypeRename         = "rename"
)

// TaskRecord models a row in the Task table.
type TaskRecord struct {
	TaskType       string        `json:"task_type"`
	SourcePath     string        `json:"source_path"`
	QuarantinePath string        `json:"quarantine_path,omitempty"`
	TargetPath     string        `json:"target_path,omitempty"`
	Status         string        `json:"status"`
	BatchId        string        `json:"batch_id,omitempty"`
	ItemTitle      string        `json:"item_title,omitempty"`
	Payload        string        `json:"payload,omitempty"`
	ErrorMessage   string        `json:"error_message,omitempty"`
	CreatedAt      string        `json:"created_at"`
	UpdatedAt      string        `json:"updated_at"`
	MediaId        sql.NullInt64 `json:"media_id"`
	TaskId         int64         `json:"task_id"`
}

// InsertTask inserts a new task record into the database.
func (d *DB) InsertTask(t *TaskRecord) (int64, error) {
	if t == nil {
		return 0, appfault.New("task record cannot be nil")
	}

	var mediaID interface{}

	if t.MediaId.Valid {
		mediaID = t.MediaId.Int64
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := d.Exec(`
		INSERT INTO Task (
			TaskType, MediaId, SourcePath, QuarantinePath, TargetPath,
			Status, BatchId, ItemTitle, Payload, ErrorMessage, CreatedAt, UpdatedAt
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.TaskType, mediaID, t.SourcePath, t.QuarantinePath, t.TargetPath,
		t.Status, t.BatchId, t.ItemTitle, t.Payload, t.ErrorMessage, now, now,
	)

	if err != nil {
		return 0, appfault.Wrap("insert task record", err)
	}

	return res.LastInsertId()
}

// UpdateTaskStatus updates the status and error message of a task.
func (d *DB) UpdateTaskStatus(taskID int64, status, errorMsg string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.Exec(`
		UPDATE Task
		SET Status = ?, ErrorMessage = ?, UpdatedAt = ?
		WHERE TaskId = ?`,
		status, errorMsg, now, taskID,
	)

	if err != nil {
		return appfault.Wrapf(err, "update task #%d status to %s", taskID, status)
	}

	return nil
}

// ListPendingQuarantineTasks returns all tasks currently in quarantined status.
func (d *DB) ListPendingQuarantineTasks() ([]TaskRecord, error) {
	rows, err := d.Query(`
		SELECT TaskId, TaskType, MediaId, SourcePath, QuarantinePath, TargetPath,
		       Status, BatchId, ItemTitle, Payload, ErrorMessage, CreatedAt, UpdatedAt
		FROM Task
		WHERE Status = ?
		ORDER BY TaskId ASC`, TaskQuarantined)

	if err != nil {
		return nil, appfault.Wrap("list quarantined tasks", err)
	}

	defer rows.Close()

	var records []TaskRecord

	for rows.Next() {
		var rec TaskRecord

		scanErr := rows.Scan(
			&rec.TaskId, &rec.TaskType, &rec.MediaId, &rec.SourcePath, &rec.QuarantinePath, &rec.TargetPath,
			&rec.Status, &rec.BatchId, &rec.ItemTitle, &rec.Payload, &rec.ErrorMessage, &rec.CreatedAt, &rec.UpdatedAt,
		)

		if scanErr != nil {
			return nil, appfault.Wrap("scan quarantined task", scanErr)
		}

		records = append(records, rec)
	}

	return records, nil
}

// GetTaskByID retrieves a single task by its primary key.
func (d *DB) GetTaskByID(taskID int64) (*TaskRecord, error) {
	row := d.QueryRow(`
		SELECT TaskId, TaskType, MediaId, SourcePath, QuarantinePath, TargetPath,
		       Status, BatchId, ItemTitle, Payload, ErrorMessage, CreatedAt, UpdatedAt
		FROM Task
		WHERE TaskId = ?`, taskID)

	var rec TaskRecord
	err := row.Scan(
		&rec.TaskId, &rec.TaskType, &rec.MediaId, &rec.SourcePath, &rec.QuarantinePath, &rec.TargetPath,
		&rec.Status, &rec.BatchId, &rec.ItemTitle, &rec.Payload, &rec.ErrorMessage, &rec.CreatedAt, &rec.UpdatedAt,
	)

	if err != nil {
		return nil, appfault.Wrapf(err, "get task #%d", taskID)
	}

	return &rec, nil
}

// GetQuarantinedTaskByMediaID returns the active quarantined task for a media item.
func (d *DB) GetQuarantinedTaskByMediaID(mediaID int64) (*TaskRecord, error) {
	row := d.QueryRow(`
		SELECT TaskId, TaskType, MediaId, SourcePath, QuarantinePath, TargetPath,
		       Status, BatchId, ItemTitle, Payload, ErrorMessage, CreatedAt, UpdatedAt
		FROM Task
		WHERE MediaId = ? AND Status = ?
		ORDER BY TaskId DESC LIMIT 1`, mediaID, TaskQuarantined)

	var rec TaskRecord
	err := row.Scan(
		&rec.TaskId, &rec.TaskType, &rec.MediaId, &rec.SourcePath, &rec.QuarantinePath, &rec.TargetPath,
		&rec.Status, &rec.BatchId, &rec.ItemTitle, &rec.Payload, &rec.ErrorMessage, &rec.CreatedAt, &rec.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &rec, nil
}

// ListTasks returns recent task history records.
func (d *DB) ListTasks(limit int) ([]TaskRecord, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := d.Query(`
		SELECT TaskId, TaskType, MediaId, SourcePath, QuarantinePath, TargetPath,
		       Status, BatchId, ItemTitle, Payload, ErrorMessage, CreatedAt, UpdatedAt
		FROM Task
		ORDER BY TaskId DESC LIMIT ?`, limit)

	if err != nil {
		return nil, appfault.Wrap("list task history", err)
	}

	defer rows.Close()

	var records []TaskRecord

	for rows.Next() {
		var rec TaskRecord

		scanErr := rows.Scan(
			&rec.TaskId, &rec.TaskType, &rec.MediaId, &rec.SourcePath, &rec.QuarantinePath, &rec.TargetPath,
			&rec.Status, &rec.BatchId, &rec.ItemTitle, &rec.Payload, &rec.ErrorMessage, &rec.CreatedAt, &rec.UpdatedAt,
		)

		if scanErr != nil {
			return nil, appfault.Wrap("scan task row", scanErr)
		}

		records = append(records, rec)
	}

	return records, nil
}
