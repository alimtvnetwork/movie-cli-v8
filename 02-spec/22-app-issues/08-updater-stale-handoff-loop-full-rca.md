# Issue 08 — Updater stale-handoff loop (full failure chain RCA)

**Severity:** Critical (every `movie update` looked successful but left the active binary frozen on v2.97.0)
**Versions involved:** v2.97.0 (frozen) → v2.118.0 → v2.119.0 → v2.120.0 → v2.121.0 (root-cause fix) → v2.122.0 (`self-replace` bootstrap) → v2.123.0 (output polish)
**Files:** `run.ps1`, `updater/script.go`, `updater/cleanup.go`, `updater/handoff.go`, `updater/self_replace.go`, `cmd/self_replace.go`, `version/info.go`, `CHANGELOG.md`

## Symptom (verbatim, last run before fix landed)

```
🎯 Active binary: E:\bin-run\movie.exe
  Version before: movie-cli v2.97.0   Commit: 6287058
  ...
  OK Deployed to: D:\bin-run\movie.exe
  !! PATH resolves 'movie' to a different binary: E:\bin-run\movie.exe
  !! Active PATH binary is in use; retrying (1/5)...
  !! Active PATH binary is in use; retrying (2/5)...
  !! Active PATH binary is in use; retrying (3/5)...
  !! Active PATH binary is in use; retrying (4/5)...
  !! Could not sync active PATH binary after retries.
movie.exe :   ⚠ Could not remove movie-update-29152.exe: Access is denied.
    + FullyQualifiedErrorId : NativeCommandError
```

Every successive `movie update` reported **the same `Version before: v2.97.0`**, even though the deploy step claimed success.

## Failure chain (5 independent bugs, one symptom)

### Bug 1 — Two `bin-run` directories on different drives

| Path           | Source                         | Contained             |
|----------------|--------------------------------|-----------------------|
| `D:\bin-run`   | `powershell.json` `deployPath` | newly-built binary    |
| `E:\bin-run`   | `$env:PATH`                    | frozen v2.97.0 binary |

The user's environment had two install locations. `powershell.json` pointed
the build pipeline at `D:`, but the active `movie` on PATH lived on `E:`.
Neither location was wrong on its own — they just diverged.

### Bug 2 — Update mode trusted `powershell.json`'s `deployPath`

`Resolve-DeployTarget` in `run.ps1` derived the deploy target from
`$config.deployPath` whenever `-TargetBinaryPath` was empty. Result: every
update wrote the new binary to `D:\bin-run\movie.exe` while the user kept
running `E:\bin-run\movie.exe`. The deploy was a no-op from the user's
perspective.

### Bug 3 — Old worker never sent `-TargetBinaryPath`

The handoff worker that `movie update` invoked was a **copy of the active
PATH binary** — i.e. v2.97.0. v2.97.0's `updater/script.go` predates
`-TargetBinaryPath` entirely. The fixes shipped in v2.118.0 / v2.119.0 /
v2.120.0 (which all required the worker to forward the target path) could
never take effect because the worker calling them was always v2.97.0.

This is the **stale-handoff trap**: the bug prevents its own fix from being
installed.

### Bug 4 — Post-deploy PATH-sync retry loop attacked the live parent

After deploy, `run.ps1` tried to copy `D:\bin-run\movie.exe` over
`E:\bin-run\movie.exe`. But `E:\bin-run\movie.exe` is *the parent process
that spawned the worker*. Windows holds an exclusive lock on a running .exe.
Five retries × 1s sleeps later, the script gave up with the visible
"in use" warning spam.

### Bug 5 — Cleanup write to stderr surfaced as PowerShell `NativeCommandError`

`updater/cleanup.go` called `os.Remove(movie-update-29152.exe)` against the
live worker copy, which is locked. The "Access is denied" error was written
to `os.Stderr`, and PowerShell's native-command pipeline wraps any stderr
output in the red `NativeCommandError` block — making a self-healing,
expected condition look like a hard crash.

## Why earlier fixes did not work

| Version  | Attempted fix                                              | Why it failed                                       |
|----------|------------------------------------------------------------|-----------------------------------------------------|
| v2.118.0 | Added `-TargetBinaryPath` plumbing                         | Only **new** workers send it; user's worker is v2.97.0 |
| v2.119.0 | New worker skips PATH-sync loop in update mode             | User's old worker still runs the loop               |
| v2.120.0 | `Normalize-LegacyUpdateArgs` for legacy parameter names    | v2.97.0 worker sends *no* target path, not a legacy one |

Every fix was correct in isolation but invisible in production because the
broken binary couldn't be replaced *by itself*.

## Root-cause fix (v2.121.0)

Single repo-side change that works regardless of which old worker calls in:

1. **`Resolve-DeployTarget` falls back to `Get-Command movie`** when
   `-Update` is set and `-TargetBinaryPath` is empty. The deploy target is
   now the binary the user is *actually running* — `powershell.json`'s
   `deployPath` is ignored in update mode unless explicitly overridden.
2. **Post-deploy PATH-sync loop is unconditionally skipped in `-Update`
   mode.** The deploy already wrote to the right file.
3. **`updater/cleanup.go` silently skips locked `*-update-*` workers** and
   routes remaining warnings to stdout. PowerShell no longer wraps cleanup
   output in `NativeCommandError`.

## Bootstrap escape hatch (v2.122.0)

Even with v2.121.0 in the repo, the user is still running v2.97.0 locally.
v2.122.0 added a `movie self-replace` command that atomically replaces the
active PATH binary with the freshly-deployed one using a rename-first copy
(rename the locked exe to `*.old`, copy the new one in, schedule the `.old`
for cleanup). This is the one-shot break-out command — after running it
once, every future `movie update` works correctly because the worker is now
≥ v2.121.0.

## Output polish (v2.123.0)

With the functional bug fixed, v2.123.0 normalised the visual output:
consistent indent levels (0 / 1 / 2 / 3 spaces), bracketed status tags
(`[ OK ]`, `[INFO]`, `[WARN]`, `[ERR ]`), no emoji noise, suppressed the
harmless `.bak` access-denied warning, and split the final summary into
distinct `from:` / `to:` rows.

## Prevention rules (added to memory)

- **Update mode must never trust `powershell.json` `deployPath`.** Always
  derive the deploy target from the active PATH binary, with
  `-TargetBinaryPath` as the only override.
- **Never write expected `os.Remove` failures on `*-update-*` artifacts to
  stderr.** PowerShell turns any stderr line into a `NativeCommandError`
  block, which looks like a crash to the user.
- **Never run a post-deploy retry loop that copies over the parent
  process's own .exe.** The lock is held for the entire parent lifetime.
- **Any updater fix must include a bootstrap path** (manual command or
  `self-replace`) so users on the broken version can break the loop without
  hand-editing files.

## Verification checklist

- [ ] `movie update` from v2.97.0 → bootstraps correctly via `self-replace` once
- [ ] `movie update` after bootstrap reports the *new* `Version before:` next run
- [ ] No "Active PATH binary is in use; retrying" lines
- [ ] No `NativeCommandError` red block at end of run
- [ ] No `.bak` access-denied warning
- [ ] Final summary shows `from: vOLD` and `to: vNEW` on separate aligned rows
