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

## Phase 3 — `movie rm` / `remove` / `delete`  ✅ DONE v2.291.0
- `cmd/movie_rm.go` (cobra root + aliases + auto-detected resolution mode).
- `cmd/movie_rm_apply.go` (preview + confirm + soft-delete + snapshot audit).
- `db/media_softdelete.go` (SoftDeleteMedia / RestoreMedia / QueryMediaIDsByWhere).
- `cmd/movie_undo_exec.go` patched: undoDelete restores existing soft-deleted row first.
- Wired into `root.go` as `movieRmCmd`.

## Phase 4 — `movie move <selector> <dest>` (selector mode)  ✅ DONE v2.292.0
- `cmd/movie_move_selector.go` (argc dispatcher + `-g` sugar + `-y` confirm skip).
- `cmd/movie_move.go` patched: argc relaxed to ≤2, dispatches by isSelectorMoveInvocation.
- Reuses BuildConditionSQL + resolveMediaByQuery; logs FileActionMove via InsertMoveHistory (undo handled by existing executeMoveUndo).
- Existing interactive `movie move` (argc 0–1) untouched.

## Phase 5 — SmartRescan reconciliation  ✅ DONE v2.293.0
- `cmd/movie_scan_reconcile.go` — 8-step orchestrator (disk/json/db diff).
- `cmd/movie_scan_hydrate.go` — JSON → DB, zero TMDb calls.
- `db/media_softdelete.go` — added `MarkMediaMissing` helper.
- `--no-reconcile` flag added to `movie scan`.
- Wired into `runMovieScan` BEFORE `executeScan`.

## Phase 6 — QA + docs + cleanup  ✅ DONE v2.294.0
- README: added File Management table rows + condition grammar callout + SmartRescan note.
- AC walk-through done. Closed gaps:
  - R3 batch-id: `applyRm` now generates one BatchID for all rows in a batch.
  - R1 sidecar removal: `removeRmSidecar` deletes the JSON sidecar on rm.
  - R6 sidecar regen: `regenSidecarFor` recreates sidecar after `movie undo`.
- Deferred to follow-up tasks (filed below):
  - report.html regen after rm/move (currently relies on next `--rest`/scan).
  - Real 50+ file QA on user machine (cannot run in sandbox).
- M5 ✅ done v2.299.0: `executeBatchMovesAtomic` in cmd/movie_move_atomic.go;
  rolls back completed FS moves in reverse order on first failure; DB writes
  only happen on full success. `--no-atomic` opt-out preserves legacy flow.

## Open decisions (deferred to future asks)
- `--purge` flag on rm to also delete the on-disk video file.
- Cross-folder global JSON cache (current scope = per-folder only).
- Auto-regenerate `report.html` immediately after rm/move.
