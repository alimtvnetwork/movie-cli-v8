# Completed Plan 12: Movie CLI Web UI, Staged Trash Deletions, TMDB Image Fallback, Reset Command, and Colorful Terminal Help

> **Plan Number:** 12  
> **Status:** Completed  
> **Completion Date:** 2026-09-20  
> **Repository:** alimtvnetwork/movie-cli-v8  

---

## 1. Executive Summary

This plan addressed comprehensive core enhancements to `movie-cli`:
1. **Web UI & REST Service (`movie ui`)**: Dedicated CLI command launching the local REST server and automatically opening the web dashboard in the system's default browser.
2. **Staged Actions & Action History**: Soft-staged deletions and modifications across UI and CLI. Deletion requests are recorded in `StagedAction` table rather than executing immediately, enabling review, undo, discard, and batch "Accept All" execution.
3. **Cross-Platform Safe Trash Deletion (`pkg/trashbin`)**: Replaced all unlinking/hard-deletion logic with safe OS Recycle Bin / Trash Bin movement across Windows (`SHFileOperationW` + PowerShell fallback), macOS (`osascript` Finder), and Linux (`gio trash` / XDG Trash spec / quarantine fallback). Zero direct `os.RemoveAll` calls on user media.
4. **AppFault Error Management Standard**: Fully standardized error management using `*appfault.AppError`, `appfault.Wrap`, and `appfault.New` across all newly created and modified handlers.
5. **TMDB Image & Screenshot Fallback**: Fixed poster dropping during IMDb cache hits, enriched database schema with `BackdropPath`, implemented multi-tier fallback querying TMDB `/images` endpoint for posters and backdrops.
6. **System Reset Command & REST Endpoint (`movie reset`)**: Added full system reset capability wiping `.movie-output` scan folders, SQLite database, JSON sidecars, cached thumbnails, and error logs with interactive and `--force` safeguards.
7. **Colorful ANSI Terminal Help System**: Branded, categorized, and styled Cobra help formatting with ANSI colors (Cyan headers, Green commands, Yellow flags, Dim descriptions) honoring `NO_COLOR` and terminal capability detection.

---

## 2. Implemented Components & Files

### A. Trash Bin Safe Deletion (`pkg/trashbin/`)
- `pkg/trashbin/trash.go`: Universal `MoveToTrash` dispatcher and `appfault` error handling.
- `pkg/trashbin/trash_windows.go`: Windows shell Recycle Bin via `SHFileOperationW` with `FO_DELETE` and `FOF_ALLOWUNDO`, plus PowerShell fallback.
- `pkg/trashbin/trash_darwin.go`: macOS Trash via AppleScript / Finder.
- `pkg/trashbin/trash_linux.go`: Linux FreeDesktop Trash via `gio trash` and XDG Trash specification.
- `pkg/trashbin/trash_fallback.go`: Local `~/.movie/trash/` quarantine fallback.
- `pkg/trashbin/trash_other.go`: Safe stub for unsupported operating systems.
- `cmd/movie_rm_apply.go`: Refactored `purgeOnDiskFile` to call `trashbin.MoveToTrash`.

### B. Database Schema & Staged Actions (`db/`)
- `db/migrate_v7.go`: Migration V7 adding `StagedAction` table with indices (`IdxStagedAction_Status`, `IdxStagedAction_Batch`) and `BackdropPath TEXT` to `Media`.
- `db/migrate.go`: Registered Migration V7 in migration chain.
- `db/media.go` & `db/media_query.go`: Added `BackdropPath` to `Media` model, queries, inserts, and scans.
- `db/staged_action.go`: Staged action domain model, queries (`InsertStagedAction`, `ListPendingStagedActions`, `GetStagedActionByID`, `UpdateStagedStatus`, `DiscardStagedAction`, `DiscardAllPendingStagedActions`).

### C. Web UI, REST API & HTML Report (`cmd/`, `templates/`)
- `cmd/movie_ui.go`: Implemented `movie ui` command with `--port`, `--host`, `--no-open`, banner, and browser launcher.
- `cmd/movie_rest_staged.go`: REST handlers for `/api/staged` (GET list, POST stage, DELETE discard all, POST single apply, POST apply all, DELETE single discard).
- `cmd/movie_rest.go`: Mounted `/api/staged`, `/api/staged/`, and `/api/system/reset` in `buildRESTMux`.
- `cmd/movie_rest_report.go`: Updated `handleMediaDelete` to stage deletion rather than hard delete.
- `templates/report.html`: Added staged changes bottom bar UI, Undo/Discard All/Accept All actions, and integrated staged changes check into REST connectivity.

### D. TMDB Image Fallback & Backdrops (`tmdb/`, `cmd/`)
- `tmdb/types.go`: Added `BackdropPath` to `SearchResult`, `MovieDetails`, `TVDetails`; added `ImageItem` and `MediaImagesResponse`.
- `tmdb/client.go`: Implemented `DownloadImage(imagePath, dst, size)`, `GetMovieImages(tmdbID)`, and `GetTVImages(tmdbID)`.
- `cmd/movie_fetch_details.go`: Enriched `applyMovieDetails`, `fetchMovieDetails`, `applyTVDetails`, and `fetchTVDetails` with poster and backdrop fallback.
- `cmd/movie_info_helpers.go` & `cmd/movie_info.go`: Updated thumbnail downloaders to fall back to `ThumbnailPath` or `BackdropPath`.
- `cmd/movie_scan_process.go` & `cmd/movie_scan_process_helpers.go`: Updated scan processing pipeline with poster and backdrop fallback.

### E. System Reset Command & REST Endpoint (`cmd/`)
- `cmd/movie_reset.go`: Implemented `movie reset` with flags (`-f/--force`, `-y/--yes`, `--keep-config`, `--all`, `--dry-run`).
- `cmd/movie_reset_helpers.go`: Target discovery across scan folders and base directories, confirmation prompt, dry-run display, wipe executor, and DB re-initialization.
- `cmd/movie_rest_handlers.go`: Implemented `handleSystemReset` for `POST /api/system/reset`.

### F. Colorful Terminal Help (`cmd/`)
- `cmd/help_formatter.go`: ANSI color helpers, `isColorEnabled()` checking `NO_COLOR` and TTY, and Cobra help wiring.
- `cmd/help_categories.go`: Grouped help formatter categorizing commands into Core Commands, Library Operations, Discovery & Play, and System & Maintenance.
- `cmd/root.go`: Registered `movieUiCmd` and `movieResetCmd`, and wired `setupColorfulHelp(rootCmd)`.

---

## 3. Verification & Compliance
- **Safe Deletions**: Zero hard `os.RemoveAll` calls on media. All deletions go to OS Recycle/Trash Bin via `pkg/trashbin`.
- **Naming Conventions**: Strict lowercase file names for all created files.
- **Git Paths**: All paths in markdown and code are relative.
- **Boolean Standards**: Positive booleans only, implicit checks, no mixed polarity conditions.
- **Error Types**: Standardized on `*appfault.AppError` and `pkg/appfault`.
