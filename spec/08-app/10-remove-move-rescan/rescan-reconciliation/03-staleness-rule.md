# Rescan Staleness Rule (1-year freshness)

Companion to `01-spec.md` (forward SmartRescan) and `02-reverse-sync-spec.md`.

## Problem

`mediaNeedsRescan` previously returned `true` only when a row was missing
`Genre`, `TmdbRating`, or `Description`. A fully-populated row was
treated as fresh forever, even if the cached metadata was years old —
ratings drift, IMDb IDs get filled in later, posters get re-uploaded.

## Rule

A `Media` row is rescanned during `movie scan` when **either**:

1. **Incomplete** — `Genre == ""` OR `TmdbRating == 0` OR `Description == ""`, OR
2. **Stale** — `now - Media.UpdatedAt > 365 days` (1 year).

Rows with an empty / unparseable `UpdatedAt` are treated as stale so
legacy entries get refreshed once on the next scan, then settle.

Constant: `cmd.MaxRescanAge = 365 * 24 * time.Hour`.

## Where it fits in the scan pipeline

The full flow `movie scan <folder>` already does:

1. **DB read first** — `GetMediaByScanDir(scanDir)` lists every Media row
   whose `OriginalFilePath` starts with the scan folder.
2. **SmartRescan reconcile** (`runSmartRescan`) — hydrates DB from JSON
   sidecars when the DB is empty, marks missing files, and detects
   converged paths. Zero TMDb calls.
3. **Disk diff** — `splitNewFromExisting` walks the disk, classifies each
   file as either *existing* (in DB) or *new*.
4. **Existing files** — handled serially on the main goroutine via
   `processExistingMedia`. For each row, `mediaNeedsRescan` decides
   between "skip" and "rescan via TMDb". This is where the staleness
   rule lives.
5. **New files** — dispatched to the parallel worker pool
   (`runParallelNewFileScan` in `cmd/movie_scan_pool.go`, `NumCPU*2`
   workers capped at 32) which performs the TMDb search + thumbnail
   download in parallel. The DB writes and JSON sidecar writes are
   serialised in the drain goroutine to avoid SQLite contention.
6. **Reverse sync** (`runReverseSync`) — repairs JSON sidecars to
   reflect DB state and updates `Media.UpdatedAt` where appropriate.
   When a reverse-sync action rewrites a sidecar it bumps the same
   timestamp the staleness rule reads.

## Marking missing files (with timestamp)

`markMissingItems` in `cmd/movie_scan_reconcile.go` walks DB rows whose
`CurrentFilePath` is no longer on disk and calls
`db.MarkMediaMissing(MediaId)`, which executes:

```sql
UPDATE Media
SET    MediaStatusId = MediaStatusMissing,
       UpdatedAt     = datetime('now')
WHERE  MediaId = ?
```

so both the DB row and (via reverse-sync rewriting the sidecar) the JSON
file carry an authoritative "last seen" timestamp.

## TMDb call budget

| Case                                   | TMDb call? |
|----------------------------------------|:----------:|
| New file (not in DB)                   | ✅ via worker pool |
| Existing file, complete, < 1 yr old    | ❌ skipped |
| Existing file, complete, ≥ 1 yr old    | ✅ rescan |
| Existing file, missing field           | ✅ rescan |
| Disk-missing row                       | ❌ marked Missing only |

A second scan of an unchanged 1 000-file library where every row was
written within the last year completes with **zero TMDb requests**.

## Implementation pointers

- Staleness check: `cmd/movie_rescan_helper.go::isMediaStale`
- Worker pool dispatcher: `cmd/movie_scan_pool.go::runParallelNewFileScan`
- DB-first read: `db.GetMediaByScanDir`, called from `runMainScanLoop`
- Missing-mark with timestamp: `db.MarkMediaMissing`
- `Media.UpdatedAt` is now selected via `mediaColumns`; the `Media`
  struct exposes it as a string (RFC3339 or SQLite `YYYY-MM-DD HH:MM:SS`).

## Override

There is intentionally no flag to disable the staleness check. Use
`--no-reconcile` (skip SmartRescan entirely) or delete the row to force
a fresh scan. To audit, run `movie ls --json` and inspect `updated_at`.
