# 17 — GitMap CI/CD Runner Parity, Zero-Storage Workflows & Coding Guidelines Sync

- **Slug:** gitmap-cicd-runner-parity-and-guidelines-sync
- **Date:** 2026-09-22
- **Version:** v2.336.0
- **Category:** learned
- **Status:** permanent

---

## 1. Executive Summary

In response to synchronization requirements across the meta-repository fleet and continuous integration enhancements, `movie-cli-v8` underwent an overhaul across four primary dimensions:
1. **Local CI/CD Runner Parity:** `03-ai-scripts/06-cicd-local-runner.py` now natively supports GitMap priority shortcuts (`run-smart`, `smart`, `--smart`, `run-incremental`, `incremental`, `--fast`, `--commits <N>`). The smart runner dynamically detects modified Go packages across working trees and commit history, executing targeted dual-queue unit tests directly into isolated temp directories.
2. **Zero-Storage GitHub Actions Mandate (Rule 7):** Routine CI workflows (`.github/workflows/ci.yml`) were refactored to eliminate all instances of `actions/upload-artifact@v4` and `actions/download-artifact@v4`. Test outputs, coverage markers, legacy module-path audit reports, and cross-platform build validations now stream directly to `$GITHUB_STEP_SUMMARY` and stdout at 0 storage cost. A dedicated `.github/workflows/purge-actions-artifacts.yml` was added to maintain a 0.0 GB Actions footprint.
3. **Canonical Prompts & Specifications Sync:** Synced all canonical prompt suites (`01-prompts/`) from `coding-guidelines`, updating `16-ci-cd/` with the canonical 7-step sequence (`01-ci-cd-fix-tweak.md` through `07-cicd-pipeline-create.md`). Synced all 18 canonical specification suites in `02-spec/`, while strictly preserving `movie-cli` application-specific specs (`02-spec/21-app`, `22-app-issues`, `23-app-db`, `24-app-ui-design-system`, and `19-main-worker-service/98-changelog.md`).
4. **Antigravity Skills Fleet Expansion:** Imported 13 new system and GitMap subsystem skills (`gitmap`, `gitmap-developer-hygiene-and-agy`, `gitmap-macro-automation-engine`, `gitmap-pipeline-and-diagnostics`, `gitmap-scanner-and-cloner`, `gitmap-split-db-engine`, `gitmap-ssh-cluster-fleet`, `app-db-architecture`, `go-appwriter-and-result`, `python-dry-caching`, `react-ui-theming-design`, `slides-deck-management`, `spec-authoring-and-validation`) and updated all execution skills, while preserving all 10 `movie-cli-*` domain skills.

---

## 2. Key Architecture Details

### A. Smart Incremental Runner Engine
- **Command Entry Points:** Supports both positional commands (`python 03-ai-scripts/06-cicd-local-runner.py run-smart`) and flag triggers (`--smart`, `--changed-only`, `--pkg <target>`, `--fast`).
- **Dynamic Package Discovery:** Uses `resolve_git_changed_packages(commits=20)` to discover dirty files across `git status --porcelain`, `git diff --name-only HEAD~20 HEAD`, and `.ai-memory/temp/recent-file-changes.json`, mapping `.go` files to package roots (`cmd`, `db`, `tmdb`, etc.).
- **Dual-Queue Isolation:** Fast tests execute in chunks across 4 concurrent workers, while slow tests execute in dedicated batches. Passing tests produce zero filesystem artifacts and remain silent.

### B. Zero-Storage CI Reporting
- **Test Gate:** Outputs tabular test results and coverage profile status directly to `$GITHUB_STEP_SUMMARY`.
- **Legacy Path Auditor:** Emits formatted report block into `$GITHUB_STEP_SUMMARY` instead of archiving text artifacts.
- **Cross-Platform Builds:** Compiles all 6 OS/arch targets, computes SHA256 checksums and binary sizes, and logs validation to `$GITHUB_STEP_SUMMARY` without uploading non-release artifacts.

### C. Multi-Repo Sync Safety
- `03-ai-scripts/38-sync-prompts-skills-scripts.py` was fortified with `is_protected()` logic covering `movie-cli-`, `gitmap-`, and `21-app*` prefixes, guaranteeing that future automated syncs never purge repository-specific assets.
