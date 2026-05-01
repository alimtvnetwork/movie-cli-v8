# Smart Rescan / Reconciliation Spec

## Trigger

Every invocation of `movie scan <folder>`. No new flag — the new
behaviour is on by default. `--no-reconcile` opt-out flag is added for
debugging.

## Pre-scan check

```
outputDir = <folder>/.movie-output
if outputDir exists and contains json/**:
    enter SmartRescan path
else:
    fall through to existing FullScan path  (no behaviour change)
```

## SmartRescan flow (matches user's mermaid)

```
1. List on-disk video files          → diskSet
2. List JSON sidecars                → jsonSet
3. Query Media WHERE
     CurrentFilePath LIKE '<folder>%'
     AND IsDeleted = 0                → dbSet

4. If dbSet is empty AND jsonSet not empty:
       hydrateFromJson(jsonSet)
       log each → ReconciliationHistory(Name=HydratedFromJson)

5. For each item in jsonSet ∪ dbSet:
       if item.path NOT in diskSet:
           markMissing(item)          # MediaStatusId=Missing
           deleteJsonSidecar(item)
           log → ReconciliationHistory(Name=RemovedMissing)

6. For each path in diskSet:
       if path NOT in (jsonSet ∪ dbSet):
           processNewFile(path)       # full TMDb pipeline
           log → ReconciliationHistory(Name=AddedNew)

7. For each path in diskSet ∩ dbSet:
       skip (Converged) — NO TMDb call
       (one summary row, not per-item, with Name=Converged)

8. Regenerate report.html once at the end.
```

## Hydration logic

`hydrateFromJson(path)` reads the sidecar, validates required fields
(`Title`, `Type`, `TmdbId` may be 0 for un-matched items), and inserts
one `Media` row with `MediaStatusId=Active`. NO TMDb call.

## Verification of disk presence

A single `os.Stat` per JSON-listed path. Use the existing parallel
worker pool (`runParallelNewFileScan`) for step 6 only — hydration and
stat checks are I/O-cheap and stay on the main goroutine.

## Files to create

- `cmd/movie_scan_reconcile.go` — orchestrator (the 8-step flow)
- `cmd/movie_scan_hydrate.go`   — JSON → DB hydration
- `db/reconciliation_history.go` — insert/query helpers
- `db/migrate_v4.go`             — adds `MediaStatus`, `Media.IsDeleted`, `Media.MediaStatusId`
- `db/migrate_v5.go`             — adds `ReconciliationActionType`, `ReconciliationHistory`

## Performance target

A second scan of an unchanged 1 000-file library completes in < 2 s
with **zero TMDb requests** and zero thumbnail downloads.
