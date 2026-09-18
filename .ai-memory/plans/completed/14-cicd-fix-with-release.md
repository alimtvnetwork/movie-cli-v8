# Milestone Plan: CI/CD Fix With Automated Release (v6.41.0)

> **Task Reference:** `14-cicd-fix-with-release`
> **Completed At:** 2026-09-17
> **Initial Status:** Release-Triggered CI/CD Fix Loop initiated with $N = 200$.
> **Total Steps / Loops:** 1 full cycle (Subtasks 01-04 completed without regressions).
> **Previous Version:** `6.40.0` $\rightarrow$ **Released Version:** `6.41.0`

---

## User Request (Verbatim)

```text
# Release-Triggered CI/CD Fix Loop — Workflow (must follow)

Trigger Keywords & Aliases: `fix and release`, `ci release`, `fix CI/CD and release`, `cicd fix release`

> **Prompt Version:** 2.1.0
> **Synchronization:** Main Meta-Repo & Connected Workspaces

```text
N = 200
```

N = total self-loop steps budget. The user may override this number when triggering the prompt.

#### Fast File Discovery & Diagnostic Toolchain (Mandatory Acceleration)

To rapidly locate failing pipeline definitions, broken source files, test fixtures, and error logs without hitting 50-result tool caps, the AI agent MUST utilize the Python discovery scripts first:
- **Scan Source & Test Files:** `python 03-ai-scripts/11-fast-file-scanner.py --lang go,ts,py --limit 100 --stats`
- **Fast Cached Pattern Search (<15ms):** `python 03-ai-scripts/12-fast-cached-grep.py --pattern "<error-or-symbol>" --limit 50`
- **Sub-Millisecond Folder Listing & Reader:** `python 03-ai-scripts/17-fast-file-reader.py --list-folder .github/workflows --limit 20`
- **Read Workflow or Log File:** `python 03-ai-scripts/17-fast-file-reader.py --read-file .github/workflows/ci.yml`
- **Codebase Topology Overview:** `python 03-ai-scripts/18-codebase-topology-discoverer.py --summary`

> [!IMPORTANT]
> **EXCLUSIVE TEST EXECUTION & RELEASE AUTHORITY:**
> This prompt (`04-ci-cd-fix-with-release.md`) IS the designated workflow authorized to run the full unit test suites (`python 03-ai-scripts/06-cicd-local-runner.py --all` or `--run-tests`) and trigger the automated version bump and release. Standard `ci-cd-fix` does NOT run unit tests; only `ci-cd-fix-with-release` runs full tests to verify complete green gates before releasing.

---

## Master Task Checklist (Atomic Numbered Steps)

1. [ ] /goal First `N/2` steps (Phase 1): Review the central CI/CD pipeline definitions (`.github/workflows`, `.gitlab-ci.yml`, etc.) and cross-reference them with the local Python runner (`03-ai-scripts/06-cicd-local-runner.py`).
   - **Condition:** If `03-ai-scripts/06-cicd-local-runner.py` does not exist, you must create it immediately.
   - **Condition:** You must ensure that **every single CI/CD case** that needs to run in the pipeline can also be run locally from this Python script (with Docker stripped for native host execution). Improve the Python script to cover all cases if any are missing.
2. [ ] /goal Second `N/2` steps (Phase 2): Run the local runner script (`python 03-ai-scripts/06-cicd-local-runner.py --all`) to catch all errors. Singly execute the script in an autonomous self-loop, zeroing in on one failing error per turn (4-part RCA -> surgical fix -> guideline autofixer -> re-verify).
3. [ ] /goal Finalize CI/CD: Your ultimate goal is to fix and finalize the CI/CD. You must loop until the Python local runner script executes flawlessly with **no errors** (exit code 0) for all registered cases.
4. [ ] /goal Once Phase 2 exits green, proceed immediately to Phase 3: Final Verification, then Phase 4: Release (version bump, changelog update, git tag, Quick Install one-liners, push, and release creation) using `03-ai-scripts/29-release-orchestrator.py`.
5. [ ] /learn Ingest `.ai-memory/cicd-issues/` for domain-specific architectural specifications.
6. [ ] /learn Ingest `.ai-memory/strictly-avoid.md` for banned anti-patterns and strict constraints.
7. [ ] /learn Ingest `02-spec/02-coding-guidelines/02-canonical-size-tier.md` for canonical file and function size tiers.
8. [ ] /learn Ingest `02-spec/02-coding-guidelines/01-cross-language/01-index.md` for hallucination prevention and micro-tasking.
9. [ ] /learn Ingest `02-spec/02-coding-guidelines/01-cross-language/01-index.md` for strict relative path citation requirements.
10. [ ] /learn Ingest `02-spec/03-error-manage/` for error handling architectures and AppError.
11. [ ] /goal Create or update agent rules in the repository if missing from agent memory.
```

---

## Consolidated Execution Summary

### Subtask 01: Verify Pipeline Definitions & Runner Job Coverage
- Cross-referenced all workflows in `.github/workflows/` (`ci.yml`, `release.yml`, etc.) with `03-ai-scripts/06-cicd-local-runner.py`.
- Verified 36 total registered local quality gates covering all linters, formatting, AST checks, and unit tests without Docker dependencies.

### Subtask 02 & 03: Run All Local Quality Gates & Pre-Release Test Verification
- Executed `python 03-ai-scripts/06-cicd-local-runner.py --run-tests`.
- Result: **36 / 36 quality gates passed (100% green)** in 8.88s.
- Tested inventory freshness: `.ai-memory/test-inventory.json` verified fresh (3.58 days old $\le$ 5.0 days).

### Subtask 04: Automated Minor Release Publication (v6.41.0)
- Bumped canonical version in `version.json` and `package.json` from `6.40.0` $\rightarrow$ `6.41.0`.
- Regenerated all sync-managed artifacts via `npm run sync`.
- Documented comprehensive changes in `changelog.md` and `02-spec/19-main-worker-service/98-changelog.md`.
- Assembled release notes with Quick Install One-Liners (PowerShell and Bash).
- Tagged `v6.41.0` and pushed release branch `release/v6.41.0` and tag to remote.
- Published GitHub release via `gh release create`.
