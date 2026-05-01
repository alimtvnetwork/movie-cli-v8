# Application Commands — Acceptance Criteria

**Version:** 1.3.0  
**Last Updated:** 2026-05-01  
**Format:** GIVEN/WHEN/THEN (E2E-test-ready)

---

## AC-01: Hello Command

**GIVEN** the CLI is installed  
**WHEN** the user runs `movie hello`  
**THEN** the output contains `👋 Hello from Movie CLI!` followed by the version string

---

## AC-02: Version Command

**GIVEN** the binary was built with `-ldflags` injecting version, commit, and build date  
**WHEN** the user runs `movie version`  
**THEN** the output shows `vX.Y.Z (commit: <hash>, built: <date>)`

**Edge Cases:**
- **GIVEN** the binary was built without ldflags **WHEN** `movie version` is run **THEN** defaults are shown: `v0.0.1-dev`, `none`, `unknown`

---

## AC-03: Self-Update Command

**GIVEN** the CLI resolves an existing clean git repository  
**WHEN** the user runs `movie self-update`  
**THEN** `git pull --ff-only` is executed and the old→new commit hashes are displayed

**Edge Cases:**
- **GIVEN** no local repo exists **WHEN** self-update runs **THEN** a fresh clone is created next to the binary and the output reports bootstrap success
- **GIVEN** git is not in PATH **WHEN** self-update runs **THEN** a clear error is shown: git not found
- **GIVEN** the working tree has uncommitted changes **WHEN** self-update runs **THEN** a clear error is shown: dirty working tree
- **GIVEN** there are no new commits in an existing repo **WHEN** self-update runs **THEN** the output says already up-to-date

---

## AC-04: Config Command

### AC-04a: Show All Config

**GIVEN** config values exist in the database  
**WHEN** `movie config` is run with no arguments  
**THEN** all config keys and values are printed  
**AND** the `tmdb_api_key` value is masked (first 4 + `...` + last 4 chars)

### AC-04b: Get Single Key

**GIVEN** a config key `movies_dir` exists  
**WHEN** `movie config get movies_dir`  
**THEN** the value of `movies_dir` is printed

### AC-04c: Set Key

**GIVEN** a valid config key  
**WHEN** `movie config set movies_dir /new/path`  
**THEN** the value is updated in the database  
**AND** a confirmation message is printed

**Edge Cases:**
- **GIVEN** an invalid config key **WHEN** `config get unknown_key` **THEN** an error is shown listing valid keys

---

## AC-05: Scan Command

**GIVEN** a directory contains video files and a TMDb API key is configured  
**WHEN** `movie scan /path/to/folder`  
**THEN** each video file is cleaned, metadata is fetched from TMDb, a poster is downloaded, and a media record is inserted into the database  
**AND** a summary is printed: total files, movies, TV shows, skipped  
**AND** a `scan_history` record is logged

**Edge Cases:**
- **GIVEN** no folder argument and no `scan_dir` config **WHEN** scan runs **THEN** an error is shown
- **GIVEN** a file's `original_file_path` already exists in the DB **WHEN** scan encounters it **THEN** the file is skipped (dedup)
- **GIVEN** no TMDb API key is available **WHEN** scan runs **THEN** a warning is printed but scanning continues without metadata
- **GIVEN** a directory entry is a subfolder containing a video file **WHEN** scan processes it **THEN** the directory name is used for title cleaning

---

## AC-06: List Command

**GIVEN** media records exist in the database  
**WHEN** `movie ls` is run  
**THEN** only records with a non-empty `current_file_path` are shown (file-backed items only)  
**AND** records created via `search` or `info` without local files are excluded  
**AND** a paginated list is shown with: number, clean title, year, rating (TMDb→IMDb fallback), type icon  
**AND** page size is read from config (default 20)

**Edge Cases:**
- **GIVEN** no file-backed media records exist **WHEN** `ls` is run **THEN** a "no media found" message is shown
- **GIVEN** 5 records exist but only 2 have `current_file_path` **WHEN** `ls` is run **THEN** only 2 items are listed
- **GIVEN** the user enters `N` **WHEN** on the last page **THEN** nothing happens or wraps to first page
- **GIVEN** the user enters a number **WHEN** on the list page **THEN** the detail view shows full metadata card

