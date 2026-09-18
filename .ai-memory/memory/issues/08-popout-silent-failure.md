---
name: Popout silent failure
description: v2.135 popout/move silently exited when no path arg + closed stdin. Fixed v2.136.0 with ResolveTargetDir helper + integration tests.
type: feature
---

# Issue 08 — Popout & move silent failure (v2.135 → v2.136.0)

## Symptom
User ran `movie popout` (no args) inside a folder with nested media. No output, no error, exit 0. Same for `movie move`.

## Root cause
`runMoviePopout` → `resolvePopoutDir(args, mc)` → if `len(args) == 0`, called `promptSourceDirectory(scanner, db, home)`. That helper does `scanner.Scan()` and returns `""` if the scan fails (closed stdin, piped input, non-TTY, EOF). The caller then did:

```go
if rootDir == "" { return }
```

→ silent exit. No error logged. No way for the user to know what happened.

## Fix (v2.136.0)
1. New `cmd/path_resolver.go` exposes `ResolveTargetDir(args, home) (string, error)`:
   - args[0] given → expand `~`, return.
   - empty → return `os.Getwd()`.
   - any failure → loud `apperror`.
2. `runMoviePopout` and `runMovieMove` rewritten to use it. They now ALWAYS print the resolved directory before scanning.
3. Old destructive folder-cleanup prompt replaced by `.temp/` compaction (see `mem://features/popout-spec`).
4. Integration tests in `cmd/movie_popout_integration_test.go` pin 7 invariants — including `TestResolveTargetDir_DefaultsToCwdWhenNoArg` which is the regression fence for this bug.

## Prevention rule
See `mem://constraints/cwd-default-rule` — every new command with an optional path arg MUST use `ResolveTargetDir`.
