# Completed Plan: Advanced Installer & Release Process Parity (GitMap Standards)

- **ID:** 16-advanced-installer-and-release-process-parity
- **Status:** COMPLETED
- **Completed Date:** 2026-09-20
- **Release Version:** v2.327.0
- **Total Self-Loop Steps:** 4 steps

## Overview & Scope

Achieved full advanced feature parity with **GitMap** (`D:\work\gitmap`) across quick installers, uninstallation lifecycles, and automated static installer verification:
1. Dual-mode execution in `install-quick.sh` (eval-mode auto-sourcing + pipe-mode banner) per GitMap Spec 108.
2. Versioned repo discovery in `install-quick.ps1` and `install-quick.sh` per GitMap Spec 95.
3. Deploy path persistence (`powershell.json` with `deployPath`) in `install-quick.ps1`.
4. Native `movie self-uninstall` (alias `movie uninstall`) in Go CLI with Windows temp-handoff self-deletion per GitMap Spec 90.
5. Modernized static installer smoke tester (`03-ai-scripts/16-installer-smoke-tester.py`) registered as 20th quality gate in local runner matrix.

## Deliverables & Accomplishments

1. **`install-quick.sh` (Dual-Mode Execution & Repo Discovery):**
   - **Eval-Mode (`eval "$(curl -fsSL .../install-quick.sh)"`):** Executes inside user's interactive shell and auto-sources the updated shell profile (`~/.zshrc` / `~/.bashrc`) so `movie` is immediately callable on the very next command.
   - **Pipe-Mode (`curl -fsSL .../install-quick.sh | bash`):** Detects non-interactive child process and displays high-contrast post-install reload instructions.
   - **Subshell Isolation:** Heavy installation tasks run in a subshell `( ... )` so `set -eu` and environment variables never pollute the user session.
   - **Corrupted Directory Cleaner & Path Sanitizer:** Cleans accidental literal `~` directories and ANSI escape directories.
   - **Versioned Repo Discovery:** Probes `-v<N+1>` sibling repos and delegates if newer repository exists.

2. **`install-quick.ps1` (Versioned Repo Discovery & Deploy Path):**
   - **RunspacePool Concurrency:** Concurrently probes `-v<N+1>`..`-v<N+20>` sibling repos using parallel HEAD requests.
   - **Deploy Path Persistence:** Saves `powershell.json` in the target folder to store `deployPath` for `run.ps1` and toolchain discovery.
   - **Invoke-Safe Error Trapping:** Captures exceptions with step-level logging and summary reports.

3. **`movie self-uninstall` Command (`cmd/movie_self_uninstall.go`):**
   - Implemented native uninstallation in Go CLI (with alias `movie uninstall`).
   - Windows Handoff: Copies running binary to `%TEMP%\movie-handoff-<pid>.exe` and schedules detached deletion via `cmd.exe /C ping ... & del /F /Q "<target>"` to bypass Windows file locks.
   - Unix: Direct binary unlink via `os.Remove`.
   - Cleans User PATH entries and `~/.movie` data (unless `--keep-data`).
   - Integrated into `cmd/root.go` and `scripts/gen-command-index.py`.

4. **Static Installer Smoke Linter (`03-ai-scripts/16-installer-smoke-tester.py`):**
   - Differentiates direct installers from delegators and uninstallers.
   - Registered in `03-ai-scripts/02-shared-engine.py` as `"Static Installer Smoke Linter"`.
   - All 7 discovered installer scripts pass 100% green.

## Verification Outcomes

- Local Quality Gates: 20/20 checks passed (100% green).
- Command Index: 71 unit tests passed, 61 commands indexed.
- Go packages: 100% passing tests and clean `go vet`.