---

## AC-07: Search Command

**GIVEN** a TMDb API key is configured  
**WHEN** `movie search "The Matrix"`  
**THEN** up to 15 results are displayed with: number, icon, title, year, rating, type  
**AND** the user can select a result to fetch full details, download poster, and save to DB

**Edge Cases:**
- **GIVEN** no TMDb API key **WHEN** search runs **THEN** the command exits with "API key required" error
- **GIVEN** the user selects 0 **WHEN** at the selection prompt **THEN** the search is cancelled
- **GIVEN** search returns no results **WHEN** query has no matches **THEN** a "no results found" message is shown

---

## AC-08: Info Command

### AC-08a: By Numeric ID

**GIVEN** a media record with ID 5 exists in the DB  
**WHEN** `movie info 5`  
**THEN** the full detail card is displayed (title, year, type, ratings, genres, director, cast, description, paths)

### AC-08b: By Title String

**GIVEN** a media record titled "Inception" exists locally  
**WHEN** `movie info inception`  
**THEN** the match is resolved using priority: exact match → prefix match → first result  
**AND** the detail card is displayed

### AC-08c: Fallback to TMDb

**GIVEN** the title is not found in the local DB and a TMDb API key exists  
**WHEN** `movie info "Unknown Title"`  
**THEN** TMDb is searched, full details + credits + poster are fetched, the record is auto-saved, and the detail card is displayed

**Edge Cases:**
- **GIVEN** the TMDb result's `tmdb_id` already exists in the DB **WHEN** fallback runs **THEN** the existing record is returned (no duplicate)

---

## AC-09: Suggest Command

**GIVEN** a TMDb API key is configured  
**WHEN** `movie suggest 5`  
**THEN** the user is prompted to choose: Movie / TV / Random  
**AND** 5 suggestions are displayed with: title, year, rating, genre names

### AC-09a: Genre-Based Suggestions

**GIVEN** the user selects "Movie" or "TV" and the library has genre data  
**WHEN** suggestions are generated  
**THEN** `TopGenres()` is used, recommendations are fetched from TMDb, remaining slots filled with trending  
**AND** existing library items are excluded from results

### AC-09b: Random Suggestions

**GIVEN** the user selects "Random"  
**WHEN** suggestions are generated  
**THEN** trending movies + TV are merged, shuffled, and deduplicated

**Edge Cases:**
- **GIVEN** the library has no media **WHEN** genre-based suggestions run **THEN** fallback to trending only

---

## AC-10: Move Command

**GIVEN** a source directory contains video files  
**WHEN** `movie move /path/to/source`  
**THEN** video files are listed with clean titles, type icons, and file sizes  
**AND** the user selects a file and destination (Movies / TV / Archive / Custom)  
**AND** the file is renamed to `Title (Year).ext` and moved  
**AND** `move_history` is logged and media record is updated

**Edge Cases:**
- **GIVEN** source and destination are on different filesystems **WHEN** move runs **THEN** `io.Copy` + `os.Remove` fallback is used instead of `os.Rename`
- **GIVEN** the copy fails mid-transfer **WHEN** fallback is active **THEN** the source file is NOT deleted and an error is reported
- **GIVEN** no argument is provided **WHEN** move starts **THEN** interactive prompt offers: Downloads / Scan Dir / Custom path

### AC-10b: Batch Move

**GIVEN** a source directory with multiple video files  
**WHEN** `movie move /path --all`  
**THEN** all video files are moved to their destination using the selected category  
**AND** each file is renamed to `Title (Year).ext`  
**AND** all moves are logged to `move_history`

**Edge Cases:**
- **GIVEN** `--all` is used with a directory containing 0 video files **WHEN** move runs **THEN** "no video files found" is shown

---

## AC-11: Rename Command

