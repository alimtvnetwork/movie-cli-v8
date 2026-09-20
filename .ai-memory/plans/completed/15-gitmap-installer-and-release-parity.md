# Completed Plan: GitMap Installer & Release Process Parity Overhaul

- **ID:** 15-gitmap-installer-and-release-parity
- **Status:** COMPLETED
- **Completed Date:** 2026-09-20
- **Release Version:** v2.326.0

## Overview & Scope

Achieved full feature parity with GitMap (`d:\work\gitmap`) across repository installers, init scripts, run companions, deployment manifests, release packaging, and cross-platform dry-run smoke testing.

## Deliverables & Accomplishments

1. **Deploy Manifest Architecture (`deploy-manifest.json`):**
   - Created root `deploy-manifest.json` specifying `appSubdir: "movie-cli"`, `binaryName: {"windows": "movie.exe", "unix": "movie"}`, and `legacyAppSubdirs: ["movie"]`.
   - Wired `Load-DeployManifest` in `install.ps1` and `load_deploy_manifest` in `install.sh` to read configuration dynamically with fallback defaults.

2. **One-Shot Repository Initialization (`init.ps1`, `init.sh`):**
   - Created root `init.ps1` (PowerShell) and `init.sh` (POSIX Bash) running environment preflight checks (Go, Git, Python), permission fixes, and installer contract dry-run validation.

3. **POSIX Run Companion (`run.sh`):**
   - Created root `run.sh` matching `run.ps1` feature-for-feature: `--no-pull`, `--force-pull`, `--no-deploy`, `--deploy-path`, `-r`/`--run`, and `-t`/`--test`.

4. **Installer Subsystem Enhancements:**
   - `install.ps1`: Configured Win32 UTF-8 Console CP (`SetConsoleOutputCP(65001)` / `SetConsoleCP(65001)`) to avoid mojibake on Windows terminals, and wired `Load-DeployManifest`.
   - `install.sh`: Dynamic `load_deploy_manifest` integration, dual-shell PATH detection, and smart reload hint.
   - `install-quick.ps1`: Added `Invoke-Safe` step wrapper and `$script:InstallErrors` list tracking.
   - `uninstall-quick.ps1`: Added `Try-SelfUninstall` checking for `movie` on PATH to invoke `movie uninstall -y` before falling back to manual disk sweep.
   - `uninstall-quick.sh`: Added `Try-SelfUninstall` check before manual sweep fallback.

5. **Cross-Platform Installer Smoke Test Suite (`.github/scripts/smoke-installer.py`):**
   - Created standalone script supporting `source` mode (compiles and asserts binary version) and `release`/`dryrun` mode (validates release artifact contracts).
   - Added `sys.stdout.reconfigure(encoding="utf-8", errors="replace")` for Windows console UTF-8 resilience.

6. **CI/CD Quality Gates & Release Workflow Integration:**
   - Added `"Installer Smoke Dry-Run"` to local runner matrix (`03-ai-scripts/02-shared-engine.py`).
   - Verified all 19 local CI/CD quality gates passed 100% green (`03-ai-scripts/06-cicd-local-runner.py --all-paths --run-tests`).
   - Enhanced `.github/workflows/release.yml` with `deploy-manifest.json` staging to `dist/`, pre-release dry-run contract check, and `Post-release installer smoke contract assert`.
   - Synchronized all new installer files into `03-ai-scripts/29-release-orchestrator.py`.

## Verification Outcomes

- Local CI/CD: 19/19 checks passed (exit code 0).
- Go tests: 100% pass across all packages (`db`, `cleaner`, `history`, `cmd`, `pkg/*`).
- Smoke runner: `python .github/scripts/smoke-installer.py dryrun` passed with exit code 0.
