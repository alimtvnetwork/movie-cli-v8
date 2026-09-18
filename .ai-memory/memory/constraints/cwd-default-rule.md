---
name: CWD default rule
description: Every command with optional [path] arg MUST default to os.Getwd(). Never silently return empty. Use ResolveTargetDir helper.
type: constraint
---

# Universal CWD-default rule

## Rule
Any cobra command that accepts an optional `[directory]` / `[path]` argument MUST:

1. Use `cmd.ResolveTargetDir(args, home)` (defined in `cmd/path_resolver.go`).
2. If no arg is given → fall back to `os.Getwd()`.
3. NEVER prompt interactively as the silent fallback. NEVER return `""` on cancel.
4. On any resolution error → log loudly via `errlog.Error(...)` and `return`.

## Why
Bug v2.135.0 → v2.136.0: `movie popout` (and `movie move`) called `promptSourceDirectory()` when no arg was given. If stdin was closed/non-TTY, the prompt returned `""`, the command exited with no output and no error. User had no idea anything happened. See `mem://issues/08-popout-silent-failure`.

## How to apply
- New commands: import `ResolveTargetDir` and use it in the very first lines of the `Run` function.
- Existing commands: `movie scan` already had its own equivalent (`resolveScanDir` in `movie_scan_helpers.go`) — leave it; it does the same thing.
- File-walk commands updated in v2.136.0: `popout`, `move`. Other commands (`ls`, `info`, `play`, `tag`, `watch`, `stats`, `cleanup`, `rescan`, `rename`, etc.) take IDs or operate on the whole DB — no path arg → no change needed.