**GIVEN** media records exist with `current_file_path` values  
**WHEN** `movie rename` is run  
**THEN** files whose names differ from `ToCleanFileName(cleanTitle, year, ext)` are listed as a preview  
**AND** user confirms with `y/N`  
**AND** confirmed renames are executed, DB paths updated, and `move_history` logged  
**AND** a summary is printed: `X/Y files renamed`

**Edge Cases:**
- **GIVEN** all files already have clean names **WHEN** rename runs **THEN** "nothing to rename" is shown
- **GIVEN** user enters `N` at confirmation **WHEN** prompted **THEN** no files are renamed

---

## AC-12: Undo Command

**GIVEN** a `move_history` record exists with `undone=0`  
**WHEN** `movie undo` is run  
**THEN** the most recent move is displayed (from → to paths)  
**AND** user confirms with `y/N`  
**AND** the file is moved back, the record is marked `undone=1`, and `current_file_path` is restored

**Edge Cases:**
- **GIVEN** no un-undone move records exist **WHEN** undo runs **THEN** "nothing to undo" is shown
- **GIVEN** the file no longer exists at `to_path` **WHEN** undo runs **THEN** an error is shown: file not found
- **GIVEN** user enters `N` at confirmation **WHEN** prompted **THEN** undo is cancelled with a message

---

## AC-13: Play Command

**GIVEN** a media record with ID 3 exists and `current_file_path` points to an existing file  
**WHEN** `movie play 3`  
**THEN** the file is opened with the platform-specific command (`open` / `xdg-open` / `cmd /c start`)

**Edge Cases:**
- **GIVEN** the media ID does not exist **WHEN** play runs **THEN** "media not found" error is shown
- **GIVEN** `current_file_path` does not exist on disk **WHEN** play runs **THEN** "file not found" error is shown

---

## AC-14: Stats Command

**GIVEN** media records exist in the database  
**WHEN** `movie stats` is run  
**THEN** the output shows: total movies, total TV shows, total count  
**AND** storage section shows: total file size, largest file, smallest file, average file size (human-readable format)  
**AND** top 10 genres with visual bar chart (`█` characters, max 30 width)  
**AND** average IMDb rating and average TMDb rating (if available)

**Edge Cases:**
- **GIVEN** no media records exist **WHEN** stats runs **THEN** all counts show 0, no storage section, and no genre chart
- **GIVEN** media exists but all have `file_size = 0` **WHEN** stats runs **THEN** the storage section is not displayed

---

## AC-15: Tag Command

### AC-15a: Add Tag

**GIVEN** media with ID 1 exists  
**WHEN** `movie tag add 1 favorite`  
**THEN** the tag is inserted into the `tags` table  
**AND** confirmation: `Tag "favorite" added to "Title (Year)"`

### AC-15b: Duplicate Tag

**GIVEN** tag "favorite" already exists on media ID 1  
**WHEN** `movie tag add 1 favorite`  
**THEN** error: `tag already exists`

### AC-15c: Remove Tag

**GIVEN** tag "favorite" exists on media ID 1  
**WHEN** `movie tag remove 1 favorite`  
**THEN** the tag is deleted and confirmation is printed

### AC-15d: Remove Non-Existent Tag

**GIVEN** tag "unknown" does not exist on media ID 1  
**WHEN** `movie tag remove 1 unknown`  
**THEN** error: `tag not found`

### AC-15e: List Tags for Media

**GIVEN** tags exist on media ID 1  
**WHEN** `movie tag list 1`  
**THEN** all tags for that media are shown

### AC-15f: List All Tags

**GIVEN** tags exist across multiple media  
**WHEN** `movie tag list` (no ID)  
**THEN** all unique tags are shown with media count, e.g., `favorite (3)`

---

## AC-16: Export Command

**GIVEN** media records exist in the database  
**WHEN** `movie export` is run  
**THEN** all media records are serialized to JSON and written to `./data/json/export/media.json`  
**AND** output shows: `Exported N items → <path>`

### AC-16a: Custom Output Path

