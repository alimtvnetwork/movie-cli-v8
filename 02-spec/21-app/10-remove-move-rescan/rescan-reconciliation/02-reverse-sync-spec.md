# Reverse Sync Spec (DB → JSON → Disk awareness)

Companion to `01-spec.md` (forward SmartRescan). Reverse sync is invoked
**inside `movie scan`** after the forward reconcile pass, and also
exposed as `movie scan --reverse-sync` / `movie scan --reverse-sync-only`
for explicit runs.

## Purpose

Forward sync answers: *"What's on disk that the DB doesn't know about?"*
Reverse sync answers: *"What does the DB / JSON know about that disk
or sidecars no longer reflect?"* — and repairs the JSON sidecars so
that an offline machine (no DB) can still rebuild state from JSON.

## Trust order (authoritative)

When sources disagree, the rescan trusts in this order:

1. **SQLite DB** — if a row exists for the scan folder, it wins.
2. **JSON sidecar** — used only when no DB row exists (cold machine,
   fresh clone). Sidecars hydrate the DB; then DB becomes authoritative.
3. **TMDb** — last resort, only for files with neither DB row nor
   sidecar.

## Trigger matrix

| Invocation                           | Forward | Reverse |
|--------------------------------------|:-------:|:-------:|
| `movie scan`                         |   ✅    |   ✅    |
| `movie scan --no-reconcile`          |   ❌    |   ❌    |
| `movie scan --no-reverse-sync`       |   ✅    |   ❌    |
| `movie scan --reverse-sync-only`     |   ❌    |   ✅    |
| `movie scan --dry-run`               |   ✅    |   ✅ (read-only) |

## Reverse sync flow (8 steps)

```
R1. Load dbSet     = Media WHERE CurrentFilePath LIKE '<folder>%'
                     (include IsDeleted rows so we can purge stale sidecars)
R2. Load jsonSet   = sidecars under .movie-output/json/**
R3. Build diskSet  = on-disk video files (reuses collectVideoFiles)

R4. For each db row that is Active and CurrentFilePath ∈ diskSet:
        if no matching sidecar OR sidecar is stale (mtime < db.UpdatedAt):
            rewriteSidecar(row)      # DB → JSON
            log → ReverseSyncedSidecar

R5. For each sidecar with no matching db row (orphan sidecar):
        if file ∈ diskSet:
            keep — forward pass will hydrate next run
        else:
            deleteSidecar
            log → RemovedOrphanSidecar

R6. For each db row marked IsDeleted=1 with sidecar still present:
        deleteSidecar
        log → RemovedDeletedSidecar

R7. For each db row with CurrentFilePath ∉ diskSet AND not yet Missing:
        markMissing(row)             # same as forward step 5
        log → ReverseDetectedMissing

R8. Print summary; emit one ReconciliationHistory(Converged) row.
```

NO TMDb calls during reverse sync. Network is never touched.

## Conflict resolution

| Source A   | Source B    | Winner | Action                                    |
|------------|-------------|--------|-------------------------------------------|
| DB (newer) | JSON (older)| DB     | Rewrite sidecar                           |
| JSON only  | (no DB)     | JSON   | Forward pass hydrates DB                  |
| DB only    | (no JSON)   | DB     | Write sidecar                             |
| DB deleted | JSON exists | DB     | Delete sidecar                            |
| Disk gone  | DB+JSON     | Disk   | Mark missing, delete sidecar              |

`UpdatedAt` is the tiebreaker. Sidecar mtime is read via `os.Stat`.

## New action types (db/reconciliation_history.go)

- `ReverseSyncedSidecar`     — sidecar (re)written from DB
- `RemovedOrphanSidecar`     — sidecar deleted (no DB row, file gone)
- `RemovedDeletedSidecar`    — sidecar deleted (DB row soft-deleted)
- `ReverseDetectedMissing`   — file gone, marked missing during reverse pass

## Files to create / edit

- **new** `cmd/movie_scan_reverse.go`     — orchestrator (8 steps)
- **new** `cmd/movie_scan_reverse_write.go` — sidecar writer (DB → JSON)
- **edit** `cmd/movie_scan.go`            — wire `--reverse-sync-only`,
                                              `--no-reverse-sync` flags
- **edit** `db/reconciliation_history.go` — add 4 new action constants
- **new** `spec/06-diagrams/17-reverse-sync-flow.mmd` — flowchart

## Performance target

Reverse sync of 1 000 unchanged files completes in < 1 s with zero disk
writes (only `os.Stat` reads). Sidecar rewrites happen only when
`UpdatedAt > sidecar.mtime`.

## CLI surface

```
movie scan                           # forward + reverse (default)
movie scan --no-reverse-sync         # forward only (legacy behaviour)
movie scan --reverse-sync-only       # reverse only, no new TMDb fetches
movie scan --dry-run --reverse-sync-only   # preview drift, no writes
```

## Acceptance criteria

1. After `movie rm <id>`, next `movie scan` removes the JSON sidecar.
2. After editing a Media row's title via SQL, next `movie scan`
   rewrites the sidecar with the new title (mtime advances).
3. Deleting a file from disk causes the next `movie scan` to mark it
   missing AND delete the sidecar in a single pass.
4. A sidecar with no DB row and no disk file is removed.
5. Zero TMDb requests in any reverse-sync code path.
6. `--dry-run --reverse-sync-only` prints planned actions but writes
   nothing to disk or DB.
