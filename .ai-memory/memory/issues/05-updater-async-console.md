# Issue: Updater Async Console — Two-Iteration Saga

> **Status**: ✅ Resolved (correctly this time, iteration 5)
> **Severity**: High
> **Files**: `updater/handoff.go`, `updater/handoff_windows.go`,
> `updater/handoff_unix.go`, `updater/script.go`, `updater/run.go`,
> `cmd/update.go`, `spec/13-self-update-app-update/03-copy-and-handoff.md`,
> `HANDOFF-LESSONS.md`
> **Iterations**: 5 (final fix 19-Apr-2026)

## TL;DR

Two opposite "fixes" for the same handoff. Iteration 3 was wrong; it
caused the real bug the user reported. Iteration 4 restores the correct
detached behaviour and adds a self-deleter for the worker binary.

## Iteration history

| # | Date | Change | Result |
|---|------|--------|--------|
| 1 | early | `cmd.Start()` + `os.Exit(0)` | Worked; console looked detached. |
| 2 | early | Added cleanup command for stray workers | OK. |
| 3 | 16-Apr-2026 | Switched to `cmd.Run()` (blocking) so console stayed attached | **BROKE Windows updates** — parent kept lock on `movie.exe`. |
| 4 | 17-Apr-2026 | Reverted to detached spawn + parent exits 0; worker self-deletes via detached `cmd /c del`; new console window keeps output visible | ✅ Fixed the blocking-lock regression. |
| 5 | 19-Apr-2026 | Passed the full original binary path through to `run.ps1` as `-TargetBinaryPath`; normalized cleanup skip-path; replaced fragile updater glyphs with ASCII-safe labels | ✅ Fixed mixed-path Windows handoff/cleanup regression and garbled output. |

## Root cause of iteration 3 regression

`cmd.Run()` is blocking, so the original `movie.exe` process never
exited during the update. On Windows that means:

1. `run.ps1`'s "active PATH binary sync" loop saw the file as locked
   → printed "Active PATH binary is in use; retrying (1..5/5)" → gave up.
2. The post-deploy `update-cleanup` step tried to delete
   `movie-update-<pid>.exe` — but that file is the worker process
   currently running the script → `Access is denied`.

The user's words: *"you didn't stop the current running process. So your
current running process should be stopped and moved to that copied
binary."* — which is the textbook definition of the handoff and exactly
what iteration 3 had stopped doing.

## Solution (iteration 4)

- `launchHandoff` now uses **`cmd.Start()` + `cmd.Process.Release()`**
  with platform-specific creation flags:
  - Windows: `CREATE_NEW_CONSOLE | CREATE_NEW_PROCESS_GROUP` so the
    worker has its own visible console window.
  - Unix: `Setsid: true`, stdio nil, fully detached.
- `Run()` returns nil after a successful spawn so `cmd/update.go` exits 0
  and the OS releases the lock on the original `movie.exe`.
- The temp PowerShell script (`updater/script.go`) ends with a tiny
  detached `cmd /c ping 127.0.0.1 -n 3 & del "<workerBinary>"` so the
  worker copy deletes itself ~2 s after exit.
- `update-cleanup` keeps `--skip-path <workerBinary>` as a safety net.

## What NOT to repeat

- ❌ Never use `cmd.Run()` (blocking) for the handoff launch.
- ❌ Never try to delete the worker binary from inside the worker process.
- ❌ Never "fix" a perceived console issue by making the parent block —
  put the fix in the worker's own console instead.
- ❌ Never trust a self-update test that only checks "deploy printed OK"
  — also verify the active PATH binary was replaced and no stray
  `movie-update-*.exe` remains.
- ❌ Never split a known full target-binary path back into separate deploy
  arguments for `run.ps1`. Pass the exact path through unchanged.

## Additional root cause found in iteration 5

The detached handoff itself was correct again, but one more bug remained in
the worker stage on Windows machines with multiple deployed `movie.exe`
copies.

The worker already knew the exact original binary path from
`--target-binary <full-path>`, but `updater/script.go` split it into
`-DeployPath` and `-BinaryNameOverride` before invoking `run.ps1`.
That reconstruction left room for update-mode deploy/verify/sync behavior to
drift toward config or PATH-resolved locations instead of staying pinned to
the exact binary the user launched.

At the same time, cleanup preservation depended on a raw `--skip-path`
string match, so quoted or padded values were more fragile than they needed
to be, and the Unicode glyphs used in updater status lines were getting
garbled in some PowerShell consoles.

## Prevention rule added in iteration 5

- `update-runner --target-binary <full-path>` must forward the **same full
  path** into `run.ps1 -TargetBinaryPath <full-path>`.
- If the updater already knows the exact executable path, later stages must
  not reconstruct it from directory + filename pieces.
- Cleanup path comparisons must normalize quotes and surrounding whitespace
  before equality checks.
- Updater-facing Windows console messages should prefer ASCII-safe status
  markers unless a UTF-8-safe path is guaranteed end-to-end.

## See also

- `spec/13-self-update-app-update/03-copy-and-handoff.md` — authoritative spec.
- `HANDOFF-LESSONS.md` — short shareable doc for other AIs / contributors.
