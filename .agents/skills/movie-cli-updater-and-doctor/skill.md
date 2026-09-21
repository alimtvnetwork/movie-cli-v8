---
name: movie-cli-updater-and-doctor
description: "Self-update copy-and-handoff mechanism, environment diagnostics, auto-repairs, and OS shell context menu integration in movie-cli-v8."
---

# Movie CLI Updater and Doctor Skill

## Overview

This subsystem maintains the health, deployment integrity, and lifecycle of `movie-cli-v8`. It solves Windows file-locking during binary self-replacement, provides diagnostic health checks with automated `--fix` remediation, and integrates right-click context menus into host OS file managers.

## Self-Update & Copy-and-Handoff (`updater/`)

### The Windows File-Lock Problem
On Windows, the operating system locks running executable files, preventing an active binary from overwriting or replacing itself on disk during `go build` or file copy.

### The 5-Step Handoff Solution (`updater/run.go`, `updater/handoff.go`)
1. **Preflight:** The active CLI checks for git availability, clean local repository status, and resolves its own absolute executable path.
2. **Worker Duplication:** Copies itself to a temporary worker executable:
   `<deploy-dir>/movie-update-<timestamp>.exe`
3. **Detached Worker Launch:** Spawns the worker executable with internal command:
   `movie-update-<timestamp>.exe update-runner --target <self-path> --repo <repo-path>`
4. **Original Exit:** The parent binary exits immediately with code `0`, releasing its file lock on disk.
5. **Rebuild & Replacement:** The worker executes `run.ps1` to pull latest changes, recompile `go build -o <self-path> main.go`, deploy the new binary, and clean up temporary files via `movie update-cleanup`.

### Update Commands
```sh
movie update                    # Self-updates from local or remote repo
movie update --repo <path>      # Points updater to a specific source repository
```

## Diagnostic Doctor Engine (`doctor/`)

The doctor package diagnoses local installation defects and repairs them automatically.

### Running Doctor
```sh
movie doctor          # Runs full diagnostic audit and prints report
movie doctor --fix    # Automatically repairs fixable defects
```

### Core Diagnostic Checks
1. **PATH Mismatch (`checkPathMismatch`):** Verifies that the running binary matches the configured deployment path in `powershell.json` (prevents running stale copies from alternate directories).
2. **Deploy Directory in PATH (`checkDeployInPath`):** Ensures the folder containing `movie.exe` is present in system or user `$PATH`.
3. **Stale Update Workers (`checkStaleWorkers`):** Detects leftover `*-update-*.exe` temporary binaries from aborted update attempts.
4. **Version Drift (`checkVersionDrift`):** Compares the running version string against the latest commit tag or source tree version.
5. **Environment Configuration (`checks_env.go`):** Validates presence of `TMDB_API_KEY` and network reachability to `api.themoviedb.org`.
6. **Git Status (`repo.go`):** Checks for uncommitted changes or detached HEAD states in source repositories.

## OS Context Menu Integration (`cmd/movie_contextmenu*.go`)

Enables right-clicking any folder in the operating system's file manager to run Movie CLI actions directly.

### Commands
```sh
movie add-contextmenu       # Registers "Movie ▸" context menu in OS shell
movie remove-contextmenu    # Removes registered context menu entries
movie contextmenu-status    # Displays current installation status
```

### Submenu Actions Registered
1. **Scan with Movie:** Runs `movie scan` in clicked directory.
2. **Rescan with Movie:** Runs `movie rescan` in clicked directory.
3. **Open Movie Report:** Opens existing `.movie-output/report.html`.
4. **Show Movie Stats:** Runs `movie stats` in terminal.

### OS Implementations
- **Windows (`movie_contextmenu_windows.go`):** Writes registry keys to `HKCU\Software\Classes\Directory\shell\Movie` and `HKCU\Software\Classes\Directory\Background\shell\Movie`.
- **macOS (`movie_contextmenu_darwin.go`):** Configures Automator Quick Action workflows.
- **Linux (`movie_contextmenu_linux.go`):** Generates `.desktop` action files under `~/.local/share/file-manager/actions/` or Nautilus Python scripts.