**GIVEN** media records exist  
**WHEN** `movie export -o ~/Desktop/library.json`  
**THEN** the JSON is written to the specified path instead of the default

**Edge Cases:**
- **GIVEN** no media records exist **WHEN** export runs **THEN** message: "No media to export"
- **GIVEN** the output directory does not exist **WHEN** export runs **THEN** the directory is created automatically
- **GIVEN** the output path is read-only **WHEN** export runs **THEN** an error is shown

---

## AC-17: Redo Command

### AC-17a: Redo Last Reverted Operation

**GIVEN** a previous `movie undo` reverted at least one move/rename within the current working directory  
**WHEN** the user runs `movie redo`  
**THEN** the most recent reverted operation is re-applied  
**AND** the affected `move_history` / `action_history` rows are flipped back to `IsReverted = 0`

### AC-17b: List Redoable Actions

**GIVEN** one or more reverted actions exist in scope  
**WHEN** `movie redo --list` is run  
**THEN** a numbered table of redoable actions is printed (id, timestamp, summary)

### AC-17c: Redo by ID

**GIVEN** a reverted `action_history` row with id 42 exists  
**WHEN** `movie redo --id 42` is run  
**THEN** only that action is re-applied

### AC-17d: Redo Entire Batch

**GIVEN** the last revert affected a batch of files  
**WHEN** `movie redo --batch` is run  
**THEN** every member of that batch is re-applied in original order

**Edge Cases:**
- **GIVEN** no reverted actions exist in scope **WHEN** redo runs **THEN** message: "Nothing to redo"
- **GIVEN** `--global` is passed **WHEN** redo runs **THEN** the cwd / path scope is ignored
- **GIVEN** `--include` / `--exclude` globs are supplied **WHEN** redo runs **THEN** matching paths are kept / dropped before execution
- **GIVEN** `--yes` is passed **WHEN** redo runs **THEN** all confirmation prompts are skipped

---

## AC-18: Doctor Command

**GIVEN** the CLI is installed  
**WHEN** the user runs `movie doctor`  
**THEN** the output reports the status of each environment check: SQLite database reachable, TMDb API key configured, `movies_dir` and `tv_dir` exist and are writable, git binary present, current binary path  
**AND** each check is marked ✅ pass or ❌ fail with a one-line reason

**Edge Cases:**
- **GIVEN** any check fails **WHEN** doctor finishes **THEN** the process exits with a non-zero code
- **GIVEN** all checks pass **WHEN** doctor finishes **THEN** the process exits with code 0

---

## AC-19: Duplicates Command

**GIVEN** the library contains two or more media rows with the same `(CleanTitle, Year)`  
**WHEN** the user runs `movie duplicates`  
**THEN** each duplicate group is printed with all member ids, sizes, and current file paths  
**AND** a summary line shows the total duplicate count

**Edge Cases:**
- **GIVEN** no duplicates exist **WHEN** the command runs **THEN** message: "No duplicates found"

---

## AC-20: Cleanup Command

**GIVEN** the library contains rows whose `CurrentFilePath` no longer exists on disk  
**WHEN** the user runs `movie cleanup`  
**THEN** the user is shown the list of orphaned rows and prompted for confirmation  
**AND** on confirmation the orphaned rows are removed from the database

**Edge Cases:**
- **GIVEN** no orphaned rows exist **WHEN** cleanup runs **THEN** message: "Library is clean"
- **GIVEN** the user declines the prompt **WHEN** cleanup runs **THEN** no rows are deleted

---

## AC-21: Watch Command

### AC-21a: Add to Watchlist

**GIVEN** a media row with id 5 exists  
**WHEN** `movie watch add 5` is run  
**THEN** a `Watchlist` row is created with `MediaID = 5` and `IsWatched = 0`

### AC-21b: Mark Done

**GIVEN** id 5 is on the watchlist  
**WHEN** `movie watch done 5` is run  
**THEN** the watchlist row is updated with `IsWatched = 1` and `WatchedAt = now`

### AC-21c: Remove from Watchlist

