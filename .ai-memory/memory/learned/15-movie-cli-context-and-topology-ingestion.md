# Movie CLI v8 Context, Topology & Memory Ingestion

> **Type:** Institutional Knowledge & Learned Memory  
> **Status:** Active & Canonical  
> **Date:** 2026-09-19  
> **Reference Prompt:** `01-prompts/03-read-write/02-read-memory-enhanced.md`  

---

## 1. Repository Identity & Architecture

- **Canonical Repository:** `alimtvnetwork/movie-cli-v8` (Go 1.22 CLI application, SQLite database engine, Cobra command suite).
- **Core Stacks:**
  - **Go:** Core application logic located in root and subdirectories (`cmd/`, `cleaner/`, `tmdb/`, `db/`, `config/`, `organizer/`, `classifier/`, `scanner/`, `watchlist/`, `stats/`, `tagger/`, `version/`, `web/`, `pkg/appfault/`).
  - **Database:** SQLite in WAL mode (`movie.db`), located under `<binary-dir>/data/movie.db`. PascalCase table names, `{Table}Id` primary keys (autoincrement), positive booleans (`Is*`, `Has*`).
  - **CLI Architecture:** Cobra framework providing 12 canonical commands: `scan`, `ls`, `search`, `info`, `suggest`, `move`, `rename`, `undo`, `play`, `stats`, `tag`, `config`.
  - **Tooling & Automation:** 35 Python helper scripts in `03-ai-scripts/` utilizing shared caching engine (`03-ai-scripts/02-shared-engine.py`).
  - **UI / Frontend:** Embedded web assets (React 18 + Vite).

---

## 2. Git Commit History Audit (Last 10 Commits)

1. `ad9ffb1` — `feat(prompts): sync latest prompts and skills from coding-guidelines`: Synchronized prompt library and skills from coding-guidelines repository.
2. `6c6712f` — `docs: update readme and what-to-read with new files`: Updated documentation links, readme references, and reading priority index.
3. `e2c6112` — `refactor(core): implement appfault and boolean dry optimizations`: Unified Go structured errors to `pkg/appfault` and eliminated explicit boolean comparisons.
4. `34a99d4` — `chore(core): migrate AI memory structure and draft DRY optimization plan`: Migrated legacy `.lovable` to `.ai-memory`, restructured specifications into `02-spec/`, and cataloged `03-ai-scripts/`.
5. `384beda` — `Fixed v6 vs v8 CI guard`: Guarded workflows to ensure consistency with `movie-cli-v8` naming.
6. `033dabe` — `Changes`: Version info updates in `version/info.go`.
7. `e263385` — `Changes`: Workflow fixes in `.github/workflows/ci.yml`.
8. `4e53010` — `Release v2.322.0`: Tagged release point for v2.322.0.
9. `0d3499f` — `Fixed guard to ban v6 refs`: CI workflow rule prohibiting legacy v6 references.
10. `7760631` — `Changes`: CI adjustments in `.github/workflows/ci.yml`.

---

## 3. Auto-Healing & Repository Hygiene

- **Lowercase Readme Auto-Healing:** Renamed legacy uppercase `README.md` to strictly lowercase `readme.md` using two-step git rename (`git mv README.md readme_temp.md && git mv readme_temp.md readme.md`).
- **Version Metadata:** Root `version.json` noted as missing at repository root; versioning currently anchored in `version/info.go` and `package.json`.
- **Git Ignore Hardening:** Ignored `.gitmap/backup/` to maintain clean git status and prevent ephemeral artifact leakage.

---

## 4. Database Schema & Conventions

- **Engine:** SQLite 3 with Write-Ahead Logging (WAL) and `PRAGMA foreign_keys = ON`.
- **Schema Contracts:**
  - Tables: `Media`, `Genre`, `MediaGenre`, `Cast`, `MediaCast`, `Language`, `FileAction`, `ScanFolder`, `ScanHistory`, `MoveHistory`, `ActionHistory`, `Tag`, `Watchlist`, `Config`, `ErrorLog`.
  - Primary Keys: Integer autoincrement named `{TableName}Id` (e.g., `MediaId`, `GenreId`).
  - Booleans: `INTEGER NOT NULL DEFAULT 0`, named with positive prefixes (`IsAdult`, `HasSubtitles`, `IsWatched`).
  - Timestamps: ISO-8601 UTC strings (`CreatedAt`, `UpdatedAt`).
  - Text Fields: Nullable `Description TEXT NULL` for entity tables; nullable `Notes TEXT NULL` and `Comments TEXT NULL` for transaction tables.

