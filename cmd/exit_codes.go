// SHARED HELPER: exit_codes.go — distinct, documented exit codes for scriptable
// exit_codes.go — distinct, documented exit codes for scriptable
// undo/redo runs. Importable from anywhere in cmd/.
//
// Stable contract — DO NOT renumber. CI scripts depend on these values.
//
// Range allocation:
//
//	0       success / nothing to do
//	2       generic error (DB open, FS failure, etc.)
//	10..19  user declined / canceled
//	20..29  filter / scope produced empty result
//
// Codes 1 and >127 are reserved (1 is shell convention for unspecified
// failure; 128+ collides with signal exit codes).
package cmd

import (
	"fmt"
	"os"
)

const (
	// ExitOK is the implicit zero-exit success path. Returned when the
	// operation completed with at least one row applied OR when the user
	// asked for a read-only listing that succeeded.
	ExitOK = 0

	// ExitGenericError covers DB open failures, filesystem errors during
	// rename/delete, malformed snapshots — anything the user can't fix
	// by retrying with a different prompt answer.
	ExitGenericError = 2

	// ExitScopeRejected is returned when the user explicitly answered "n"
	// (or anything that wasn't yes/global/list) at the cwd-scope
	// confirmation prompt. Distinguishable in CI from per-row decline.
	ExitScopeRejected = 10

	// ExitRowDeclined is returned when the user proceeded past the scope
	// prompt but said "n" at the per-row "Undo this? [y/N]" prompt.
	// Also used when the scanner returned EOF mid-prompt (which scripts
	// often hit when piping commands).
	ExitRowDeclined = 11

	// ExitNothingMatched is returned when the filter dropped everything
	// (no row in scope to act on). Useful for CI loops that want to
	// short-circuit instead of treating "no work" as success.
	ExitNothingMatched = 20
)

// exitLabel returns a short human-readable description for logs.
func exitLabel(code int) string {
	switch code {
	case ExitOK:
		return "ok"
	case ExitGenericError:
		return "error"
	case ExitScopeRejected:
		return "scope rejected"
	case ExitRowDeclined:
		return "row declined"
	case ExitNothingMatched:
		return "nothing matched"
	}
	return "unknown"
}

// osExit is the indirection point so tests can stub the real os.Exit.
// Production code MUST call exitWithCode (not osExit directly) so the
// final summary line is always printed.
var osExit = os.Exit

// exitFootPrintf renders the trailing "exit: N (label)" footer. Lifted
// to a var so tests can capture it without touching stdout.
var exitFootPrintf = func(format string, a ...interface{}) {
	fmt.Println()
	fmt.Printf(format+"\n", a...)
}

// exitWithCode is the single os.Exit chokepoint for undo/redo runners.
// Code 0 returns silently (no extra output) so happy-path runs stay
// clean. Non-zero codes always print a labeled footer and call
// os.Exit so the shell sees the real exit status.
func exitWithCode(code int) {
	if code == ExitOK {
		return
	}
	exitFootPrintf("exit: %d (%s)", code, exitLabel(code))
	osExit(code)
}