**GIVEN** id 5 is on the watchlist  
**WHEN** `movie watch rm 5` is run  
**THEN** the watchlist row is deleted

### AC-21d: List Watchlist

**GIVEN** the watchlist has entries  
**WHEN** `movie watch list` is run  
**THEN** a table of watchlist items prints with title, year, added date, watched status

**Edge Cases:**
- **GIVEN** the media id does not exist **WHEN** any subcommand runs **THEN** an error is shown and exit code is non-zero

---

## AC-22: History Command

**GIVEN** at least one move or rename has been recorded  
**WHEN** the user runs `movie history`  
**THEN** a table of recent operations prints with id, timestamp, action type, source, destination, IsReverted flag

**Edge Cases:**
- **GIVEN** no history rows exist **WHEN** the command runs **THEN** message: "No history yet"

---

## AC-23: DB Command

**GIVEN** the SQLite database file exists  
**WHEN** the user runs `movie db`  
**THEN** the resolved database file path, schema version, table count, and total row count are printed

**Edge Cases:**
- **GIVEN** the database file is missing **WHEN** the command runs **THEN** the file is created with the latest schema and a "fresh database" message is shown

---

## AC-24: REST Command

**GIVEN** the database is reachable  
**WHEN** the user runs `movie rest`  
**THEN** an HTTP server starts on the configured port and serves the dashboard plus JSON endpoints  
**AND** the bound URL is printed (e.g. `http://localhost:8765`)

**Edge Cases:**
- **GIVEN** the configured port is in use **WHEN** rest starts **THEN** an error is shown and exit code is non-zero
- **GIVEN** the user presses Ctrl+C **WHEN** the server is running **THEN** the server shuts down cleanly

---

## AC-25: Logs Command

**GIVEN** the application has written log entries  
**WHEN** the user runs `movie logs`  
**THEN** the most recent log entries are streamed to stdout in chronological order

**Edge Cases:**
- **GIVEN** no log file exists **WHEN** the command runs **THEN** message: "No logs available"

---

## AC-26: CD Command

**GIVEN** a folder name argument matches a known media root (e.g. `movies`, `tv`)  
**WHEN** the user runs `movie cd movies`  
**THEN** the resolved absolute path is printed so a shell wrapper can `cd` into it

**Edge Cases:**
- **GIVEN** the folder name is unknown **WHEN** the command runs **THEN** an error is shown listing valid folder names

---

## AC-27: Rescan Command

**GIVEN** the library contains media rows that lack TMDb metadata  
**WHEN** the user runs `movie rescan`  
**THEN** each row is re-queried against TMDb and updated with fresh metadata  
**AND** a summary prints: scanned, updated, failed counts

---

## AC-28: Rescan-Failed Command

**GIVEN** previous TMDb lookups recorded misses in `TmdbCacheMiss`  
**WHEN** the user runs `movie rescan-failed`  
**THEN** only the rows that previously failed are re-queried  
**AND** successful matches clear the corresponding miss entry

**Edge Cases:**
- **GIVEN** no failed lookups exist **WHEN** the command runs **THEN** message: "No failed lookups to retry"

---

## AC-29: Popout Command

### AC-29a: Discover Nested Videos

**GIVEN** the target directory contains subfolders with video files  
**WHEN** `movie popout` is run with no flags  
**THEN** all nested video files are listed with their proposed destination at the directory root

### AC-29b: Apply Popout

**GIVEN** the discovery list is non-empty  
**WHEN** the user confirms the prompt  
**THEN** each nested video is moved to the root of the target directory  
**AND** every move is logged to `move_history` for undo

### AC-29c: Cleanup Empty Folders

**GIVEN** popout left behind empty subfolders  
**WHEN** popout completes  
**THEN** every now-empty subfolder is removed

**Edge Cases:**
- **GIVEN** no nested videos exist **WHEN** popout runs **THEN** message: "Nothing to pop out"
- **GIVEN** a destination filename collision exists **WHEN** popout runs **THEN** the file is renamed with a numeric suffix and the action is recorded

---

## AC-30: Discover Command

