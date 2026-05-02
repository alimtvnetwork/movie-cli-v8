// movie_move_atomic.go — atomic batched move with rollback on partial failure.
//
// executeBatchMovesAtomic moves files one-by-one. On the first failure it
// rolls back every already-completed move (file back to its source path,
// plus pruning any destination directories we created in this batch) and
// aborts before any DB writes happen for the failed run.
//
// The DB tracking via trackMove only runs when ALL files succeed, so DB
// state remains consistent — no partial MoveHistory rows on rollback.
package cmd

import (
	"fmt"
	"os"

	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
)

// moveAtomicOff lets the user opt out of atomic semantics.
var moveAtomicOff bool

// completedMove remembers a successful filesystem move so it can be undone.
type completedMove struct {
	srcPath    string
	destPath   string
	destDir    string
	createdDir bool
}

// executeBatchMovesAtomic performs the batch with all-or-nothing semantics.
func executeBatchMovesAtomic(database *db.DB, moves []moveItem) {
	completed := make([]completedMove, 0, len(moves))
	for i := range moves {
		ok, created := performOneMove(&moves[i])
		if !ok {
			rollbackMoves(completed)
			fmt.Printf("\n  ⛔ Aborted at file %d/%d. Rolled back %d move(s).\n",
				i+1, len(moves), len(completed))
			return
		}
		completed = append(completed, completedMove{
			srcPath: moves[i].srcPath, destPath: moves[i].destPath,
			destDir: moves[i].destDir, createdDir: created,
		})
	}
	persistBatchMoves(database, moves)
	fmt.Printf("\n  ✅ All %d files moved successfully (atomic).\n", len(moves))
	regenerateReports(database)
}

// performOneMove creates the dest dir (if needed) and moves one file.
// Returns (ok, createdDir) — createdDir is true when we made the dir.
func performOneMove(m *moveItem) (bool, bool) {
	createdDir := false
	if _, statErr := os.Stat(m.destDir); os.IsNotExist(statErr) {
		createdDir = true
	}
	if err := os.MkdirAll(m.destDir, 0755); err != nil {
		errlog.Error("Cannot create dir %s: %v", m.destDir, err)
		return false, false
	}
	if err := MoveFile(m.srcPath, m.destPath); err != nil {
		errlog.Error("Failed to move %s: %v", m.fileInfo.Name(), err)
		return false, createdDir
	}
	return true, createdDir
}

// rollbackMoves reverses every successful move in reverse order.
func rollbackMoves(completed []completedMove) {
	for i := len(completed) - 1; i >= 0; i-- {
		c := completed[i]
		if err := MoveFile(c.destPath, c.srcPath); err != nil {
			errlog.Warn("Rollback failed for %s → %s: %v", c.destPath, c.srcPath, err)
			continue
		}
		if c.createdDir {
			_ = os.Remove(c.destDir)
		}
	}
}

// persistBatchMoves writes all DB tracking rows after a successful batch.
func persistBatchMoves(database *db.DB, moves []moveItem) {
	for i := range moves {
		trackMove(TrackMoveInput{
			Database: database, Result: moves[i].result, FileInfo: moves[i].fileInfo,
			SrcPath: moves[i].srcPath, DestPath: moves[i].destPath, CleanName: moves[i].cleanName,
		})
	}
}
