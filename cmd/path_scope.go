// path_scope.go — shared path-scoping helpers for history commands
// (movie undo / movie redo / future history filters).
//
// Rule (mem://constraints/cwd-default-rule + this file):
//
//	movie undo  [path]   → scope = path or cwd. --global to override.
//	movie redo  [path]   → scope = path or cwd. --global to override.
//
// "In scope" means: ANY path stored on the action (FromPath, ToPath,
// MediaSnapshot.original_path / compact_path / file_path) is rooted under
// the resolved scope directory.
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// ScopeFilter bundles the directory scope plus optional include/exclude
// glob patterns. All matchers operate on any string path field stored on
// the action / move record (same any-path rule used by ActionInScope).
//
// Semantics:
//   - Includes empty → no include constraint (everything passes the
//     include phase). Otherwise at least one path on the record must
//     match at least one include pattern.
//   - Excludes empty → no exclude constraint. Otherwise the record is
//     dropped if ANY of its paths matches ANY exclude pattern.
//   - Excludes are evaluated AFTER includes, so excludes always win.
//
// Glob syntax is filepath.Match (POSIX shell style: *, ?, [class]).
// Patterns are matched against:
//  1. the full path  (e.g. "/movies/2024/Inception/*.mkv")
//  2. the basename   (e.g. "*.srt")
//  3. the basename of every parent directory (e.g. "Trash")
//
// This makes both "*.mkv" and "Inception" useful without the user having
// to know which form ended up in the snapshot.
type ScopeFilter struct {
	Dir      string   // normalized scope dir ("" → no dir filter / --global)
	Includes []string // glob patterns
	Excludes []string // glob patterns
	// UserProvidedPath is true when the user passed an explicit [path]
	// positional argument or --global. False means the scope was inferred
	// from cwd, in which case interactive flows should confirm with the
	// user before acting.
	UserProvidedPath bool
	// AssumeYes is true when the user passed --yes / -y / --assume-yes.
	// It bypasses BOTH the cwd-scope confirmation prompt AND the per-row
	// "Undo this? [y/N]" prompt — designed for scripted runs.
	AssumeYes bool
}

// HasGlobs reports whether any include or exclude pattern is set.
func (f ScopeFilter) HasGlobs() bool {
	return len(f.Includes) > 0 || len(f.Excludes) > 0
}

// matchAny tries each pattern against full path, basename, and every
// ancestor basename. Returns true on first match.
func matchAny(patterns []string, p string) bool {
	if p == "" || len(patterns) == 0 {
		return false
	}
	candidates := pathCandidates(p)
	for _, pat := range patterns {
		for _, c := range candidates {
			if ok, _ := filepath.Match(pat, c); ok {
				return true
			}
		}
	}
	return false
}

