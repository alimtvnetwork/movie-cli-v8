# Movie CLI Codebase Deep Dive & Domain Skills Architecture

> **Type:** Institutional Knowledge & Learned Memory  
> **Status:** Active & Canonical  
> **Date:** 2026-09-21  
> **Reference:** Comprehensive codebase survey and domain skills scaffolding  

---

## 1. Executive Summary

This document records the end-to-end architecture, subsystems, command flows, and newly authored domain-specific Antigravity skills for `movie-cli-v8` (`github.com/alimtvnetwork/movie-cli-v8`). A specialized 7-part domain skills set has been scaffolded into `.agents/skills/`, empowering AI agents to make future changes, refactorings, and feature additions with zero hallucination.

---

## 2. Codebase Topology & Subsystem Map

| Subsystem | Core Packages / Paths | Primary Responsibility |
|---|---|---|
| **CLI & Commands** | `cmd/root.go`, `cmd/movie_*.go` | Cobra command definitions, flag binding, and terminal UI rendering. |
| **Filename Cleaner** | `cleaner/parse.go` | Token parsing, 30+ junk regexes, release groups, and TV format detection. |
| **Database Engine** | `db/*.go` (`open.go`, `schema_*.go`, `migrate_*.go`) | SQLite 3 WAL engine, PascalCase schema, v1–v7 migrations, and queries. |
| **TMDb Metadata** | `tmdb/*.go`, `db/imdb_lookup_cache.go` | API v3 client, rate limiter, DuckDuckGo/IMDb fallback, and cache. |
| **File Operations** | `cmd/movie_move*.go`, `cmd/movie_popout*.go`, `cmd/movie_rename.go` | Interactive/selector move, popout extraction, and batch renames. |
| **Undo / Redo** | `cmd/movie_undo*.go`, `cmd/movie_redo*.go` | Multi-level reversibility via `move_history` and `action_history`. |
| **Trash Bin** | `pkg/trashbin/*.go`, `cmd/movie_rm.go` | OS-native recycle bin integration (Windows COM, macOS, Linux). |
| **REST Server & UI** | `cmd/movie_rest*.go`, `src/*`, `templates/*` | Local HTTP API on port 8086, React 18 dashboard, and HTML reports. |
| **Updater & Doctor** | `updater/*.go`, `doctor/*.go` | Copy-and-handoff Windows update engine and diagnostic repairs. |
| **Context Menu** | `cmd/movie_contextmenu*.go` | OS file-manager right-click shell integration (registry, desktop). |
| **AI Tooling** | `03-ai-scripts/*.py` | 37 high-speed Python linters, code generators, and CI runners. |

---

## 3. Dedicated Domain Skills Catalog

The following skills have been authored in `.agents/skills/` adhering to strictly lowercase naming (`skill.md`) and relative git path mandates:

1. **`movie-cli-scanner-and-cleaner` (`.agents/skills/movie-cli-scanner-and-cleaner/skill.md`):**
   - Video extensions, regex junk patterns, TV season detection (`S01E02`), SmartRescan reconciliation, and `.movie-output/` generation.
2. **`movie-cli-database-architecture` (`.agents/skills/movie-cli-database-architecture/skill.md`):**
   - SQLite 3 WAL pragmas, PascalCase table DDL, `{Table}Id` PK conventions, v1–v7 migrations, and query interfaces.
3. **`movie-cli-tmdb-and-metadata` (`.agents/skills/movie-cli-tmdb-and-metadata/skill.md`):**
   - TMDb API v3 client, credentials resolution, rate-limiting backoff, DuckDuckGo/IMDb scraping fallback, and persistent lookup caching.
4. **`movie-cli-file-operations-and-history` (`.agents/skills/movie-cli-file-operations-and-history/skill.md`):**
   - Interactive & selector move modes, atomic batch rollbacks, popout `.temp/` compaction, batch renames, and scoped undo/redo.
5. **`movie-cli-rest-api-and-web-ui` (`.agents/skills/movie-cli-rest-api-and-web-ui/skill.md`):**
   - Local REST API on `:8086`, React 18 + Vite frontend structure, Tailwind/Radix components, and standalone HTML reports.
6. **`movie-cli-updater-and-doctor` (`.agents/skills/movie-cli-updater-and-doctor/skill.md`):**
   - 5-step copy-and-handoff updater bypassing Windows file locks, diagnostic health checks with `--fix`, and OS context menus.
7. **`movie-cli-testing-and-qa` (`.agents/skills/movie-cli-testing-and-qa/skill.md`):**
   - Test suites, centralized test inventory (`.ai-memory/test-inventory.json`), strict owner-command testing rules, and `t.TempDir()` isolation.

---

## 4. Key Developer Constraints & Guardrails

- **Boolean Standard:** All booleans must use `is` or `has` prefixes; total ban on `== true` explicit evaluations and mixed polarity (`if isA && !isB`).
- **Error Handling:** Go functions must return `*appfault.AppError` or `Result[T]` enriched with `.WithPath()` and `.WithVar()`; never swallow errors.
- **Testing Prohibition:** Never execute unit tests or CI runners during routine turns unless explicitly requested by the user (`--no-tests` required in CI scripts).
- **Zero-Storage CI:** GitHub Actions workflows must not upload artifacts via `actions/upload-artifact`.
