# Self-Update Architecture — GitMap Gold Standard

> Canonical reference for self-updating movie-cli based on the GitMap Gold Standard
> ([`02-spec/13-generic-cli/22-self-update-gold-standard.md`](02-spec/13-generic-cli/22-self-update-gold-standard.md)).

---

## 1. Primary Flow: Canonical Remote Installer (Default)

Most users install `movie-cli` via prebuilt release binaries from GitHub and do not have a local
source checkout or Go toolchain.

1. `movie update` runs `RunRemoteUpdate()` by default:
   - Fetches official `install.ps1` (Windows) or `install.sh` (Unix).
   - Executes the installer directly with `-InstallDir <dir>`, inheriting stdin/stdout/stderr.
   - The installer verifies SHA256, applies rename-first deploy, and refreshes the binary in seconds.
2. If offline or if `--source-rebuild` is supplied, falls back to the source-rebuild handoff flow.

---

## 2. Secondary Flow: Two-Phase Handoff for Source Rebuilds (`--source-rebuild`)

When rebuilding from source via `run.ps1`:

1. **Phase 1 — Handoff from active binary:**
   - Active binary copies itself to `movie-update-<pid>.exe`.
   - Launches `movie-update-<pid>.exe update-runner` using **`cmd.Run()` (foreground/blocking)**.
   - Parent stays attached, stdout/stderr/stdin are inherited.

2. **Phase 2 — Build and rename-first deploy:**
   - The handoff copy executes `run.ps1 -Update -TargetBinaryPath <path>`.
   - `run.ps1` builds `./bin/movie.exe`.
   - Windows allows renaming a running executable: `movie.exe` is renamed to `movie.exe.old`.
   - The newly built binary is copied directly into place.
   - The worker verifies the deployed binary and completes.

3. **Phase 3 — Synchronous cleanup:**
   - Because `cmd.Run()` blocked until the worker finished, the worker process is now dead.
   - The parent unblocks, immediately deletes `movie-update-<pid>.exe`, and runs `update-cleanup`.
   - The terminal session remains stable and uninterrupted.

---

## 3. Hard Prohibitions

- Never use `cmd.Start()` + detached console for update handoff.
- Never directly overwrite a running binary without rename-first.
- Never disable CI/CD quality gates.
