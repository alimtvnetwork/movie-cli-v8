// movie_redo_handlers.go — redo subcommand handlers (list, by-id, batch, last).
package cmd

import (
	"bufio"
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

func showRedoableList(database *db.DB, f ScopeFilter) int {
	fmt.Println("⏩ Recent redoable operations")
	printScopeBanner(f)
	fmt.Println()

	moveSkipped := countRedoableMoveSkipped(database, f)
	actionSkipped := countRedoableActionSkipped(database, f)

	redoableMoves := printRedoableMoves(database, f)
	redoableActions := printRedoableActions(database, f)

	if redoableMoves == 0 && redoableActions == 0 {
		fmt.Println("  📭 Nothing to redo in this scope.")
	}

	printPreviewSummary(PreviewSummary{
		Verb:           "Redo",
		MatchedMoves:   redoableMoves,
		MatchedActions: redoableActions,
		SkippedMoves:   moveSkipped,
		SkippedActions: actionSkipped,
	})
	if redoableMoves+redoableActions == 0 {
		return ExitNothingMatched
	}
	return ExitOK
}

func countRedoableMoveSkipped(database *db.DB, f ScopeFilter) int {
	raw, _ := database.ListMoveHistory(undoMoveScanLimit)
	kept := FilterMovesWith(raw, f)
	return countScopeSkipped(countRevertedMoves(raw), countRevertedMoves(kept))
}

func countRedoableActionSkipped(database *db.DB, f ScopeFilter) int {
	raw, _ := database.ListActions(undoActionScanLimit)
	kept := FilterActionsWith(raw, f)
	return countScopeSkipped(countReverted(raw), countReverted(kept))
}

func countRevertedMoves(moves []db.MoveRecord) int {
	count := 0
	for _, m := range moves {
		if m.IsReverted {
			count++
		}
	}
	return count
}

func printRedoableMoves(database *db.DB, f ScopeFilter) int {
	rawMoves, _ := database.ListMoveHistory(undoMoveScanLimit)
	moves := FilterMovesWith(rawMoves, f)
	count := 0
	for _, m := range moves {
		if m.IsReverted {
			count++
		}
	}
	if count > 0 {
		fmt.Printf("  📁 Moves / Renames  — %d ready to redo:\n", count)
		for _, m := range moves {
			if !m.IsReverted {
				continue
			}
			fmt.Printf("    [move-%d]  %s → %s  (%s)\n", m.ID, m.FromPath, m.ToPath, m.MovedAt)
		}
		fmt.Println()
	}
	return count
}

func printRedoableActions(database *db.DB, f ScopeFilter) int {
	rawActions, _ := database.ListActions(undoActionScanLimit)
	actions := FilterActionsWith(rawActions, f)
	count := countReverted(actions)
	if count == 0 {
		return 0
	}
	fmt.Printf("  📋 Actions  — %d ready to redo:\n", count)
	for _, a := range actions {
		if !a.IsReverted {
			continue
		}
		fmt.Printf("    [action-%d]  %s  %s  (%s%s)\n",
			a.ActionHistoryId, a.FileActionId, actionDetail(a), a.CreatedAt, batchSuffix(a.BatchId))
	}
	fmt.Println()
	return count
}

func redoActionByID(database *db.DB, scanner *bufio.Scanner, id int64) int {
	action, err := database.GetActionByID(id)
	if err != nil {
		errlog.Error("Cannot find action %d: %v", id, err)
		return ExitGenericError
	}
	if !action.IsReverted {
		fmt.Printf("⚠️  Action %d is not reverted — nothing to redo.\n", id)
		return ExitNothingMatched
	}

	fmt.Printf("⏩ Redo action %d (%s):\n", action.ActionHistoryId, action.FileActionId)
	if action.Detail != "" {
		fmt.Printf("   %s\n", action.Detail)
	}
	LogRedoActionTarget(action)
	if !confirmRedo(scanner) {
		return ExitRowDeclined
	}

	if err := executeActionRedo(database, action); err != nil {
		errlog.Error("Redo action %d failed: %v", id, err)
		return ExitGenericError
	}
	fmt.Printf("✅ Action %d redone successfully.\n", action.ActionHistoryId)
	return ExitOK
}

func redoMoveByID(database *db.DB, scanner *bufio.Scanner, id int64) int {
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
	if !target.IsReverted {
		fmt.Printf("⚠️  Move %d is not reverted — nothing to redo.\n", id)
		return ExitNothingMatched
	}

	fmt.Println("⏩ Redo move:")
	fmt.Printf("   %s → %s\n", target.FromPath, target.ToPath)
	LogRedoMoveTarget(target)
	if !confirmRedo(scanner) {
		return ExitRowDeclined
	}

	if err := executeMoveRedo(database, target); err != nil {
		errlog.Error("Redo move %d failed: %v", id, err)
		return ExitGenericError
	}
	fmt.Printf("✅ Move %d redone successfully.\n", target.ID)
	return ExitOK
}

func redoLastBatch(database *db.DB, scanner *bufio.Scanner, f ScopeFilter) int {
	f, ok := ConfirmCwdScopeWithPreview(scanner, f, "Redo", redoableCountsFn(database))
	if !ok {
		return ExitScopeRejected
	}
	batchID := findLastRevertedBatchInScope(database, f)
	if batchID == "" {
		fmt.Println("📭 No reverted batch operations to redo in this scope.")
		return ExitNothingMatched
	}

	batchActions, err := database.ListActionsByBatch(batchID)
	if err != nil {
		errlog.Error("Cannot read batch %s: %v", batchID, err)
		return ExitGenericError
	}

	scoped := FilterActionsWith(batchActions, f)
	redoable := countReverted(scoped)
	skipped := countScopeSkipped(countReverted(batchActions), redoable)
	if redoable == 0 {
		fmt.Println("📭 Batch has no reverted actions to redo (or all filtered out).")
		printHistorySummary(HistorySummary{Verb: "Redo", Skipped: skipped})
		return ExitNothingMatched
	}

	shortBatch := batchID
	if len(shortBatch) > 8 {
		shortBatch = shortBatch[:8]
	}

	fmt.Printf("⏩ Redo batch %s (%d actions, %d skipped by filter):\n",
		shortBatch, redoable, skipped)
	printRevertedActions(scoped)
	if !confirmRedo(scanner) {
		return ExitRowDeclined
	}

	failed := executeRedoBatch(database, scoped)
	printRedoBatchResult(shortBatch, redoable, failed)
	printHistorySummary(HistorySummary{
		Verb:     "Redo",
		Matched:  redoable,
		Executed: redoable - failed,
		Failed:   failed,
		Skipped:  skipped,
	})
	if failed > 0 && failed == redoable {
		return ExitGenericError
	}
	return ExitOK
}

func redoLastOperation(database *db.DB, scanner *bufio.Scanner, f ScopeFilter) int {
	f, ok := ConfirmCwdScopeWithPreview(scanner, f, "Redo", redoableCountsFn(database))
	if !ok {
		return ExitScopeRejected
	}
	lastMove := pickLastRedoableMove(database, f)
	lastAction := pickLastRedoableAction(database, f)
	skipped := countRedoableMoveSkipped(database, f) +
		countRedoableActionSkipped(database, f)

	haveMove := lastMove != nil
	haveAction := lastAction != nil

	if !haveMove && !haveAction {
		fmt.Println("📭 No reverted operations to redo in this scope.")
		printHistorySummary(HistorySummary{Verb: "Redo", Skipped: skipped})
		return ExitNothingMatched
	}

	if haveMove && !haveAction {
		return runSingleRedoMove(database, scanner, lastMove, skipped)
	}

	if haveAction && !haveMove {
		return runSingleRedoAction(database, scanner, lastAction, skipped)
	}

	// Both available — pick the newest reverted
	if lastAction.CreatedAt >= lastMove.MovedAt {
		return runSingleRedoAction(database, scanner, lastAction, skipped)
	}
	return runSingleRedoMove(database, scanner, lastMove, skipped)
}

func runSingleRedoMove(database *db.DB, scanner *bufio.Scanner, m *db.MoveRecord, skipped int) int {
	code := redoSingleMoveCode(database, scanner, m)
	executed, failed := 0, 0
	switch code {
	case ExitOK:
		executed = 1
	case ExitGenericError:
		failed = 1
	}
	printHistorySummary(HistorySummary{
		Verb: "Redo", Matched: 1, Executed: executed, Failed: failed, Skipped: skipped,
	})
	return code
}

func runSingleRedoAction(database *db.DB, scanner *bufio.Scanner, a *db.ActionRecord, skipped int) int {
	code := redoSingleActionCode(database, scanner, a)
	executed, failed := 0, 0
	switch code {
	case ExitOK:
		executed = 1
	case ExitGenericError:
		failed = 1
	}
	printHistorySummary(HistorySummary{
		Verb: "Redo", Matched: 1, Executed: executed, Failed: failed, Skipped: skipped,
	})
	return code
}

func redoSingleMoveCode(database *db.DB, scanner *bufio.Scanner, m *db.MoveRecord) int {
	fmt.Println("⏩ Redo last move:")
	fmt.Printf("   %s → %s\n", m.FromPath, m.ToPath)
	LogRedoMoveTarget(m)
	if !confirmRedo(scanner) {
		return ExitRowDeclined
	}
	if err := executeMoveRedo(database, m); err != nil {
		errlog.Error("Redo failed: %v", err)
		return ExitGenericError
	}
	fmt.Println("✅ Redo successful!")
	return ExitOK
}

func redoSingleActionCode(database *db.DB, scanner *bufio.Scanner, a *db.ActionRecord) int {
	printActionRedo(a)
	LogRedoActionTarget(a)
	if !confirmRedo(scanner) {
		return ExitRowDeclined
	}
	if err := executeActionRedo(database, a); err != nil {
		errlog.Error("Redo failed: %v", err)
		return ExitGenericError
	}
	fmt.Println("✅ Redo successful!")
	return ExitOK
}

// pickLastRedoableMove returns the newest reverted move under scope.
func pickLastRedoableMove(database *db.DB, f ScopeFilter) *db.MoveRecord {
	moves, err := database.ListMoveHistory(undoMoveScanLimit)
	if err != nil {
		return nil
	}
	for _, m := range FilterMovesWith(moves, f) {
		if !m.IsReverted {
			continue
		}
		return &m
	}
	return nil
}

// pickLastRedoableAction returns the newest reverted action under scope.
func pickLastRedoableAction(database *db.DB, f ScopeFilter) *db.ActionRecord {
	actions, err := database.ListActions(undoActionScanLimit)
	if err != nil {
		return nil
	}
	for _, a := range FilterActionsWith(actions, f) {
		if !a.IsReverted {
			continue
		}
		return &a
	}
	return nil
}

// findLastRevertedBatchInScope finds the most recent reverted batch where at
// least one action passes the scope+glob filter. Empty filter → unfiltered.
func findLastRevertedBatchInScope(database *db.DB, f ScopeFilter) string {
	actions, err := database.ListActions(200)
	if err != nil {
		return ""
	}
	for _, a := range actions {
		if !a.IsReverted || a.BatchId == "" {
			continue
		}
		if !batchTouchesScope(database, a.BatchId, f) {
			continue
		}
		return a.BatchId
	}
	return ""
}

// redoableCountsFn mirrors undoableCountsFn for the redo flows.
func redoableCountsFn(database *db.DB) ScopePreviewFn {
	return func(f ScopeFilter) (int, int) {
		rawMoves, _ := database.ListMoveHistory(undoMoveScanLimit)
		rawActions, _ := database.ListActions(undoActionScanLimit)
		return countRevertedMoves(FilterMovesWith(rawMoves, f)),
			countReverted(FilterActionsWith(rawActions, f))
	}
}
