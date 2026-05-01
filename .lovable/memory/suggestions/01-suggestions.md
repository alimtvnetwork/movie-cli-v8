# Suggestions Tracker

> **Last Updated**: 01-May-2026

## Status Legend
- ✅ Done — implemented and verified
- 🔲 Open — not started

---

## ✅ Completed

| # | Suggestion | Completed | Notes |
|---|-----------|-----------|-------|
| S01 | Fix timestamp bug in move-log.json | 17-Mar-2026 | Replaced `"now"` with `time.Now().Format(time.RFC3339)` |
| S02 | Refactor large files (>200 lines) | 17-Mar-2026 | Split `movie_move.go` and `db/sqlite.go` |
| S03 | Extract shared TMDb fetch logic | 17-Mar-2026 | `fetchMovieDetails()`/`fetchTVDetails()` in `movie_info.go` |
| S04 | Cross-drive move fallback (copy+delete) | 05-Apr-2026 | `MoveFile()` detects EXDEV, falls back to copy+remove |
| S05 | Add confirmation prompt to `movie undo` | 10-Apr-2026 | Already implemented with `[y/N]` prompt |
| S06 | Add GIVEN/WHEN/THEN acceptance criteria | 10-Apr-2026 | 16 ACs covering all commands + export + batch move |
| S07 | Document shared helper locations | 10-Apr-2026 | Annotated movie_info.go, movie_resolve.go, movie_move_helpers.go, movie_scan_json.go |
| S08 | Clarify `movie ls` filter rule | 09-Apr-2026 | Only file-backed (scanned) items shown |
| S09 | Implement `movie tag` command | 06-Apr-2026 | `cmd/movie_tag.go` + `db/tags.go` |
| S10 | Add file size stats to `movie stats` | 10-Apr-2026 | Total, largest, smallest, average |
| S11 | Add error handling spec | 10-Apr-2026 | TMDb rate limits, DB locks, offline mode, filesystem errors |
| S12 | Update README.md with full docs | 10-Apr-2026 | 620+ lines, all commands, install, build |
| S13 | Batch move (`--all` flag) | 09-Apr-2026 | Move all video files from source at once |
| S14 | JSON metadata per movie/TV on scan | 09-Apr-2026 | `cmd/movie_scan_json.go` |
| S15 | Use `DiscoverByGenre` in suggest | 09-Apr-2026 | Genre-based discovery integrated |
| S16 | CI pipeline (lint, test, vuln scan) | 10-Apr-2026 | ci.yml + vulncheck.yml |
| S17 | Retry logic with exponential backoff | 11-Apr-2026 | 429 rate-limit handling, 3 retries |
| S18 | Add `movie duplicates` command | 10-Apr-2026 | Detect by TMDb ID, filename, or size |
| S19 | Add `movie cleanup` command | 10-Apr-2026 | Find/remove stale DB entries |
| S20 | Integration tests with SQLite fixtures | 11-Apr-2026 | db/db_test.go + db/testhelper_test.go |
| S21 | Apply error log spec v2 to ci.yml | 10-Apr-2026 | Per-stage error logs, summary assembly |
| S22 | Add `movie watch` / watchlist | 11-Apr-2026 | to-watch/watched tracking |
| S23 | Console-safe updater handoff | 16-Apr-2026 | Synchronous execution, exit code propagation, gitmap pattern |
| S24 | Guideline violations audit | 16-Apr-2026 | 280+ violations catalogued, 7-phase remediation plan |
| S25 | Nested-if refactoring (top 20 files) | 16-Apr-2026 | Early returns, guard clauses, extracted helpers |
| S26 | Magic strings → constants | 24-Apr-2026 | `cmd/constants.go` holds shared messages incl. `msgDatabaseError` |
| S27 | fmt.Errorf → apperror.Wrap() | 24-Apr-2026 | Only remaining `fmt.Errorf` is inside `apperror/apperror.go` (correct) |
| S28 | Oversized functions split | 24-Apr-2026 | Phase 5 complete — all funcs ≤15 lines |
| S29 | Oversized files split | 24-Apr-2026 | Phase 6/7 complete — all files <300 lines |
| S30 | Single PascalCase DB | 24-Apr-2026 | 21 tables + 8 views in `db/schema_tables.go`, `db/views.go` |
| S31 | Migration runner + SchemaVersion | 24-Apr-2026 | `db/migrate.go`, `db/schema_version.go`, 3 versioned migrations |
| S32 | FileAction seed (15 actions) | 24-Apr-2026 | `db/seed.go` seeds Move/Rename/Delete/Popout/Restore/etc. |
| S33 | REST API server mode | shipped earlier | `cmd/movie_rest*.go` — HTML dashboard + JSON endpoints. Tracker was stale. |
| S34 | Remove/Move/SmartRescan epic | 01-May-2026 | M5 atomic move + rollback, --purge, auto report regen, global JSON cache. |
| S35 | Acceptance criteria docs | 01-May-2026 | spec/08-app/97-acceptance-criteria.md v1.3.0 — added AC-36..AC-40 covering rm/--purge, atomic move, history reconcile, contextmenu+telemetry, global cache. |

---

## 🔲 Open — Priority Order

| # | Suggestion | Priority | Description |
|---|-----------|----------|-------------|
| — | (none open — backlog drained) | — | Add new suggestions here as they arise. |

## 🚫 Forbidden — Do NOT Suggest

- ❌ **Watchlist TMDb account sync (pull/push)** — user banned. See `mem://constraints/no-tmdb-account-sync`. Local JSON `movie watch export/import` is the final design.

---

*Tracker updated: 01-May-2026*
