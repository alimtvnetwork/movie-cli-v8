---
name: Reverse-sync pass
description: movie scan reverse-sync (DB→JSON→disk awareness) — flags, action types, trust order
type: feature
---
Reverse sync runs after forward SmartRescan. Treats SQLite DB as authoritative
and reconciles JSON sidecars + disk awareness to match. Zero TMDb calls.

Flags:
- `--no-reverse-sync` — skip reverse pass (default: enabled)
- `--reverse-sync-only` — skip forward scan, run reverse only
- `--dry-run` — preview reverse actions, write nothing

Trust order: DB > JSON sidecar > TMDb (last resort).

8 steps (R1..R8):
- R4 ReverseSyncedSidecar: rewrite stale sidecar (mtime < UpdatedAt)
- R5 RemovedOrphanSidecar: delete sidecar with no DB row + no disk file
- R6 RemovedDeletedSidecar: delete sidecar for IsDeleted=1 row
- R7 ReverseDetectedMissing: mark missing + delete sidecar when file gone

New action constants (db/migrate_v6 seeds them):
- ReconActionReverseSyncedSidecar=5
- ReconActionRemovedOrphanSidecar=6
- ReconActionRemovedDeletedSidecar=7
- ReconActionReverseDetectedMissing=8

Files: cmd/movie_scan_reverse.go, cmd/movie_scan_reverse_write.go,
db/reverse_sync_query.go, db/migrate_v6.go.
Spec: spec/08-app/10-remove-move-rescan/rescan-reconciliation/02-reverse-sync-spec.md
Diagram: spec/06-diagrams/17-reverse-sync-flow.mmd