**GIVEN** a TMDb API key is configured  
**WHEN** the user runs `movie discover [genre]`  
**THEN** TMDb's discover endpoint is called for that genre and the results are printed in a table  
**AND** the user can pick an item to save to the library

**Edge Cases:**
- **GIVEN** no TMDb key is configured **WHEN** discover runs **THEN** an error is shown asking to set the key
- **GIVEN** an unknown genre is passed **WHEN** discover runs **THEN** an error lists the supported genres

---

## AC-31: Cache Command

### AC-31a: List Cache

**GIVEN** TMDb cache rows exist  
**WHEN** `movie cache` is run  
**THEN** a summary prints with hit count, miss count, and last update timestamp

### AC-31b: Backfill Cache

**GIVEN** library rows have TMDb ids but no cached payloads  
**WHEN** `movie cache backfill` is run  
**THEN** each missing payload is fetched and stored in `TmdbCache`

### AC-31c: Forget Entry

**GIVEN** a cache entry exists for `cleanTitle = "Inception", year = 2010`  
**WHEN** `movie cache forget Inception 2010` is run  
**THEN** that cache entry is deleted

### AC-31d: Clear Misses

**GIVEN** miss entries exist in `TmdbCacheMiss`  
**WHEN** `movie cache clear-misses` is run  
**THEN** all rows in `TmdbCacheMiss` are deleted

---

## AC-32: Changelog Command

**GIVEN** the binary embeds the project changelog  
**WHEN** the user runs `movie changelog`  
**THEN** the full changelog is printed, newest entries first

### AC-32a: Latest Only

**GIVEN** the changelog has at least one release  
**WHEN** `movie changelog --latest` is run  
**THEN** only the most recent release section is printed

---

## AC-33: Update Command

**GIVEN** a clean local repo with new commits on the remote  
**WHEN** the user runs `movie update`  
**THEN** the binary pulls latest, rebuilds, deploys, and prints `Updated: vOLD → vNEW`

**Edge Cases:**
- **GIVEN** the repo is at the latest commit **WHEN** update runs **THEN** message: "Already up to date" and no rebuild
- **GIVEN** git is not in PATH **WHEN** update runs **THEN** an error is shown and exit code is non-zero
- **GIVEN** the repo has uncommitted changes **WHEN** update runs **THEN** message: "repository has local changes; commit or stash them"
- **GIVEN** no local repo exists **WHEN** update runs **THEN** a fresh clone is created next to the binary and the user is told to run `movie update` again
- **GIVEN** the version did not change after build **WHEN** update finishes **THEN** warning: "Version unchanged — was version/info.go bumped?"
- **GIVEN** the deploy step fails **WHEN** `run.ps1` cannot copy the new binary **THEN** the `.bak` backup is restored
- **GIVEN** a successful update **WHEN** build and deploy complete **THEN** `movie update-cleanup` runs automatically and `movie changelog --latest` is displayed

See [Self-Update Acceptance Criteria](../13-self-update-app-update/07-acceptance-criteria.md) for the full set.

---

## AC-34: Update-Cleanup Command

**GIVEN** leftover temp files from a previous update exist  
**WHEN** the user runs `movie update-cleanup`  
**THEN** all `movie-update-*` and `*.bak` files in the binary directory are removed  
**AND** the currently running binary is NOT deleted

---

## AC-35: Self-Replace Command

**GIVEN** an update worker has produced a new binary at `<bin>.new`  
**WHEN** the worker invokes `movie self-replace --target <bin>`  
**THEN** the running binary is swapped atomically with the new one and the previous version is preserved as `<bin>.bak`

**Edge Cases:**
- **GIVEN** the target binary is locked (Windows) **WHEN** self-replace runs **THEN** the operation is retried with a backoff and falls back to `.bak` restore on failure

---

## Cross-References

- [Overview](./00-overview.md)
- [Project Spec](./01-project-spec.md)
- [Error Management AC](../02-error-manage-spec/97-acceptance-criteria.md)
- [Seedable Config AC](../04-seedable-config-architecture/98-acceptance-criteria.md)
