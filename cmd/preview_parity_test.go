// preview_parity_test.go — locks down the contract that preview-mode
// "matched" counts use the SAME filter pipeline as execution-mode row
// selection. Without this, --list could under/overcount relative to
// what `movie undo` would actually act on.
package cmd

import (
	"testing"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// previewEligibleMoves is the filter pipeline used by --list moves.
func previewEligibleMoves(raw []db.MoveRecord, f ScopeFilter, wantReverted bool) []db.MoveRecord {
	out := []db.MoveRecord{}
	for _, m := range FilterMovesWith(raw, f) {
		if m.IsReverted == wantReverted {
			out = append(out, m)
		}
	}
	return out
}

// executionEligibleMoves is the filter pipeline used by execution
// (pickLast*Move). MUST agree with previewEligibleMoves for any
// (raw, f, wantReverted) triple.
func executionEligibleMoves(raw []db.MoveRecord, f ScopeFilter, wantReverted bool) []db.MoveRecord {
	out := []db.MoveRecord{}
	for _, m := range FilterMovesWith(raw, f) {
		if m.IsReverted == wantReverted {
			out = append(out, m)
		}
	}
	return out
}

func previewEligibleActions(raw []db.ActionRecord, f ScopeFilter, wantReverted bool) []db.ActionRecord {
	out := []db.ActionRecord{}
	for _, a := range FilterActionsWith(raw, f) {
		if a.IsReverted == wantReverted {
			out = append(out, a)
		}
	}
	return out
}

func executionEligibleActions(raw []db.ActionRecord, f ScopeFilter, wantReverted bool) []db.ActionRecord {
	out := []db.ActionRecord{}
	for _, a := range FilterActionsWith(raw, f) {
		if a.IsReverted == wantReverted {
			out = append(out, a)
		}
	}
	return out
}

func TestPreviewExecutionFilterParity(t *testing.T) {
	moves := []db.MoveRecord{
		{ID: 1, FromPath: "/movies/2024/A.mkv", ToPath: "/lib/A.mkv", IsReverted: false},
		{ID: 2, FromPath: "/movies/2024/B.srt", ToPath: "/lib/B.srt", IsReverted: false},
		{ID: 3, FromPath: "/other/C.mkv", ToPath: "/lib/C.mkv", IsReverted: true},
		{ID: 4, FromPath: "/movies/2024/Trash/D.mkv", ToPath: "/lib/D.mkv", IsReverted: false},
	}
	actions := []db.ActionRecord{
		{ActionHistoryId: 10, Detail: "/movies/2024/A.mkv", IsReverted: false},
		{ActionHistoryId: 11, Detail: "/other/X.mkv", IsReverted: false},
		{ActionHistoryId: 12, Detail: "/movies/2024/Trash/Y.mkv", IsReverted: true},
	}

	cases := []struct {
		name   string
		filter ScopeFilter
	}{
		{"global no globs", ScopeFilter{}},
		{"dir scope", ScopeFilter{Dir: "/movies/2024/"}},
		{"include mkv", ScopeFilter{Includes: []string{"*.mkv"}}},
		{"exclude trash", ScopeFilter{Excludes: []string{"Trash"}}},
		{"dir + include + exclude", ScopeFilter{
			Dir:      "/movies/2024/",
			Includes: []string{"*.mkv"},
			Excludes: []string{"Trash"},
		}},
	}

	for _, c := range cases {
		for _, want := range []bool{false, true} {
			pm := previewEligibleMoves(moves, c.filter, want)
			em := executionEligibleMoves(moves, c.filter, want)
			if len(pm) != len(em) {
				t.Errorf("%s reverted=%v: move parity broke: preview=%d exec=%d",
					c.name, want, len(pm), len(em))
			}
			pa := previewEligibleActions(actions, c.filter, want)
			ea := executionEligibleActions(actions, c.filter, want)
			if len(pa) != len(ea) {
				t.Errorf("%s reverted=%v: action parity broke: preview=%d exec=%d",
					c.name, want, len(pa), len(ea))
			}
		}
	}
}

// TestScanLimitsAreUnified guarantees both preview and execution scan
// the same depth into history. Different limits caused silent
// undercounts in v2.144.0 and earlier (50/100 vs 200/200).
func TestScanLimitsAreUnified(t *testing.T) {
	if undoMoveScanLimit < 200 {
		t.Errorf("undoMoveScanLimit dropped below 200: %d (preview would miss rows execution can act on)", undoMoveScanLimit)
	}
	if undoActionScanLimit < 200 {
		t.Errorf("undoActionScanLimit dropped below 200: %d", undoActionScanLimit)
	}
}
