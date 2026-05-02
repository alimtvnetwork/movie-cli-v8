// movie_history_log.go — shared, structured log lines for undo/redo so
// troubleshooting logs always carry the same fields in the same order:
//
//	[undo] kind=Move        move_id=42  before=/dst/x.mkv after=/src/x.mkv
//	[undo] kind=Delete      action_id=7 before=<deleted>  after=<restored>
//	[redo] kind=Rename      move_id=42  before=/src/x.mkv after=/dst/x.mkv
//	[redo] kind=ScanAdd     action_id=9 before=<absent>   after=<inserted>
//
// The `before`/`after` columns describe filesystem state from the user's
// point of view — for moves they are real paths; for action-history
// entries that have no path (Delete, ScanAdd, etc.) they are sentinel
// tokens so the column count stays constant and grep/awk pipelines work.
package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// Sentinels used when an action does not have real before/after paths.
// Kept as constants so callers and log scrapers share one source of truth.
const (
	pathAbsent    = "<absent>"
	pathDeleted   = "<deleted>"
	pathRestored  = "<restored>"
	pathInserted  = "<inserted>"
	pathUnchanged = "<unchanged>"
)

// LogUndoTarget prints the structured target line for a move undo.
// before = current on-disk location (ToPath), after = where it goes (FromPath).
func LogUndoMoveTarget(m *db.MoveRecord) {
	fmt.Printf("[undo] kind=%-12s move_id=%d   before=%s after=%s\n",
		moveKindLabel(m), m.ID, m.ToPath, m.FromPath)
}

// LogRedoMoveTarget mirrors LogUndoMoveTarget for the redo direction.
func LogRedoMoveTarget(m *db.MoveRecord) {
	fmt.Printf("[redo] kind=%-12s move_id=%d   before=%s after=%s\n",
		moveKindLabel(m), m.ID, m.FromPath, m.ToPath)
}

// LogUndoActionTarget prints the structured target line for an action undo.
func LogUndoActionTarget(a *db.ActionRecord) {
	before, after := actionPaths(a, true)
	fmt.Printf("[undo] kind=%-12s action_id=%d before=%s after=%s\n",
		a.FileActionId, a.ActionHistoryId, before, after)
}

// LogRedoActionTarget prints the structured target line for an action redo.
func LogRedoActionTarget(a *db.ActionRecord) {
	before, after := actionPaths(a, false)
	fmt.Printf("[redo] kind=%-12s action_id=%d before=%s after=%s\n",
		a.FileActionId, a.ActionHistoryId, before, after)
}

// moveKindLabel resolves the FileActionId on a MoveRecord (stored as a
// raw int) into the same display label ActionRecord uses, so undo/redo
// logs read identically across both history tables.
func moveKindLabel(m *db.MoveRecord) string {
	return db.FileActionType(m.FileActionId).String()
}

// actionPaths returns (before, after) sentinels describing the
// filesystem-visible effect of undoing (forUndo=true) or redoing
// (forUndo=false) the given action. Sentinels are used because most
// action types operate on DB rows, not files.
func actionPaths(a *db.ActionRecord, forUndo bool) (string, string) {
	switch a.FileActionId {
	case db.FileActionDelete, db.FileActionScanRemove:
		if forUndo {
			return pathDeleted, pathRestored
		}
		return pathRestored, pathDeleted
	case db.FileActionScanAdd:
		if forUndo {
			return pathInserted, pathAbsent
		}
		return pathAbsent, pathInserted
	case db.FileActionRestore:
		if forUndo {
			return pathRestored, pathDeleted
		}
		return pathDeleted, pathRestored
	}
	return pathUnchanged, pathUnchanged
}
