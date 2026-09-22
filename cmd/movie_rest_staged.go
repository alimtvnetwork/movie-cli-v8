// movie_rest_staged.go — REST API handlers for staged actions and batch application.
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
	"github.com/alimtvnetwork/movie-cli-v8/pkg/trashbin"
)

type stagedCreateRequest struct {
	ActionType      string `json:"action_type"`
	SourcePath      string `json:"source_path,omitempty"`
	DestinationPath string `json:"destination_path,omitempty"`
	MediaID         int64  `json:"media_id"`
	DeleteFolder    bool   `json:"delete_folder,omitempty"`
}

func handleStagedList(database *db.DB, w http.ResponseWriter, r *http.Request) {
	records, err := database.ListPendingStagedActions()
	if err != nil {
		writeRestError(w, http.StatusInternalServerError, "STAGED_QUERY_FAILED", err.Error())
		return
	}

	if records == nil {
		records = []db.StagedActionRecord{}
	}

	writeRestJSON(w, http.StatusOK, records)
}

func handleStagedCreate(database *db.DB, w http.ResponseWriter, r *http.Request) {
	var req stagedCreateRequest
	if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
		writeRestError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "malformed JSON payload")
		return
	}

	actionType := db.StagedActionType(strings.ToLower(req.ActionType))
	if actionType != db.StagedDelete && actionType != db.StagedMove && actionType != db.StagedRename {
		writeRestError(w, http.StatusBadRequest, "INVALID_ACTION_TYPE", "action_type must be delete, move, or rename")
		return
	}

	var sourcePath string
	var snapshot string
	var nullMediaID sql.NullInt64

	if req.MediaID > 0 {
		nullMediaID = sql.NullInt64{Int64: req.MediaID, Valid: true}
		media, fetchErr := database.GetMediaByID(req.MediaID)
		if fetchErr == nil {
			if media != nil {
				sourcePath = media.CurrentFilePath
				if req.DeleteFolder {
					if media.CurrentFilePath != "" {
						sourcePath = filepath.Dir(media.CurrentFilePath)
					}
				}

				snapBytes, _ := json.Marshal(media)
				snapshot = string(snapBytes)
			}
		}
	}

	if req.SourcePath != "" {
		sourcePath = req.SourcePath
	}

	rec := &db.StagedActionRecord{
		ActionType:      actionType,
		MediaId:         nullMediaID,
		SourcePath:      sourcePath,
		DestinationPath: req.DestinationPath,
		MediaSnapshot:   snapshot,
		Status:          db.StatusPending,
	}

	insertedID, insErr := database.InsertStagedAction(rec)
	if insErr != nil {
		writeRestError(w, http.StatusInternalServerError, "STAGED_INSERT_FAILED", insErr.Error())
		return
	}

	rec.StagedActionId = insertedID
	writeRestJSON(w, http.StatusCreated, rec)
}

func handleStagedDiscard(database *db.DB, w http.ResponseWriter, id int64) {
	if err := database.DiscardStagedAction(id); err != nil {
		writeRestError(w, http.StatusInternalServerError, "DISCARD_FAILED", err.Error())
		return
	}

	writeRestJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"id":      id,
		"status":  db.StatusDiscarded,
	})
}

func handleStagedDiscardAll(database *db.DB, w http.ResponseWriter) {
	if err := database.DiscardAllPendingStagedActions(); err != nil {
		writeRestError(w, http.StatusInternalServerError, "DISCARD_ALL_FAILED", err.Error())
		return
	}

	writeRestJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"status":  db.StatusDiscarded,
	})
}

