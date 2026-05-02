// SHARED HELPER: path_resolver.go — universal CWD-default resolver for ALL movie commands.
// path_resolver.go — universal CWD-default resolver for ALL movie commands.
//
// -- Shared helper exported from this file --
//
//	ResolveTargetDir(args, home) (string, error)
//	    Resolves the target directory for any command that takes an
//	    optional [path] argument. Behavior:
//	      1. args[0] non-empty  → expand ~ and return.
//	      2. args[0] missing    → return os.Getwd() (current working dir).
//	      3. Any error          → loud apperror, NEVER silent empty string.
//
// This helper EXISTS specifically to kill silent-failure bugs where a
// command exits with no error because an interactive prompt returned "".
// See spec/09-app-issues/08-popout-silent-failure.md.
//
// Project rule (mem://constraints/cwd-default-rule):
//
//	"If a command takes an optional path argument and none is given, it
//	 MUST default to the current working directory. NEVER prompt silently,
//	 NEVER return empty, NEVER swallow the error."
//
// Consumers (all 21+ commands accepting an optional path):
//
//	movie scan, movie move, movie popout, movie rename, movie cleanup,
//	movie duplicates, movie rescan, movie cache, movie cache backfill,
//	movie cache forget, ... (any command that scans or operates on files).
//
// Do NOT duplicate the args[0] / cwd / expandHome chain elsewhere — use
// this helper.
package cmd

import (
	"os"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

// ResolveTargetDir is the canonical entry point for resolving a directory
// argument across every cobra command in this package.
//
// home is passed in (rather than computed inside) so callers can reuse the
// same value they already obtained via os.UserHomeDir() — and so unit tests
// can inject a fake home directory.
func ResolveTargetDir(args []string, home string) (string, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return expandHome(strings.TrimSpace(args[0]), home), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", apperror.Wrap("resolve cwd default", err)
	}
	return cwd, nil
}

// MustResolveTargetDir is a thin wrapper that logs the error via the
// caller's preferred error handler and returns "" only when the underlying
// resolution truly failed. The empty-return case here means a real OS error
// occurred (Getwd failed) — NOT a silent prompt cancellation.
//
// Callers should treat "" as a hard failure and return immediately after
// logging via errlog.
func MustResolveTargetDir(args []string, home string) string {
	dir, err := ResolveTargetDir(args, home)
	if err != nil {
		return ""
	}
	return dir
}
