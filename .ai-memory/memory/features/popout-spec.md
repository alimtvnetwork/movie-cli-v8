---
name: Popout spec
description: movie popout flattens nested media to root and compacts non-media folders into <root>/.temp/. --auto-compact skips prompt. Full undo via batch_id.
type: feature
---

# `movie popout` — full spec (v2.136.0)

## Behavior
1. Resolve target dir via `ResolveTargetDir(args, home)` → defaults to cwd.
2. Walk subtree (depth flag, default 3) → discover all video files in subfolders.
3. Move each media file to root with cleaned filename (or original if `--no-rename`).
4. After moves, scan every direct subfolder of root. If a folder has zero media files (recursively) OR is empty → it's a **compaction candidate**.
5. Compaction phase:
   - Default: prompt `[a/s/l/N]` (compact-all / select / list / no).
   - With `--auto-compact`: skip prompt, compact every candidate.
6. Each compaction MOVES the folder into `<root>/.temp/`. Never deletes. Collisions → `-2`, `-3` suffix.
7. Each compaction inserts an `ActionHistory` row with `FileActionCompact` (= 15) tagged with the run's `batch_id`.

## Flags
- `--dry-run` — preview only
- `--no-rename` — keep original filename
- `--depth N` — max subfolder depth (default 3)
- `--auto-compact` — skip the y/N prompt for `.temp/` compaction

## Undo / Redo
- `movie undo --batch <id>` reverts every move + every compact in the batch.
  Compact rows are restored by `undoCompact` in `cmd/movie_popout_restore.go`,
  which parses the JSON snapshot `{original_path, compact_path}` and moves
  the folder back from `<root>/.temp/Folder` to its original location.
  Refuses to overwrite if the original path already exists.
- `movie redo --batch <id>` re-applies them via `redoCompact` (symmetric).

## Files
- `cmd/movie_popout.go` — entry point, orchestration, preview, prompt, batch ID
- `cmd/movie_popout_discover.go` — `discoverNestedVideos`, `discoverAllSubdirs`
- `cmd/movie_popout_cleanup.go` — `compactNonMediaFolders`, `folderHasMedia`, `compactFolder`
- `cmd/movie_popout_summary.go` — `printPopoutSummary` final report (files moved + folders compacted + batch undo hint)
- `cmd/movie_popout_restore.go` — `undoCompact` / `redoCompact` for FileActionCompact rows; parses `{original_path, compact_path}` snapshot
- `cmd/movie_popout_integration_test.go` — 7 integration tests covering discovery, classification, compaction, cwd-default

## Database
- `FileActionCompact = 15` enum + seed row in `db/action_history.go` and `db/seed.go`.
- `db/open_test.go` updated: expects 15 seeded actions, not 14.
