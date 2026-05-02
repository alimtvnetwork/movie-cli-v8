// movie_undo_handlers.go — undo subcommand handlers (list, by-id, batch, last).
package cmd

import (
	"bufio"
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

// History fetch limits — used by BOTH preview (--list) and execution
// flows so they apply scope/glob filters to the same row set. Changing
// these in one place changes the entire undo/redo experience uniformly.
const (
	undoMoveScanLimit   = 200
	undoActionScanLimit = 200
)

func showUndoableList(database *db.DB, f ScopeFilter) int {
	fmt.Println("⏪ Recent undoable operations")
	printScopeBanner(f)
	fmt.Println()

	moveSkipped := countUndoableMoveSkipped(database, f)
	actionSkipped := countUndoableActionSkipped(database, f)

	undoableMoves := printUndoableMoves(database, f)
	undoableActions := printUndoableActions(database, f)

	if undoableMoves == 0 && undoableActions == 0 {
		fmt.Println("  📭 Nothing to undo in this scope.")
	}

	printPreviewSummary(PreviewSummary{
		Verb:           "Undo",
		MatchedMoves:   undoableMoves,
		MatchedActions: undoableActions,
		SkippedMoves:   moveSkipped,
		SkippedActions: actionSkipped,
	})
	if undoableMoves+undoableActions == 0 {
		return ExitNothingMatched
	}
	return ExitOK
}

// countUndoableMoveSkipped returns how many non-reverted moves were
// dropped by the current filter (for list-mode summary).
func countUndoableMoveSkipped(database *db.DB, f ScopeFilter) int {
	raw, _ := database.ListMoveHistory(undoMoveScanLimit)
	kept := FilterMovesWith(raw, f)
	return countScopeSkipped(countUndoableMoves(raw), countUndoableMoves(kept))
}

func countUndoableActionSkipped(database *db.DB, f ScopeFilter) int {
	raw, _ := database.ListActions(undoActionScanLimit)
	kept := FilterActionsWith(raw, f)
	return countScopeSkipped(countNonReverted(raw), countNonReverted(kept))
}

func countUndoableMoves(moves []db.MoveRecord) int {
	count := 0
	for _, m := range moves {
		if !m.IsReverted {
			count++
		}
	}
	return count
}

func printUndoableMoves(database *db.DB, f ScopeFilter) int {
	rawMoves, _ := database.ListMoveHistory(undoMoveScanLimit)
	moves := FilterMovesWith(rawMoves, f)
	count := 0
	for _, m := range moves {
		if !m.IsReverted {
			count++
		}
	}
	if count > 0 {
		fmt.Printf("  📁 Moves / Renames  — %d ready to undo:\n", count)
		for _, m := range moves {
			if m.IsReverted {
				continue
			}
			fmt.Printf("    [move-%d]  %s → %s  (%s)\n", m.ID, m.FromPath, m.ToPath, m.MovedAt)
		}
		fmt.Println()
	}
	return count
}

func printUndoableActions(database *db.DB, f ScopeFilter) int {
	rawActions, _ := database.ListActions(undoActionScanLimit)
	actions := FilterActionsWith(rawActions, f)
	count := countNonReverted(actions)
	if count == 0 {
		return 0
	}
	fmt.Printf("  📋 Actions  — %d ready to undo:\n", count)
	for _, a := range actions {
		if a.IsReverted {
			continue
		}
		fmt.Printf("    [action-%d]  %s  %s  (%s%s)\n",
			a.ActionHistoryId, a.FileActionId, actionDetail(a), a.CreatedAt, batchSuffix(a.BatchId))
	}
	fmt.Println()
	return count
}

func countNonReverted(actions []db.ActionRecord) int {
	count := 0
	for _, a := range actions {
		if !a.IsReverted {
			count++
		}
	}
	return count
}

func actionDetail(a db.ActionRecord) string {
	if a.Detail != "" {
		return a.Detail
	}
	return a.FileActionId.String()
}

func batchSuffix(batchID string) string {
	if batchID == "" {
		return ""
	}
	short := batchID
	if len(short) > 8 {
		short = short[:8]
	}
	return fmt.Sprintf("  batch:%s", short)
}

func undoActionByID(database *db.DB, scanner *bufio.Scanner, id int64) int {
	action, err := database.GetActionByID(id)
	if err != nil {
		errlog.Error("Cannot find action %d: %v", id, err)
		return ExitGenericError
	}
	if action.IsReverted {
		fmt.Printf("⚠️  Action %d has already been reverted.\n", id)
		return ExitNothingMatched
	}

	fmt.Printf("⏪ Undo action %d (%s):\n", action.ActionHistoryId, action.FileActionId)
	if action.Detail != "" {
		fmt.Printf("   %s\n", action.Detail)
	}
	LogUndoActionTarget(action)
	if !confirmUndo(scanner) {
		return ExitRowDeclined
	}

	if err := executeActionUndo(database, action); err != nil {
		errlog.Error("Undo action %d failed: %v", id, err)
		return ExitGenericError
	}
	fmt.Printf("✅ Action %d reverted successfully.\n", action.ActionHistoryId)
	return ExitOK
}

func undoMoveByID(database *db.DB, scanner *bufio.Scanner, id int64) int {
	moves, err := database.ListMoveHistory(1000)
	if err != nil {
		errlog.Error("Cannot read move history: %v", err)
		return ExitGenericError
	}
	var target *db.MoveRecord
	for i := range moves {
		if moves[i].ID == id {
			target = &moves[i]
			break
		}
	}
	if target == nil {
		errlog.Error("Move %d not found.", id)
		return ExitGenericError
	}
	if target.IsReverted {
		fmt.Printf("⚠️  Move %d has already been reverted.\n", id)
		return ExitNothingMatched
	}

	fmt.Println("⏪ Undo move:")
	fmt.Printf("   %s → %s\n", target.ToPath, target.FromPath)
	LogUndoMoveTarget(target)
	if !confirmUndo(scanner) {
		return ExitRowDeclined
	}

	if err := executeMoveUndo(database, target); err != nil {
		errlog.Error("Undo move %d failed: %v", id, err)
		return ExitGenericError
	}
	fmt.Printf("✅ Move %d reverted successfully.\n", target.ID)
	return ExitOK
}

func undoLastBatch(database *db.DB, scanner *bufio.Scanner, f ScopeFilter) int {
	f, ok := ConfirmCwdScopeWithPreview(scanner, f, "Undo", undoableCountsFn(database))
	if !ok {
		return ExitScopeRejected
	}
	batchID := findLastUndoableBatch(database, f)
	if batchID == "" {
		fmt.Println("📭 No batch operations to undo in this scope.")
		return ExitNothingMatched
	}

	batchActions, err := database.ListActionsByBatch(batchID)
	if err != nil {
		errlog.Error("Cannot read batch %s: %v", batchID, err)
		return ExitGenericError
	}

	scoped := FilterActionsWith(batchActions, f)
	undoable := countUndoable(scoped)
	skipped := countScopeSkipped(countUndoable(batchActions), undoable)
	if undoable == 0 {
		fmt.Println("📭 Batch already reverted (or all actions filtered out).")
		printHistorySummary(HistorySummary{Verb: "Undo", Skipped: skipped})
		return ExitNothingMatched
	}

	fmt.Printf("⏪ Undo batch %s (%d actions, %d skipped by filter):\n",
		batchID[:8], undoable, skipped)
	printUndoableActionsList(scoped)
	if !confirmUndo(scanner) {
		return ExitRowDeclined
	}

	failed := executeUndoBatch(database, scoped)
	printUndoBatchResult(batchID[:8], undoable, failed)
	printHistorySummary(HistorySummary{
		Verb:     "Undo",
		Matched:  undoable,
		Executed: undoable - failed,
		Failed:   failed,
		Skipped:  skipped,
	})
	if failed > 0 && failed == undoable {
		return ExitGenericError
	}
	return ExitOK
}

func undoLastOperation(database *db.DB, scanner *bufio.Scanner, f ScopeFilter) int {
	f, ok := ConfirmCwdScopeWithPreview(scanner, f, "Undo", undoableCountsFn(database))
	if !ok {
		return ExitScopeRejected
	}
	lastMove := pickLastUndoableMove(database, f)
	lastAction := pickLastUndoableAction(database, f)
	skipped := countUndoableMoveSkipped(database, f) +
		countUndoableActionSkipped(database, f)

	haveMove := lastMove != nil
	haveAction := lastAction != nil

	if !haveMove && !haveAction {
		fmt.Println("📭 No operations to undo in this scope.")
		printHistorySummary(HistorySummary{Verb: "Undo", Skipped: skipped})
		return ExitNothingMatched
	}

	if haveMove && !haveAction {
		return runSingleUndoMove(database, scanner, lastMove, skipped)
	}

	if haveAction && !haveMove {
		return runSingleUndoAction(database, scanner, lastAction, skipped)
	}

	if lastAction.CreatedAt >= lastMove.MovedAt {
		return runSingleUndoAction(database, scanner, lastAction, skipped)
	}
	return runSingleUndoMove(database, scanner, lastMove, skipped)
}

// runSingleUndoMove wraps undoSingleMove with summary reporting.
// Returns the exit code so cobra Run can propagate it via os.Exit.
func runSingleUndoMove(database *db.DB, scanner *bufio.Scanner, m *db.MoveRecord, skipped int) int {
	code := undoSingleMoveCode(database, scanner, m)
	executed, failed := 0, 0
	switch code {
	case ExitOK:
		executed = 1
	case ExitGenericError:
		failed = 1
	}
	printHistorySummary(HistorySummary{
		Verb: "Undo", Matched: 1, Executed: executed, Failed: failed, Skipped: skipped,
	})
	return code
}

// runSingleUndoAction wraps undoSingleAction with summary reporting.
func runSingleUndoAction(database *db.DB, scanner *bufio.Scanner, a *db.ActionRecord, skipped int) int {
	code := undoSingleActionCode(database, scanner, a)
	executed, failed := 0, 0
	switch code {
	case ExitOK:
		executed = 1
	case ExitGenericError:
		failed = 1
	}
	printHistorySummary(HistorySummary{
		Verb: "Undo", Matched: 1, Executed: executed, Failed: failed, Skipped: skipped,
	})
	return code
}

// undoSingleMoveCode returns ExitOK on success, ExitRowDeclined on
// confirm decline, or ExitGenericError on filesystem failure.
func undoSingleMoveCode(database *db.DB, scanner *bufio.Scanner, m *db.MoveRecord) int {
	fmt.Println("⏪ Last move operation:")
	fmt.Printf("   %s → %s\n", m.ToPath, m.FromPath)
	LogUndoMoveTarget(m)
	if !confirmUndo(scanner) {
		return ExitRowDeclined
	}
	if err := executeMoveUndo(database, m); err != nil {
		errlog.Error("Undo failed: %v", err)
		return ExitGenericError
	}
	fmt.Println("✅ Undo successful!")
	return ExitOK
}

// undoSingleActionCode mirrors undoSingleMoveCode for actions.
func undoSingleActionCode(database *db.DB, scanner *bufio.Scanner, a *db.ActionRecord) int {
	printActionUndo(a)
	LogUndoActionTarget(a)
	if !confirmUndo(scanner) {
		return ExitRowDeclined
	}
	if err := executeActionUndo(database, a); err != nil {
		errlog.Error("Undo failed: %v", err)
		return ExitGenericError
	}
	fmt.Println("✅ Undo successful!")
	return ExitOK
}

// pickLastUndoableMove returns the newest non-reverted move under scope.
func pickLastUndoableMove(database *db.DB, f ScopeFilter) *db.MoveRecord {
	moves, err := database.ListMoveHistory(undoMoveScanLimit)
	if err != nil {
		return nil
	}
	// Use the canonical filter pipeline so preview and execution agree
	// on what "in scope" means.
	for _, m := range FilterMovesWith(moves, f) {
		if m.IsReverted {
			continue
		}
		return &m
	}
	return nil
}

// pickLastUndoableAction returns the newest non-reverted action under scope.
func pickLastUndoableAction(database *db.DB, f ScopeFilter) *db.ActionRecord {
	actions, err := database.ListActions(undoActionScanLimit)
	if err != nil {
		return nil
	}
	for _, a := range FilterActionsWith(actions, f) {
		if a.IsReverted {
			continue
		}
		return &a
	}
	return nil
}

func findLastUndoableBatch(database *db.DB, f ScopeFilter) string {
	actions, err := database.ListActions(200)
	if err != nil {
		errlog.Error("Cannot read action history: %v", err)
		return ""
	}
	for _, a := range actions {
		if a.IsReverted || a.BatchId == "" {
			continue
		}
		if !batchTouchesScope(database, a.BatchId, f) {
			continue
		}
		return a.BatchId
	}
	return ""
}

// batchTouchesScope returns true if any action in the batch passes the
// dir scope AND glob filter (when set). f.Dir == "" with no globs → true.
func batchTouchesScope(database *db.DB, batchID string, f ScopeFilter) bool {
	if f.Dir == "" && !f.HasGlobs() {
		return true
	}
	rows, err := database.ListActionsByBatch(batchID)
	if err != nil {
		return false
	}
	for _, a := range rows {
		if !ActionInScope(a, f.Dir) {
			continue
		}
		if !f.HasGlobs() || ActionMatchesGlobs(a, f) {
			return true
		}
	}
	return false
}

func countUndoable(actions []db.ActionRecord) int {
	count := 0
	for _, a := range actions {
		if !a.IsReverted {
			count++
		}
	}
	return count
}

func printUndoableActionsList(actions []db.ActionRecord) {
	for _, a := range actions {
		if a.IsReverted {
			continue
		}
		fmt.Printf("   • %s: %s\n", a.FileActionId, actionDetail(a))
	}
}

func executeUndoBatch(database *db.DB, actions []db.ActionRecord) int {
	failed := 0
	for i := len(actions) - 1; i >= 0; i-- {
		if actions[i].IsReverted {
			continue
		}
		if err := executeActionUndo(database, &actions[i]); err != nil {
			errlog.Warn("Failed to undo action %d: %v", actions[i].ActionHistoryId, err)
			failed++
		}
	}
	return failed
}

func printUndoBatchResult(shortBatch string, undoable, failed int) {
	if failed == 0 {
		fmt.Printf("✅ Batch %s reverted (%d actions).\n", shortBatch, undoable)
		return
	}
	fmt.Printf("⚠️  Batch %s: %d reverted, %d failed.\n", shortBatch, undoable-failed, failed)
}

// undoableCountsFn returns a ScopePreviewFn that, given a filter,
// reports how many non-reverted moves and actions are currently in
// scope. Used by the cwd-scope confirmation prompt so the user sees a
// live "would act on N moves, N actions" line before confirming.
func undoableCountsFn(database *db.DB) ScopePreviewFn {
	return func(f ScopeFilter) (int, int) {
		rawMoves, _ := database.ListMoveHistory(undoMoveScanLimit)
		rawActions, _ := database.ListActions(undoActionScanLimit)
		return countUndoableMoves(FilterMovesWith(rawMoves, f)),
			countNonReverted(FilterActionsWith(rawActions, f))
	}
}
