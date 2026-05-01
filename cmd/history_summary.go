// SHARED HELPER: history_summary.go — shared post-operation summary for movie undo/redo.
// history_summary.go — shared post-operation summary for movie undo/redo.
//
// Whenever an undo/redo flow finishes (single op, batch, or just `--list`),
// we want the user to see four numbers:
//
//	matched   — rows that passed the dir scope + glob filter
//	executed  — rows we actually undid/redid (matched − failed for batch;
//	            1 for single op when it succeeded)
//	failed    — rows where the filesystem operation errored out
//	skipped   — rows that exist in the DB but were dropped by the
//	            dir-scope or glob filter (i.e. "out of scope")
//
// Single-op flows just set Matched=1 + Executed=1 (or Failed=1) and
// compute Skipped from raw vs filtered counts.
package cmd

import "fmt"

// HistorySummary captures the counters reported at the end of an
// undo/redo run. Verb is "Undo" or "Redo" — used in the banner.
type HistorySummary struct {
	Verb     string // "Undo" or "Redo"
	Matched  int    // rows passing the filter
	Executed int    // rows actually applied
	Failed   int    // rows that failed during execution
	Skipped  int    // rows dropped by dir/glob filter
}

// printHistorySummary writes a 4-line block. Always called even when all
// counters are zero — gives the user explicit "nothing happened" feedback
// instead of silent exit.
func printHistorySummary(s HistorySummary) {
	fmt.Println()
	fmt.Printf("📊 %s summary\n", s.Verb)
	fmt.Printf("   ✅ executed:          %d\n", s.Executed)
	fmt.Printf("   ⚠️  failed:           %d\n", s.Failed)
	fmt.Printf("   🔍 matched filter:    %d\n", s.Matched)
	fmt.Printf("   🚫 skipped (out of scope): %d\n", s.Skipped)
}

// countScopeSkipped returns the number of rows the filter dropped.
// Negative results clamp to zero (defensive — should never happen).
func countScopeSkipped(raw, kept int) int {
	diff := raw - kept
	if diff < 0 {
		return 0
	}
	return diff
}

// PreviewSummary is the structured counter block shown by `--list`
// (movie undo --list / movie redo --list). It is intentionally separate
// from HistorySummary because preview output should never advertise
// "executed: 0" / "failed: 0" — nothing was applied.
//
// Verb is "Undo" or "Redo" (not "(preview)" — printPreviewSummary adds
// the suffix itself so callers can't drift from the convention).
//
// All three "matched" columns are scoped + glob-filtered counts of rows
// the user could act on right now. The "skipped" columns are rows that
// exist in the underlying history but were dropped by the dir scope or
// `--include` / `--exclude` patterns.
type PreviewSummary struct {
	Verb            string // "Undo" or "Redo"
	AlreadyDoneHint string // optional copy: "already reverted" / "not reverted yet"
	MatchedMoves    int    // moves passing filter and still actionable
	MatchedActions  int    // actions passing filter and still actionable
	SkippedMoves    int    // moves dropped by filter
	SkippedActions  int    // actions dropped by filter
}

// printPreviewSummary writes a labeled preview block. Designed to make
// it impossible for the user to confuse "what the filter matched" with
// "what would actually be undone/redone if you ran the command".
//
// Example (movie undo --list):
//
//	📋 Undo preview (no changes made)
//	   🎯 ready to undo:     4   (1 moves, 3 actions)
//	   🚫 out of scope:      9   (2 moves, 7 actions)
//	   ℹ️  run without --list to actually undo.
func printPreviewSummary(s PreviewSummary) {
	matched := s.MatchedMoves + s.MatchedActions
	skipped := s.SkippedMoves + s.SkippedActions
	verb := s.Verb
	if verb == "" {
		verb = "History"
	}
	fmt.Println()
	fmt.Printf("📋 %s preview (no changes made)\n", verb)
	fmt.Printf("   🎯 ready to %s:    %d   (%d moves, %d actions)\n",
		verbLower(verb), matched, s.MatchedMoves, s.MatchedActions)
	fmt.Printf("   🚫 out of scope:     %d   (%d moves, %d actions)\n",
		skipped, s.SkippedMoves, s.SkippedActions)
	if matched > 0 {
		fmt.Printf("   ℹ️  run `movie %s` (without --list) to actually %s.\n",
			verbLower(verb), verbLower(verb))
	}
}

// verbLower normalises "Undo"/"Redo" → "undo"/"redo" without pulling in
// strings just for one ToLower call (file already keeps imports minimal).
func verbLower(v string) string {
	switch v {
	case "Undo":
		return "undo"
	case "Redo":
		return "redo"
	}
	return v
}
