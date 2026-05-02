// movie_undo.go — movie undo: reverts the last state-changing operation.
//
// Supports undoing:
//   - File moves/renames  (from move_history)
//   - Deletions           (from action_history)
//   - Scan additions      (from action_history)
//   - Scan removals       (from action_history)
//   - Rescan updates      (from action_history)
//
// Flags:
//
//	--list           Show recent undoable actions
//	--id <id>        Undo a specific action_history record
//	--batch          Undo entire last batch
//	--move-id <id>   Undo a specific move_history record
package cmd

import (
	"bufio"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var (
	undoListFlag  bool
	undoBatchFlag bool
	undoGlobal    bool
	undoAssumeYes bool
	undoActionID  int64
	undoMoveID    int64
	undoIncludes  []string
	undoExcludes  []string
)

var movieUndoCmd = &cobra.Command{
	Use:   "undo [path]",
	Short: "Undo the last operation (move, rename, delete, scan)",
	Long: `Reverts the most recent state-changing operation.

Scope:
  By default, only history rooted under the current working directory
  (or the optional [path] argument) is considered. Pass --global to
  undo across the entire database like in older versions.

Flags:
  --list           Show recent undoable actions
  --id <id>        Undo a specific action_history record by ID
  --move-id <id>   Undo a specific move_history record by ID
  --batch          Undo the entire last batch (e.g. a full scan)
  --global         Ignore the cwd / [path] scope
  --yes, -y        Skip every confirmation prompt (scripted runs)
  --include <glob> Keep only actions whose paths match this glob (repeatable)
  --exclude <glob> Drop actions whose paths match this glob (repeatable)`,
	Run: runMovieUndo,
}

func init() {
	movieUndoCmd.Flags().BoolVar(&undoListFlag, "list", false, "Show recent undoable actions")
	movieUndoCmd.Flags().BoolVar(&undoBatchFlag, "batch", false, "Undo entire last batch")
	movieUndoCmd.Flags().BoolVar(&undoGlobal, "global", false, "Ignore cwd / path scope")
	movieUndoCmd.Flags().BoolVarP(&undoAssumeYes, "yes", "y", false, "Skip all confirmation prompts (also: --assume-yes)")
	movieUndoCmd.Flags().BoolVar(&undoAssumeYes, "assume-yes", false, "Alias for --yes")
	movieUndoCmd.Flags().Int64Var(&undoActionID, "id", 0, "Undo specific action_history record")
	movieUndoCmd.Flags().Int64Var(&undoMoveID, "move-id", 0, "Undo specific move_history record")
	movieUndoCmd.Flags().StringSliceVar(&undoIncludes, "include", nil, "Glob pattern to keep (repeatable)")
	movieUndoCmd.Flags().StringSliceVar(&undoExcludes, "exclude", nil, "Glob pattern to drop (repeatable)")
}

func runMovieUndo(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		os.Exit(ExitGenericError)
	}
	defer database.Close()

	scanner := bufio.NewScanner(os.Stdin)
	home, _ := os.UserHomeDir()
	filter := buildScopeFilter(args, home, undoGlobal, undoIncludes, undoExcludes, undoAssumeYes)

	exitWithCode(dispatchUndo(database, scanner, filter))
}

// dispatchUndo routes to the right handler and returns its exit code.
func dispatchUndo(database *db.DB, scanner *bufio.Scanner, filter ScopeFilter) int {
	switch {
	case undoListFlag:
		return showUndoableList(database, filter)
	case undoActionID > 0:
		return undoActionByID(database, scanner, undoActionID)
	case undoMoveID > 0:
		return undoMoveByID(database, scanner, undoMoveID)
	case undoBatchFlag:
		return undoLastBatch(database, scanner, filter)
	}
	return undoLastOperation(database, scanner, filter)
}
