# Completed Plan: Improve Installer and Release Process Following GitMap

## User Request (Verbatim)

```text
improve the installer and release process follwoing

gitmap

D:\work\gitmap

# Release-Triggered CI/CD Fix Loop — Workflow (must follow)
```

---

## Execution Summary & Delivered Outcomes

Following the architectural patterns and implementation standards in `D:\work\gitmap`, the installer and release ecosystem of `movie-cli` has been overhauled from legacy source-compiling scripts into production-grade binary release installers, complete with quick one-liners, uninstaller suites, and release workflow dry-run contract verification.

### 1. Root Binary Release Installers Overhauled
- **`install.ps1` (Windows)**:
  - Completely replaced legacy source-compiling script with production binary release installer.
  - Downloads precompiled binaries (`movie-vX.Y.Z-windows-<arch>.zip`) directly from GitHub Releases (`alimtvnetwork/movie-cli-v8`).
  - Strict SHA-256 verification against `checksums.txt`.
  - Default install directory: `$env:LOCALAPPDATA\movie-cli` (with backward-compatible detection of legacy `$env:LOCALAPPDATA\movie`).
  - Persistent User `PATH` modification via Windows registry (`HKCU:\Environment`) and immediate current-session `$env:PATH` refresh.
  - Full CLI parameter matrix: `-Version`, `-InstallDir`, `-Arch`, `-NoPath`, `-Uninstall`, `-DryRun`, `-KeepData`, `-PurgeData`, `-Force`, `-NoDiscovery`, `-ProbeCeiling`.
  - Smart upgrade detection: detects existing installed version and reports upgrading vs reinstalling vs clean install.
  - Dry-run mode (`-DryRun`): outputs structured telemetry (`dryrun.*=`) and validates asset URL matching regex `^movie-v\d+\.\d+\.\d+-windows-(amd64|arm64)\.zip$` without disk modifications.
- **`install.sh` (Linux / macOS)**:
  - Completely replaced legacy source-compiling script with production POSIX binary installer.
  - Self re-executes under `bash` when piped via `curl -fsSL ... | sh`.
  - Platform detection supporting Linux, macOS (`darwin`), and Windows Git Bash / MSYS (`linux` fallback).
  - Architecture detection supporting `amd64` (`x86_64`) and `arm64` (`aarch64`).
  - Downloads `movie-vX.Y.Z-<os>-<arch>.tar.gz` and verifies SHA-256 checksums (`sha256sum` or macOS `shasum -a 256`).
  - Installs to `~/.local/bin` (XDG standard) and updates `.bashrc`, `.zshrc`, and `.profile` without duplicating PATH entries.
  - Full CLI flag matrix: `--version`, `--dir`, `--arch`, `--no-path`, `--dry-run`, `-n`, `--uninstall`, `--keep-data`, `--purge-data`, `--force`, `-f`.
  - Upgrade detection and structured dry-run reporting (`dryrun.*=`).

### 2. Quick Installers & Uninstallers Created
- **`install-quick.ps1`**:
  - Interactive / non-interactive quick installer for Windows.
  - Prompts for installation directory and architecture override when run interactively, or passes arguments cleanly to `install.ps1`.
- **`install-quick.sh`**:
  - Interactive / non-interactive quick installer for Linux / macOS.
  - Automatically invokes `install.sh` with defaults or specified options.
- **`uninstall-quick.ps1`**:
  - Clean uninstaller for Windows.
  - Removes binary and wrapper directories (`$env:LOCALAPPDATA\movie-cli` and legacy `$env:LOCALAPPDATA\movie`).
  - Cleans user `PATH` registry and current session `$env:PATH`.
  - Prompts before purging `~/.movie` database unless `-PurgeData` or `-Force` is provided.
- **`uninstall-quick.sh`**:
  - Clean uninstaller for Linux / macOS.
  - Removes binary from `~/.local/bin`, cleans shell profile PATH lines (`.bashrc`, `.zshrc`, `.profile`), and prompts for `~/.movie` cleanup.

### 3. Release Workflow & Contract Enhancement (`.github/workflows/release.yml`)
- **Version-Specific Script Generation**:
  - Enhanced `dist/install.ps1` generation with full parameter support (`-InstallDir`, `-Arch`, `-NoPath`, `-DryRun`, `-Force`), upgrade detection, and dry-run reporting.
  - Enhanced `dist/install.sh` generation with full flag support (`--version`, `--dir`, `--arch`, `--no-path`, `--dry-run`, `--force`), upgrade detection, and dry-run reporting.
- **Release Asset Staging**:
  - Automatically stages `install-quick.ps1`, `install-quick.sh`, `uninstall-quick.ps1`, and `uninstall-quick.sh` into `dist/` so they are attached directly to GitHub Releases.
- **Hard Gate Dry-Run Contract Validation**:
  - Added CI step `Validate generated install scripts contract via dry-run`:
    - Validates `dist/install.sh --dry-run` and `dist/install.ps1 -DryRun`.
    - Validates root `install.sh --dry-run` and `install.ps1 -DryRun`.
    - Validates syntax of `dist/install-quick.*` and `dist/uninstall-quick.*`.
- **Release Notes Documentation**:
  - Prominently embeds:
    - Quick Install (Windows PowerShell & Linux/macOS Bash)
    - Interactive Quick Install one-liners
    - Generic & Version-Pinned Install commands from `main`
    - Clean Uninstaller one-liners

---

## Verification Results

1. **Powershell Dry-Run Validation:**
   ```powershell
   pwsh -NoProfile -ExecutionPolicy Bypass -File .\install.ps1 -DryRun -Version v2.324.0
   # Exit code: 0
   # dryrun.name_matches_contract=True
   # OK install.ps1 dry-run passed for v2.324.0 (amd64)
   ```

2. **POSIX Bash Dry-Run Validation:**
   ```bash
   ./install.sh --dry-run --version v2.324.0
   # Exit code: 0
   # OK install.sh dry-run passed for v2.324.0 (linux/amd64)
   ```

3. **Syntax Checks:**
   - `bash -n install.sh install-quick.sh uninstall-quick.sh`: PASSED
   - `pwsh Get-Command -Syntax install.ps1, install-quick.ps1, uninstall-quick.ps1`: PASSED

4. **18-Gate Local CI Runner:**
   ```bash
   python 03-ai-scripts/06-cicd-local-runner.py --all-paths --run-tests
   # Output: "🎉 All quality gates passed successfully! Codebase is 100% green."
   # Exit code: 0
   ```

5. **Version Synchronization:**
   ```bash
   python 03-ai-scripts/14-version-sync-checker.py
   # Output: "✅ Version synchronization verified: v2.324.0 (0.80ms)"
   # Exit code: 0
   ```
