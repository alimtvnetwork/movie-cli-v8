# Project Memory

## Core
Go 1.22 CLI project (NOT web). Binary: `movie`. Ignore Lovable build errors.
One file per command, max ~200 lines. Shared helpers in movie_info.go and movie_resolve.go.
File naming: `01-name-of-file.md`. Keep folder file counts small.
Plans & suggestions tracked in single files, not per-item files.
Never modify `.release` folder. Any code change bumps at least minor version.
ALWAYS bump version/version.go after every code change. Never forget.
Malaysia timezone (UTC+8) for timestamps. No milestone file (readm.txt removed).
Root spec files: lowercase (spec.md, ai-handoff.md, development-log.md). Keep README.md uppercase.
Spec resequenced: foundation 01-06, app at 08, app-issues at 09. Issues in spec/09-app-issues/.
Error spec flattened: spec/02-error-manage-spec/ (no nested subfolder).
HTML JS: single API_BASE variable for all REST calls. Never repeat URL.
Boolean names: never use negative words (un/not/no). Use positive semantic synonyms with Is/Has prefix.
Zero nested if. Max 2 conditions per if. No else after return. Functions ≤15 lines. Files ≤300 lines. Max 3 params.
No magic strings — use constants/enums. No fmt.Errorf — use apperror.Wrap().

## Memories
- [Project overview](mem://01-project-overview) — Go CLI, command tree, architecture, file structure
- [Conventions](mem://02-conventions) — Code style, naming, build, deploy, config keys
- [Plan](mem://workflow/01-plan) — Done/pending task tracker, prioritized backlog
- [AI success plan](mem://workflow/01-ai-success-plan) — 7 rules for 98% AI success rate
- [Suggestions](mem://suggestions/01-suggestions) — Active suggestion tracker with priority levels
- [Reliability report](mem://reports/01-reliability-risk-report) — Failure map, corrective actions, readiness decision
- [Guideline violations audit](mem://audit/01-guideline-violations) — Full audit: nested ifs, magic strings, oversized funcs/files, 7-phase fix plan
- [Version bump rule](mem://preferences/version-bump) — Always bump version after every code change
- [API base variable](mem://preferences/api-base-variable) — JS must use single API_BASE variable, never repeat URL
- [Boolean naming](mem://constraints/boolean-no-negative-words) — IsUndone→IsReverted; never use un/not/no in boolean names
- [Timestamp bug](mem://issues/01-timestamp-bug) — Fixed: hardcoded "now" → RFC3339
- [Duplicate TMDb fetch](mem://issues/02-duplicate-tmdb-fetch) — Fixed: shared helpers
- [Large files](mem://issues/03-large-files) — Fixed: split to <200 lines
- [CI log commit loop](mem://issues/04-ci-log-commit-loop) — Constraint: CI log commits must never trigger new runs; kill feature if loops occur
- [Director FK 787](mem://issues/05-director-fk-stale-lastinsertid) — Fixed v2.304.0: ensureDirector trusted LastInsertId after INSERT OR IGNORE → stale rowid → MediaDirector FK 787; now SELECTs canonical PK like EnsureGenre
- [Season/Episode audit](mem://issues/06-season-episode-stale-lastinsertid) — Fixed v2.305.0: InsertSeason/InsertEpisode had same latent bug; now SELECT canonical PK. Rule: any INSERT OR IGNORE / ON CONFLICT DO UPDATE whose PK is used as FK MUST SELECT PK back, never trust LastInsertId.
- [LastInsertId lint guard](mem://features/lastinsertid-lint-guard) — v2.306.0 CI guard scripts/check-lastinsertid-anti-pattern.sh fails if any db/ file mixes INSERT OR IGNORE / ON CONFLICT with LastInsertId() in real code; comment-aware; wired into ci.yml lint job
- [Parallel scan](mem://features/parallel-scan) — Worker pool (NumCPU*2, cap 32), TMDb 40 req/s limiter, batch progress, auto-open report.html
- [Context menu](mem://features/context-menu) — `movie add-contextmenu` for Windows/Linux/macOS submenu (Scan/Rescan/Report/Stats), clicks logged to action_history
- [Remove/Move/SmartRescan plan](mem://features/remove-move-rescan-plan) — Spec at spec/08-app/10-remove-move-rescan/; 6 phases mapped onto existing Media schema (no parallel Movie table)
- [Global JSON cache](mem://features/global-json-cache) — Per-user TmdbID-keyed mirror at ~/.movie/cache/json; survives folder moves; disable via MOVIE_NO_GLOBAL_CACHE=1
