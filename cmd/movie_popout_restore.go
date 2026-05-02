// movie_popout_restore.go — undo/redo handlers for FileActionCompact entries.
//
// When `movie popout` compacts a non-media folder it records a row with
// FileActionCompact and a JSON snapshot of the form:
//
//	{"original_path":"/abs/path/Folder","compact_path":"/abs/path/.temp/Folder"}
//
// This file owns the symmetric filesystem operations:
//
//   - undoCompact: move <compact_path> back to <original_path>
//   - redoCompact: move <original_path> forward into <compact_path>
//
// Both update IsReverted on the action row so list/redo/undo flows stay
// consistent. See mem://features/popout-spec.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// compactSnapshot is the on-disk JSON shape emitted by compactFolder().
// Kept package-private; only used by undo/redo here.
type compactSnapshot struct {
	OriginalPath string `json:"original_path"`
	CompactPath  string `json:"compact_path"`
}

// parseCompactSnapshot deserializes the snapshot JSON for a FileActionCompact
// row. Returns an apperror with action ID context on parse failures.
func parseCompactSnapshot(a *db.ActionRecord) (compactSnapshot, error) {
	var snap compactSnapshot
	if a.MediaSnapshot == "" {
		return snap, apperror.New("no snapshot for compact action %d", a.ActionHistoryId)
	}
	if err := json.Unmarshal([]byte(a.MediaSnapshot), &snap); err != nil {
		return snap, apperror.Wrapf(err, "parse compact snapshot %d", a.ActionHistoryId)
	}
	if snap.OriginalPath == "" || snap.CompactPath == "" {
		return snap, apperror.New("incomplete compact snapshot for action %d", a.ActionHistoryId)
	}
	return snap, nil
}

// undoCompact moves a compacted folder back from .temp/ to its original path.
func undoCompact(database *db.DB, a *db.ActionRecord) error {
	snap, err := parseCompactSnapshot(a)
	if err != nil {
		return err
	}
	if statErr := requireDirExists(snap.CompactPath); statErr != nil {
		return statErr
	}
	if _, existsErr := os.Stat(snap.OriginalPath); existsErr == nil {
		return apperror.New("cannot restore: %s already exists", snap.OriginalPath)
	}
	if mkErr := os.MkdirAll(filepath.Dir(snap.OriginalPath), 0755); mkErr != nil {
		return apperror.Wrapf(mkErr, "create parent for %s", snap.OriginalPath)
	}
	if mvErr := MoveFile(snap.CompactPath, snap.OriginalPath); mvErr != nil {
		return apperror.Wrap("restore compacted folder", mvErr)
	}
	fmt.Printf("   📦↩  Restored: .temp/%s  →  %s\n",
		filepath.Base(snap.CompactPath), snap.OriginalPath)
	return database.MarkActionReverted(a.ActionHistoryId)
}

// redoCompact re-applies a previously undone compaction by moving the folder
// back into <root>/.temp/.
func redoCompact(database *db.DB, a *db.ActionRecord) error {
	snap, err := parseCompactSnapshot(a)
	if err != nil {
		return err
	}
	if statErr := requireDirExists(snap.OriginalPath); statErr != nil {
		return statErr
	}
	if mkErr := os.MkdirAll(filepath.Dir(snap.CompactPath), 0755); mkErr != nil {
		return apperror.Wrapf(mkErr, "create parent for %s", snap.CompactPath)
	}
	if mvErr := MoveFile(snap.OriginalPath, snap.CompactPath); mvErr != nil {
		return apperror.Wrap("re-compact folder", mvErr)
	}
	fmt.Printf("   📦   Re-compacted: %s  →  .temp/%s\n",
		snap.OriginalPath, filepath.Base(snap.CompactPath))
	return database.MarkActionRestored(a.ActionHistoryId)
}

// requireDirExists returns an apperror when the path is missing or not a dir.
func requireDirExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return apperror.New("path not found: %s", path)
		}
		return apperror.Wrapf(err, "cannot access %s", path)
	}
	if !info.IsDir() {
		return apperror.New("not a directory: %s", path)
	}
	return nil
}
