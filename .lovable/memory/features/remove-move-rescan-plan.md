---
name: Remove · Move · Smart Rescan plan
description: Phased implementation plan for `movie rm` aliases, selector-mode `move`, and SmartRescan reconciliation; mapped onto existing Media schema (no parallel Movie table)
type: feature
---

# Remove · Move · Smart Rescan — Implementation Plan

Spec lives at `spec/08-app/10-remove-move-rescan/`. This plan tracks
phases. Each phase ends with a version bump and a verification step.

## Phase 1 — Schema migrations (DB only, zero behaviour change)
- `db/migrate_v4.go` — `MediaStatus` lookup + seed (Active/Removed/Moved/Missing); add `Media.IsDeleted` BOOL DEFAULT 0, `Media.MediaStatusId` INT DEFAULT 1.
- `db/migrate_v5.go` — `ReconciliationActionType` lookup + seed (HydratedFromJson/RemovedMissing/AddedNew/Converged); `ReconciliationHistory` table.
- `db/reconciliation_history.go` — Insert/Query helpers.
- Update `db/seed.go` for the two new lookups.
- Verify: open DB, run migrations, `psql/sqlite3 .tables` shows new tables.

## Phase 2 — Condition-expression engine (SHARED for rm + move)  ✅ DONE v2.289.0
- `cmd/movie_condition.go` — tokeniser + SQL builder + safety caps.
- `cmd/movie_condition_test.go` — table-driven grammar tests.
- No behaviour change in any existing command.

## Phase 3 — `movie rm` / `remove` / `delete`
- `cmd/movie_rm.go` (cobra root + aliases + help examples).
- `cmd/movie_rm_resolve.go` (delegate to `resolveMediaByQuery` or expr).
- `cmd/movie_rm_apply.go` (soft-delete + JSON unlink + audit + html regen).
- Wire into `root.go`. Verify undo restores via existing `FileActionDelete` handler.

## Phase 4 — `movie move <selector> <dest>` (selector mode)
- `cmd/movie_move_selector.go` (argc dispatcher + `-g` sugar expander).
- `cmd/movie_move_apply.go` (atomic batched mover + history rows + html regen).
- Existing interactive `movie move` (argc 0–1) untouched.

## Phase 5 — SmartRescan reconciliation
- `cmd/movie_scan_reconcile.go` (8-step orchestrator).
- `cmd/movie_scan_hydrate.go` (JSON → DB hydration, zero TMDb calls).
- `--no-reconcile` opt-out flag added to `movie scan`.
- Wire into `runMovieScan` BEFORE `executeScan`.

## Phase 6 — QA + docs + cleanup
- README section: rm/remove/delete examples + smart rescan note.
- Acceptance-criteria test pass on a real folder of 50+ video files.
- Bump to `v2.290.0` (minor for the feature set).

## Open decisions (deferred to future asks)
- `--purge` flag on rm to also delete the on-disk video file.
- Cross-folder global JSON cache (current scope = per-folder only).