// pathCandidates returns [fullPath, basename, parent-basename, ...].
func pathCandidates(p string) []string {
	clean := filepath.Clean(p)
	out := []string{clean, filepath.Base(clean)}
	dir := filepath.Dir(clean)
	for dir != "." && dir != string(filepath.Separator) && dir != "" {
		out = append(out, filepath.Base(dir))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return out
}

// collectActionPaths returns every string path stored on an action.
func collectActionPaths(a db.ActionRecord) []string {
	paths := []string{}
	if a.Detail != "" {
		paths = append(paths, a.Detail)
	}
	if a.MediaSnapshot == "" {
		return paths
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(a.MediaSnapshot), &raw); err != nil {
		return paths
	}
	return appendStringValues(paths, raw)
}

func appendStringValues(out []string, m map[string]interface{}) []string {
	for _, v := range m {
		switch t := v.(type) {
		case string:
			out = append(out, t)
		case map[string]interface{}:
			out = appendStringValues(out, t)
		}
	}
	return out
}

// MoveMatchesGlobs applies include/exclude globs to a move record.
func MoveMatchesGlobs(m db.MoveRecord, f ScopeFilter) bool {
	paths := []string{m.FromPath, m.ToPath}
	return pathsPassFilter(paths, f)
}

// ActionMatchesGlobs applies include/exclude globs to an action record.
func ActionMatchesGlobs(a db.ActionRecord, f ScopeFilter) bool {
	return pathsPassFilter(collectActionPaths(a), f)
}

// pathsPassFilter implements the include-then-exclude logic.
func pathsPassFilter(paths []string, f ScopeFilter) bool {
	if len(f.Excludes) > 0 {
		for _, p := range paths {
			if matchAny(f.Excludes, p) {
				return false
			}
		}
	}
	if len(f.Includes) == 0 {
		return true
	}
	for _, p := range paths {
		if matchAny(f.Includes, p) {
			return true
		}
	}
	return false
}

// FilterMovesWith applies dir scope + globs in one pass.
func FilterMovesWith(moves []db.MoveRecord, f ScopeFilter) []db.MoveRecord {
	out := make([]db.MoveRecord, 0, len(moves))
	for _, m := range moves {
		if !MoveInScope(m, f.Dir) {
			continue
		}
		if f.HasGlobs() && !MoveMatchesGlobs(m, f) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// FilterActionsWith applies dir scope + globs in one pass.
func FilterActionsWith(actions []db.ActionRecord, f ScopeFilter) []db.ActionRecord {
	out := make([]db.ActionRecord, 0, len(actions))
	for _, a := range actions {
		if !ActionInScope(a, f.Dir) {
			continue
		}
		if f.HasGlobs() && !ActionMatchesGlobs(a, f) {
			continue
		}
		out = append(out, a)
	}
	return out
}

// scopeFromArgs returns the resolved scope directory and a "global" flag.
// When isGlobal is true, the returned dir is "" and callers should not filter.
func scopeFromArgs(args []string, home string, isGlobal bool) string {
	if isGlobal {
		return ""
	}
	dir, err := ResolveTargetDir(args, home)
	if err != nil {
		return ""
	}
	return normalizeScope(dir)
}

// buildScopeFilter is the canonical builder used by undo/redo cobra runs.
// It composes the dir scope with the user-supplied include/exclude globs.
func buildScopeFilter(args []string, home string, isGlobal bool, includes, excludes []string, assumeYes bool) ScopeFilter {
	return ScopeFilter{
		Dir:      scopeFromArgs(args, home, isGlobal),
		Includes: trimEmpty(includes),
		Excludes: trimEmpty(excludes),
		// User explicitly steered the scope when they passed a [path] arg
		// or --global. Otherwise we silently fell back to cwd and need to
		// confirm before doing anything destructive.
		UserProvidedPath: isGlobal || len(args) > 0,
		AssumeYes:        assumeYes,
	}
}

func trimEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// printScopeBanner prints a small "scope:" / "include:" / "exclude:" line
// before list output so the user always sees what's being filtered.
// Imported here to avoid duplicating the formatting in undo & redo files.
func printScopeBanner(f ScopeFilter) {
	if f.Dir != "" {
		fmt.Printf("   scope:    %s\n", f.Dir)
	}
	if len(f.Includes) > 0 {
		fmt.Printf("   include:  %s\n", strings.Join(f.Includes, ", "))
	}
	if len(f.Excludes) > 0 {
		fmt.Printf("   exclude:  %s\n", strings.Join(f.Excludes, ", "))
	}
}

// normalizeScope returns a clean absolute-style suffix-friendly form.
func normalizeScope(dir string) string {
	clean := filepath.Clean(dir)
	if !strings.HasSuffix(clean, string(filepath.Separator)) {
		clean += string(filepath.Separator)
	}
	return clean
}

// pathInScope reports whether p is the scope dir itself or sits under it.
func pathInScope(p, scope string) bool {
	if scope == "" || p == "" {
		return scope == ""
	}
	clean := filepath.Clean(p) + string(filepath.Separator)
	return strings.HasPrefix(clean, scope)
}

// MoveInScope returns true when either side of the move touches scope.
func MoveInScope(m db.MoveRecord, scope string) bool {
	if scope == "" {
		return true
	}
	return pathInScope(m.FromPath, scope) || pathInScope(m.ToPath, scope)
}

// ActionInScope inspects the snapshot JSON for any path field that lives
// under scope. We do NOT rely on a typed struct because different actions
// emit different snapshot shapes (compact uses original_path/compact_path,
// scan uses file_path inside Media, etc.).
func ActionInScope(a db.ActionRecord, scope string) bool {
	if scope == "" {
		return true
	}
	if a.Detail != "" && pathInScope(a.Detail, scope) {
		return true
	}
	return snapshotTouchesScope(a.MediaSnapshot, scope)
}

// snapshotTouchesScope decodes the snapshot as a generic map and walks every
// string value, returning true on the first one rooted under scope.
func snapshotTouchesScope(snapshot, scope string) bool {
	if snapshot == "" {
		return false
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(snapshot), &raw); err != nil {
		return false
	}
	return mapHasScopedString(raw, scope)
}

func mapHasScopedString(m map[string]interface{}, scope string) bool {
	for _, v := range m {
		if valueHasScope(v, scope) {
			return true
		}
	}
	return false
}

func valueHasScope(v interface{}, scope string) bool {
	switch t := v.(type) {
	case string:
		return pathInScope(t, scope)
	case map[string]interface{}:
		return mapHasScopedString(t, scope)
	}
	return false
}

// FilterMoves returns only the moves rooted under scope.
func FilterMoves(moves []db.MoveRecord, scope string) []db.MoveRecord {
	if scope == "" {
		return moves
	}
	out := make([]db.MoveRecord, 0, len(moves))
	for _, m := range moves {
		if MoveInScope(m, scope) {
			out = append(out, m)
		}
	}
	return out
}

// FilterActions returns only the actions touching scope (any-path rule).
func FilterActions(actions []db.ActionRecord, scope string) []db.ActionRecord {
	if scope == "" {
		return actions
	}
	out := make([]db.ActionRecord, 0, len(actions))
	for _, a := range actions {
		if ActionInScope(a, scope) {
			out = append(out, a)
		}
	}
	return out
}

// ConfirmCwdScope asks the user to confirm an inferred cwd scope before
// running a destructive history operation. It is a no-op (returns the
// original filter, true) when:
//   - the user already passed an explicit [path] argument
//   - the user passed --global
//   - the scope dir is empty (== global) for any other reason
//
// When the scope WAS inferred from cwd, the user can:
//
//	[Enter] / y → proceed with the cwd scope
//	g          → switch to --global (returns a global filter, true)
//	n / q      → cancel (returns _, false)
//	l          → re-print the scope details (handy after scrolling)
//
// The verb is "Undo" or "Redo" — used purely for the prompt copy.
//
// previewCounts is an optional callback that, given the current filter,
// returns (matchedMoves, matchedActions). When non-nil and at least one
// of the counts is > 0, the prompt also shows a "would act on" line so
// the user can sanity-check the filter before committing. Pass nil to
// suppress (e.g. for tests or non-interactive contexts).
type ScopePreviewFn func(ScopeFilter) (matchedMoves, matchedActions int)

func ConfirmCwdScope(scanner *bufio.Scanner, f ScopeFilter, verb string) (ScopeFilter, bool) {
	return ConfirmCwdScopeWithPreview(scanner, f, verb, nil)
}

// ConfirmCwdScopeWithPreview is the richer form used by the cobra
// handlers — it can show a live "would act on N moves / N actions"
// estimate before the user confirms. ConfirmCwdScope wraps it with a
// nil callback for callers that don't have DB access.
func ConfirmCwdScopeWithPreview(scanner *bufio.Scanner, f ScopeFilter, verb string, previewCounts ScopePreviewFn) (ScopeFilter, bool) {
	if f.UserProvidedPath || f.Dir == "" {
		return f, true
	}
	if f.AssumeYes {
		fmt.Printf("🎯 %s scope (auto-confirmed via --yes): %s\n", verb, f.Dir)
		printScopeFilterDetails(f, previewCounts)
		return f, true
	}
	for {
		printCwdScopePrompt(verb, f, previewCounts)
		if !scanner.Scan() {
			return f, false
		}
		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "", "y", "yes":
			return f, true
		case "g", "global":
			fmt.Println("   ↳ switching to --global scope")
			return ScopeFilter{
				Includes:         f.Includes,
				Excludes:         f.Excludes,
				UserProvidedPath: true,
				AssumeYes:        f.AssumeYes,
			}, true
		case "l", "list", "show":
			// Re-print and loop — useful when the prompt scrolled off.
			continue
		case "n", "no", "q", "quit":
			fmt.Println("   ↳ canceled")
			return f, false
		default:
			fmt.Println("   ↳ unrecognized choice — canceled")
			return f, false
		}
	}
}

// printCwdScopePrompt renders the multi-line scope confirmation block.
func printCwdScopePrompt(verb string, f ScopeFilter, previewCounts ScopePreviewFn) {
	fmt.Printf("\n🎯 %s scope detected from current directory:\n", verb)
	printScopeFilterDetails(f, previewCounts)
	fmt.Print("   Use this scope?  [Y]es / [g]lobal / [l]ist again / [n]o : ")
}

// printScopeFilterDetails prints the dir, every active include/exclude
// glob, and (when a callback is provided) a "would act on" preview.
func printScopeFilterDetails(f ScopeFilter, previewCounts ScopePreviewFn) {
	fmt.Printf("   📂 directory:   %s\n", f.Dir)
	if len(f.Includes) > 0 {
		fmt.Printf("   ✅ include:     %s\n", strings.Join(f.Includes, ", "))
	} else {
		fmt.Println("   ✅ include:     (no glob — every path under directory)")
	}
	if len(f.Excludes) > 0 {
		fmt.Printf("   🚫 exclude:     %s\n", strings.Join(f.Excludes, ", "))
	}
	if previewCounts != nil {
		moves, actions := previewCounts(f)
		fmt.Printf("   🔢 would act on: %d moves, %d actions\n", moves, actions)
	}
}