func handleStagedApplyAll(database *db.DB, w http.ResponseWriter) {
	records, err := database.ListPendingStagedActions()
	if err != nil {
		writeRestError(w, http.StatusInternalServerError, "QUERY_FAILED", err.Error())
		return
	}

	batchID := fmt.Sprintf("batch_%s", time.Now().Format("20060102_150405"))
	appliedCount := 0
	failedCount := 0
	results := make([]map[string]interface{}, 0, len(records))

	for i := range records {
		rec := &records[i]
		applyErr := applySingleStagedRecord(database, rec, batchID)
		if applyErr != nil {
			failedCount++
			_ = database.UpdateStagedStatus(rec.StagedActionId, db.StatusFailed, applyErr.Error())
			results = append(results, map[string]interface{}{
				"id":     rec.StagedActionId,
				"status": "failed",
				"error":  applyErr.Error(),
			})
		} else {
			appliedCount++
			_ = database.UpdateStagedStatus(rec.StagedActionId, db.StatusApplied, "")
			results = append(results, map[string]interface{}{
				"id":     rec.StagedActionId,
				"status": "applied",
			})
		}
	}

	writeRestJSON(w, http.StatusOK, map[string]interface{}{
		"batch_id": batchID,
		"applied":  appliedCount,
		"failed":   failedCount,
		"results":  results,
	})
}

func handleStagedApplySingle(database *db.DB, w http.ResponseWriter, id int64) {
	rec, err := database.GetStagedActionByID(id)
	if err != nil {
		writeRestError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("staged action #%d not found", id))
		return
	}

	batchID := fmt.Sprintf("single_%d_%s", id, time.Now().Format("150405"))
	if applyErr := applySingleStagedRecord(database, rec, batchID); applyErr != nil {
		_ = database.UpdateStagedStatus(id, db.StatusFailed, applyErr.Error())
		writeRestError(w, http.StatusInternalServerError, "APPLY_FAILED", applyErr.Error())
		return
	}

	_ = database.UpdateStagedStatus(id, db.StatusApplied, "")
	writeRestJSON(w, http.StatusOK, map[string]interface{}{
		"id":       id,
		"status":   "applied",
		"batch_id": batchID,
	})
}

func applySingleStagedRecord(database *db.DB, rec *db.StagedActionRecord, batchID string) error {
	switch rec.ActionType {
	case db.StagedDelete:
		if rec.SourcePath != "" {
			if _, statErr := os.Stat(rec.SourcePath); statErr == nil {
				if trashErr := trashbin.MoveToTrash(rec.SourcePath); trashErr != nil {
					return appfault.Wrapf(trashErr, "move %s to trash", rec.SourcePath)
				}
			}
		}

		if rec.MediaId.Valid {
			if softErr := database.SoftDeleteMedia(rec.MediaId.Int64); softErr != nil {
				return appfault.Wrapf(softErr, "soft delete media #%d", rec.MediaId.Int64)
			}
			_, _ = database.InsertActionSimple(db.ActionSimpleInput{
				FileAction: db.FileActionDelete,
				MediaID:    rec.MediaId.Int64,
				Snapshot:   rec.MediaSnapshot,
				Detail:     "moved to trash via staged apply",
				BatchID:    batchID,
			})
		}

		return nil

	case db.StagedMove, db.StagedRename:
		if rec.SourcePath == "" || rec.DestinationPath == "" {
			return appfault.New("source and destination path required for move/rename")
		}

		destDir := filepath.Dir(rec.DestinationPath)
		if mkErr := os.MkdirAll(destDir, 0755); mkErr != nil {
			return appfault.Wrapf(mkErr, "create destination dir %s", destDir)
		}

		if renErr := os.Rename(rec.SourcePath, rec.DestinationPath); renErr != nil {
			return appfault.Wrapf(renErr, "move file %s -> %s", rec.SourcePath, rec.DestinationPath)
		}

		if rec.MediaId.Valid {
			if updErr := database.UpdateMediaPath(rec.MediaId.Int64, rec.DestinationPath); updErr != nil {
				return appfault.Wrapf(updErr, "update media #%d path", rec.MediaId.Int64)
			}

			_, _ = database.InsertActionSimple(db.ActionSimpleInput{
				FileAction: db.FileActionMove,
				MediaID:    rec.MediaId.Int64,
				Snapshot:   rec.MediaSnapshot,
				Detail:     fmt.Sprintf("moved to %s", rec.DestinationPath),
				BatchID:    batchID,
			})
		}

		return nil

	default:
		return appfault.New("unknown staged action type: %s", rec.ActionType)
	}
}

func writeRestJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func writeRestError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	env := appfault.NewErrorEnvelope(statusCode, code, message, "")
	_ = json.NewEncoder(w).Encode(env)
}
