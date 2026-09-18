---
name: Rescan staleness rule
description: 1-year freshness gate for Media rows; complete+fresh skip TMDb, otherwise rescan
type: feature
---

`movie scan` skips TMDb for an existing Media row only when it is BOTH
complete (Genre, TmdbRating, Description all populated) AND younger than
`cmd.MaxRescanAge = 365 * 24 * time.Hour`. Empty / unparseable
`UpdatedAt` counts as stale so legacy rows refresh once.

Pipeline (verified v2.321.0):
1. `db.GetMediaByScanDir` — DB read first.
2. `runSmartRescan` — hydrate from JSON, mark missing (writes
   `UpdatedAt = datetime('now')` via `MarkMediaMissing`).
3. `splitNewFromExisting` — disk diff.
4. Existing rows → `processExistingMedia` → `mediaNeedsRescan`
   (incomplete OR stale).
5. New rows → `runParallelNewFileScan` worker pool (NumCPU*2, cap 32,
   TMDb 40 req/s limiter).
6. `runReverseSync` — repairs sidecars, bumps `UpdatedAt`.

Files:
- `cmd/movie_rescan_helper.go::mediaNeedsRescan / isMediaStale / MaxRescanAge`
- `cmd/movie_scan_pool.go::runParallelNewFileScan`
- `db/media.go` — `Media.UpdatedAt` field added to struct + `mediaColumns`
- `db/media_softdelete.go::MarkMediaMissing`

Spec: `spec/08-app/10-remove-move-rescan/rescan-reconciliation/03-staleness-rule.md`

No flag disables the staleness check; use `--no-reconcile` or delete the
row to force a fresh scan.