---

## 5. CODE RED Rules & Strict Avoidances

1. **No Explicit True Checks (TOTAL BAN):** Never evaluate booleans explicitly against `true` or `false` (`if isReady == true`). Evaluate implicitly: `if isReady`.
2. **No Mixed Polarity:** Never combine positive and negative checks in the same condition (`if isA && !isB`).
3. **Never Disable CI/CD:** Total ban on commenting out, bypassing, or deleting CI/CD steps or linters.
4. **Strict Lowercase Naming:** All files and directories must use strictly lowercase naming (e.g. `readme.md`, `skill.md`).
5. **Strict Relative Git Paths:** Total ban on absolute filesystem paths (`file:///`, `C:\...`, `/home/...`).
6. **Go Structured Error Handling:** All Go packages must return `*appfault.AppError` or monadic `Result[T]`. Never swallow errors.
7. **Zero-Storage GitHub Actions:** Total ban on `actions/upload-artifact` in routine CI workflows.
8. **No Tests Without Owner Command:** Total ban on running unit tests in routine tasks unless explicitly commanded by repository owner (`--no-tests` flag required).

---

## 6. Pending Plans & Ambiguities Register

- **Active Pending Plans (5):**
  1. `pending/02-slides-system-overhaul.md` — Slides system overhaul.
  2. `pending/04-guideline-prompt-and-installer-upgrade.md` — Guideline prompt and installer enhancements.
  3. `pending/07-movie-cli-migration-and-optimization.md` — Migration and DRY optimization of `movie-cli-v8`.
  4. `pending/09-update-prompts-and-release.md` — Prompt updates and release lifecycle.
  5. `pending/11-code-red-refactor-remediation.md` — Remediate enum, boolean, and error wrapper violations.
- **Active Subtasks (4):**
  1. `plans/subtasks/01-slides-system-overhaul/01-consolidated-tasks.md`
  2. `plans/subtasks/03-guideline-prompt-and-installer-upgrade/01-consolidated-tasks.md`
  3. `plans/subtasks/07-movie-cli-migration-and-optimization/01-migration.md`
  4. `plans/subtasks/07-movie-cli-migration-and-optimization/02-dry-plan.md`
- **CI/CD Issues Absorbed (7):** Documented in `.ai-memory/cicd-issues/readme.md` (all 7 historical issues resolved; 0 active).
- **Open Ambiguities (1):** `.ai-memory/ambiguous-questions/01-new-ambiguity/01-pluggable-logger-backend-and-uber-zap-migration.md` (Pluggable logger backend vs Uber Zap migration).
- **Resolved Ambiguities (0):** None in `.ai-memory/ambiguous-questions/02-ambiguity-resolved/`.

---

## 7. Antigravity Skills & Rules Scaffolding

- **Skills Scaffolded in `.agents/skills/`:**
  - `coding-guidelines/skill.md`: Cross-language guideline enforcement.
  - `read-memory-enhanced/skill.md`: Context retrieval and repo ingestion runbook.
  - `execute-pending-tasks/skill.md`: Micro-tasking and self-loop execution engine.
  - `ci-cd-fix/skill.md`: Diagnostic pipeline remediation adhering to zero-storage rules.
  - `release-management/skill.md`: Semantic release and artifact publishing protocol.
- **Rules Scaffolded in `.agents/rules/`:**
  - `boolean-principles.md`: Positive boolean conventions and polarity bans.
  - `error-handling.md`: Structured `*appfault.AppError` and error context wrapping.
  - `zero-storage-ci.md`: Actions storage quota protection.
  - `coding-constraints.md`: File length, function bounds, and parameter structs.
