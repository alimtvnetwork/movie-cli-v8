# Project Plan & Status

> **Last Updated**: 01-May-2026

## ✅ Completed

### Core CLI Structure
- [x] Root command with Cobra (`movie-cli`)
- [x] `hello` command with version display
- [x] `version` command with ldflags injection
- [x] `self-update` → migrated to `update` command (gitmap console-safe handoff)

### Movie Management Commands
- [x] `movie config` — get/set configuration with masked API key display
- [x] `movie scan` — folder scanning with TMDb metadata + poster download
- [x] `movie ls` — paginated list with interactive navigation + detail view
- [x] `movie search` — live TMDb search, select, save to DB
- [x] `movie info` — local DB lookup → TMDb fallback → auto-persist
- [x] `movie suggest` — genre-based recommendations + trending fallback
- [x] `movie move` — interactive browse, move, track history (cross-drive support)
- [x] `movie rename` — batch clean rename with undo tracking
- [x] `movie undo` — revert last move/rename operation (with confirmation prompt)
- [x] `movie play` — open file with system default player (cross-platform)
- [x] `movie stats` — counts, genre chart, average ratings, file sizes
- [x] `movie tag` — add/remove/list tags
- [x] `movie export` — export library data
- [x] `movie duplicates` — detect by TMDb ID, filename, or size
- [x] `movie cleanup` — find/remove stale DB entries
- [x] `movie watch` — watchlist: to-watch/watched tracking

### Infrastructure
- [x] SQLite database with migrations (single movie.db)
- [x] TMDb API client (search, details, credits, recommendations, trending, posters, retry with backoff)
- [x] Filename cleaner (junk removal, year extraction, TV detection, slugs)
- [x] Makefile with build + cross-compile targets
- [x] build.ps1 / run.ps1 PowerShell deploy script
- [x] spec.md — full project specification
- [x] Shared resolver helper (`movie_resolve.go`)
- [x] apperror package for error wrapping
- [x] errlog package for structured error logging

### Bug Fixes & Refactoring
- [x] Fixed timestamp bug — `saveHistoryLog` now uses RFC3339
- [x] Deduplicated TMDb fetch logic — shared helpers
- [x] Split large files to <200 lines
- [x] Cross-drive move fallback (copy+delete for EXDEV)
- [x] Undo confirmation prompt
- [x] Console-safe updater handoff (sync, exit code propagation) ✅ 16-Apr-2026
- [x] Nested-if refactoring — top 20 files cleaned ✅ 16-Apr-2026
- [x] Guideline violations audit — 280+ violations catalogued ✅ 16-Apr-2026
- [x] Guideline Phase 3 — magic strings → constants (`cmd/constants.go`) ✅
- [x] Guideline Phase 4 — `fmt.Errorf` → `apperror.Wrap()` (only remaining is inside apperror itself) ✅
- [x] Guideline Phase 5 — oversized functions split ✅
- [x] Guideline Phase 6 — >3 params → option structs ✅
- [x] Guideline Phase 7 — final consistency pass, 0 violations ✅

### Documentation & CI
- [x] README.md, spec.md, ai-handoff.md, development-log.md
- [x] .lovable/memory structure
- [x] AI success rate plan
- [x] Reliability risk report
- [x] CI pipeline (lint, test, vuln scan)
- [x] Release pipeline with cross-compile + install scripts
- [x] Integration tests with SQLite fixtures

### Spec Restructuring (Phase 1-5) ✅
- [x] All 5 phases complete

### PowerShell Automation (Phase 1-8) ✅
- [x] All 8 phases complete

### Database Redesign v2.0.0 (15-Apr-2026) ✅
- [x] Schema diagram, design spec, state/history, popout, migration spec
- [x] Collection table, Tag M-N via MediaTag, 14 FileAction types
- [x] Removed Split DB — single movie.db

---

## ✅ Recently Completed (24-Apr-2026 audit)

### Database Implementation (P0) — DONE
- [x] PascalCase schema in Go (`db/schema_tables.go`) — 21 tables, single `movie.db`
- [x] SchemaVersion tracking + migration runner (`db/migrate.go`, `db/schema_version.go`)
- [x] FileAction seeded with 15 predefined rows (`db/seed.go`)
- [x] 8 database views created (`db/views.go`): VwMediaDetail, VwMediaGenreList,
      VwMediaCastList, VwMediaFull, VwMoveHistoryDetail, VwActionHistoryDetail,
      VwScanHistoryDetail, VwMediaTag
- [x] 3 versioned migrations registered (v1 initial, v2 ImdbLookupCache, v3 TmdbId+MediaType)

### Code Alignment (P1) — DONE
- [x] All commands use PascalCase column names
- [x] `movie_info.go` / `movie_resolve.go` aligned with new Media table

---

## 🔲 Pending — Prioritized Backlog

### Phase 3: Spec Completeness (P2)
- [x] Acceptance criteria (GIVEN/WHEN/THEN) for all commands ✅ 01-May-2026 (v1.3.0, AC-01..AC-40)
- [x] Shared helper docs — `// SHARED HELPER:` banners on 17 cmd/ helper files ✅ 01-May-2026 (v2.303.0)

### Phase 4: Future Enhancements (P3)
- [x] Director normalization table (`db/director.go`) ✅
- [x] Season/Episode tables for TV series (`db/season.go`) ✅
- [x] REST API server mode with HTML dashboard ✅ (`cmd/movie_rest*.go`)
- 🚫 **Watchlist TMDb account sync — FORBIDDEN.** See `mem://constraints/no-tmdb-account-sync`. Local JSON sync only.

---

## 🚫 Known Issues

- **Stale local repo** — RESOLVED in tracker. Action required on user's machine only:
  ```
  git fetch origin && git reset --hard origin/main && git clean -fd
  ```
  Cannot be executed by AI. See `.lovable/pending-issues/01-local-repo-stale.md`.

---

## Next Task Selection

All P0/P1/P2/P3 code work is complete. Only manual/user-side items remain:

1. **Real-OS QA** — run `spec/08-app/11-context-menu-qa-checklist.md` and 50+ file scan on user machine
2. **Resolve stale local repo** — user must run `git fetch origin && git reset --hard origin/main && git clean -fd`
