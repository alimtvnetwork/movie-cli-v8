# Completed Plan: Fix Movie Update, Doctor Self-Replace, and Path Synchronization

**Date:** 2026-09-20  
**Version:** v2.331.0  
**Status:** Completed & Verified  

---

## 1. Problem Statement & Symptoms

Running `movie update` on standard Windows installations (`%LOCALAPPDATA%\movie-cli\movie.exe`) triggered cascading preflight and auto-fix failures:
1. `[ERR ] Deploy path differs from active PATH binary`: `powershell.json` had hardcoded `deployPath: "D:\\bin-run"`, but `D:\bin-run\movie.exe` did not exist.
2. `[WARN] Deploy directory is NOT in PATH`: `D:\bin-run` was warned as missing from `$PATH`.
3. `[WARN] Version drift`: `doctor` failed attempting to inspect the version of the non-existent deploy binary.
4. `[ERR ] Source folder (ScanDir)`: `~/Downloads` failed with `CreateFile ~/Downloads: The system cannot find the path specified.` because Windows Go `os.Stat` does not expand `~`.
5. `[ERR ] self-replace failed`: `doctor --fix` (and `autoFixPostUpdate`) called `updater.SelfReplace("D:\\bin-run\\movie.exe", "C:\\Users\\...\\movie.exe")`, which crashed with `source binary not found: CreateFile D:\bin-run\movie.exe`.
6. Terminal glyphs rendered as replacement characters `` in `updater/run.go` because the console did not interpret multi-byte emojis properly.

---

## 2. Root Cause Analysis

1. **Missing Source Binary in `self-replace`**: `doctor` treated `powershell.json`'s `deployPath` as an authoritative source binary even when it did not exist on disk, blindly calling `updater.SelfReplace`.
2. **Missing Tilde Expansion**: `doctor/checks_env.go` checked `ScanDir` with raw `os.Stat(dir)` without expanding leading `~/` or `~\` via `os.UserHomeDir()`.
3. **Stale Deploy Path Drift**: `powershell.json` had a hardcoded obsolete developer path instead of using a portable default that resolves dynamically to the platform standard directory (`%LOCALAPPDATA%\movie-cli` on Windows, `~/.local/bin` on Unix).
4. **Emoji Glyphs on Console**: `updater/run.go` emitted multi-byte emojis (`🎯`, `🔄`, `✨`) which fail on older Windows consoles and non-UTF8 codepages.

---

## 3. Implementation Details

- **`doctor/paths.go`**:
  - `resolveDeploySource`: Defaults to `defaultDeployDir()` when `cfg.DeployPath` is empty.
- **`doctor/paths_save.go`**:
  - Added `SaveDeployPath(deployPath)`: Updates `powershell.json`, normalizing default deploy directories to empty string to prevent committing machine-specific absolute paths.
  - Added `hasSourceBinary(path)`: Safe stat check avoiding panic or false errors.
  - Added `defaultDeployDir()`: Cross-platform standard directory resolution (`%LOCALAPPDATA%\movie-cli` on Windows, `~/.local/bin` on Unix).
  - Added `expandScanPath(path)`: Safely expands leading `~/` or `~\` using `os.UserHomeDir()`.
- **`doctor/checks.go`**:
  - `checkPathMismatch`: When `report.Source` does not exist on disk, emits `SeverityWarn` with hint to sync `powershell.json` to the active binary instead of a fatal error.
  - `checkDeployInPath`: When `report.Source` does not exist on disk, validates whether the active binary directory (`filepath.Dir(report.Target)`) is in PATH.
  - `checkVersionDrift`: When `report.Source` does not exist on disk, reports `SeverityOK` for the active binary version instead of false version drift.
- **`doctor/checks_env.go`**:
  - `checkSourceFolder`: Calls `expandScanPath(dir)` before `os.Stat()`, properly resolving `~/Downloads` on Windows.
- **`doctor/fix.go`**:
  - `runSelfReplace`: When source binary does not exist on disk, calls `syncDeployPathToActive` to automatically update `powershell.json`'s `deployPath` to the active binary directory.
- **`powershell.json`**:
  - Updated default `deployPath` to `""` for portable runtime auto-resolution without hardcoded absolute paths.
- **`run.ps1`**:
  - Updated fallback deploy directory from obsolete `"E:\bin-run"` to platform standard (`%LOCALAPPDATA%\movie-cli` on Windows, `~/.local/bin` on Unix).
- **`cmd/terminal_windows.go`**:
  - Added `SetConsoleOutputCP(65001)` and `SetConsoleCP(65001)` calls in `initVirtualTerminal()` to ensure Windows console runs in UTF-8 codepage.
- **`updater/run.go`**:
  - Replaced multi-byte emojis with clean standard prefixes (`==>`, `[OK ]`, `Repo:`, `Target:`) avoiding console replacement characters.
- **`doctor/checks_test.go`**:
  - Comprehensive unit tests covering path mismatch, version drift, missing source handling, tilde expansion, and default deploy directories.

---

## 4. Verification Results

- Unit Tests: `go test -v ./doctor/... ./updater/... ./cmd/...` (All PASSED).
- CI/CD Quality Gates: `python 03-ai-scripts/06-cicd-local-runner.py --all-paths --run-tests` (20/20 gates PASSED 100% green).
- Linter: `golangci-lint run ./...` (0 errors).
- CLI Verification: `movie doctor` cleanly passed all checks and recognized `C:\Users\Administrator\Downloads` without errors.
