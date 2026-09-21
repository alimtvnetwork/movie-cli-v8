---
name: movie-cli-file-operations-and-history
description: "File organization, interactive and selector move, batch rename, popout compaction, trash bin deletion, and atomic undo/redo in movie-cli-v8."
---

# Movie CLI File Operations and History Skill

## Overview

The file operations subsystem of `movie-cli-v8` handles moving, renaming, extracting (popout), and deleting media files on disk, ensuring every operation is recorded in `move_history` and `action_history` with full, multi-level undo and redo capabilities.

## Move Architecture (`cmd/movie_move*.go`)

### 1. Interactive Move Mode
```sh
movie move [directory] [flags]
```
- Browses target folder (defaults to cwd if omitted).
- Lists discovered video files with human-readable file sizes.
- Prompts destination directory with auto-expansion of `~` (home directory).
- Flags:
  - `--all`: Moves all video files in directory at once without prompting per file.
  - `--no-atomic`: Disables atomic rollback; continues past individual failures.

### 2. Selector Move Mode (Bulk Query Relocation)
```sh
movie move <selector> <destination> [flags]
```
- Moves matching database records directly to a new directory.
- Selectors supported:
  - Numeric Media ID: `movie move 42 ~/Movies`
  - Exact or partial Title: `movie move "Inception" ~/Movies`
  - Query Expression: `movie move "rating < 5 AND year >= 2010" ~/Archive`
  - Genre flag sugar: `movie move -g Horror ~/Movies/Horror` (equivalent to `g = Horror`)
  - `--yes`, `-y`: Bypasses confirmation prompt for automated scripting.

### 3. Atomic Batch Execution (`cmd/movie_move_atomic.go`)
When moving batches of files, if any single file copy or move fails (e.g. disk full, permission denied):
- The atomic mover immediately halts.
- Any files already moved in that batch are rolled back to their original source paths.
- Database records are restored to prevent partial state drift.

## Popout & Folder Compaction (`cmd/movie_popout*.go`)

Downloads and rips often nest media files inside release subfolders containing samples, nfo files, and subtitles.
```sh
movie popout [directory] [flags]
```
- **Discovery:** Recursively finds video files up to `--depth N` (default `3`).
- **Extraction:** Moves video files up to the root target directory with cleaned filenames.
- **Non-Destructive Compaction:** After videos are extracted, leftover folders are moved to `<root>/.temp/` rather than deleted permanently.
- **Flags:**
  - `--dry-run`: Previews extraction and compaction without touching disk.
  - `--no-rename`: Moves nested videos up while preserving raw file names.
  - `--auto-compact`: Bypasses the confirmation prompt before compacting to `.temp/`.

## Batch Rename (`cmd/movie_rename.go`)

```sh
movie rename
```
- Identifies library items where `CurrentFilePath` basename does not match `CleanTitle (Year).ext`.
- Generates a preview table of current vs proposed filenames.
- Prompts for user confirmation before renaming on disk and updating the database.

## Undo & Redo Architecture (`cmd/movie_undo*.go`, `cmd/movie_redo*.go`)

### History Tracking Tables
- `move_history`: Records `MediaId`, `FromPath`, `ToPath`, `OriginalFileName`, `NewFileName`, `IsReverted`, `MovedAt`.
- `action_history`: Records `FileActionId`, `MediaId`, `MediaSnapshot` (full JSON state), `Detail`, `BatchId`, `IsReverted`, `CreatedAt`.

### Undo Scoping & Flags
```sh
movie undo [path] [flags]
```
- **Scoped by Default:** Reverts only actions that occurred within the current working directory (or explicit `[path]`).
- `--global`: Reverts operations across the entire database regardless of path scope.
- `--batch`: Reverts the entire last batch (e.g., all moves from a single `scan` or `move --all`).
- `--id <N>`: Reverts a specific `action_history` entry by primary key.
- `--move-id <N>`: Reverts a specific `move_history` entry.
- `--include <glob>` / `--exclude <glob>`: Filters which file operations to revert.
- `--yes`, `-y`: Bypasses interactive confirmation.

### Redo Command
```sh
movie redo
```
- Re-executes the most recently undone operation by reading un-reverted records from history.

## Safe Deletion & Trash Bin (`pkg/trashbin/`, `cmd/movie_rm.go`)

- **Soft Delete in DB:** Marks media item as deleted in database while retaining history.
- **OS Recycle Bin Integration (`pkg/trashbin/`):**
  - Windows: Uses COM interface `IFileOperation` or shell API to send files to the Windows Recycle Bin.
  - macOS: Uses `osascript` AppleScript to move files to the user's Trash.
  - Linux: Complies with FreeDesktop.org Trash specification (`$XDG_DATA_HOME/Trash`).
  - Fallback: Non-destructive move to local `.trash/` directory.
