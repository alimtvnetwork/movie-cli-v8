---
name: movie-cli-taxonomy-and-discovery
description: "Media search, recommendations, user tagging, TV season/episode tracking, watchlist management, and default player playback in movie-cli-v8."
---

# Movie CLI Taxonomy and Discovery Skill

## Overview

This subsystem handles user interaction with the media collection once scanned: searching TMDb without local files, inspecting detailed metadata, managing custom tags, tracking TV seasons and watched episodes, curating a watchlist, getting recommendations, and launching playback via the OS default video player.

## Core Commands & Workflows

### 1. Search & Remote Ingestion (`cmd/movie_search*.go`)
```sh
movie search "Inception" [flags]
```
- Queries TMDb API for titles matching query string.
- Automatically fetches ratings, genres, cast, crew, overview, and poster URL.
- Saves the record into `movie.db` (does NOT require a local video file to exist).
- Flags:
  - `--format json`: Emits machine-readable JSON array of search results to stdout.
  - `--format table`: Prints formatted terminal table (non-interactive).

### 2. Detailed Metadata Inspection (`cmd/movie_info*.go`)
```sh
movie info 42               # Lookup by local database Media ID
movie info "Breaking Bad"   # Lookup by title (local first, falls back to TMDb)
```
- Checks local database first. If not found by title, queries TMDb API, saves to DB, and displays.
- Shows release year, runtime, ratings (IMDb & TMDb), director, cast list, genres, file path, and size.
- Flags: `--format json`, `--format table`.

### 3. User Tagging (`cmd/movie_tag.go`)
```sh
movie tag add <media-id> <tag>       # e.g., movie tag add 1 favorite
movie tag remove <media-id> <tag>    # Removes specific tag
movie tag list                       # Shows all tags with counts
movie tag list <media-id>            # Shows tags on specific item
```
- Managed via `Tag` and `MediaTag` join tables.
- Unique constraints prevent duplicate tags on the same media item.

### 4. TV Season & Episode Tracking (`cmd/movie_tv*.go`)
```sh
movie tv seasons "Breaking Bad"        # Lists all seasons and episode counts
movie tv episodes "Breaking Bad" 1     # Lists episodes for Season 1
movie tv mark "Breaking Bad" S01E03    # Marks specific episode as watched
movie tv unmark "Breaking Bad" S01E03  # Reverts episode to pending
```
- Backed by `Season` and `Episode` tables.
- Episode code pattern `S(\d{1,3})E(\d{1,3})` parses input codes.

### 5. Watchlist Management (`cmd/movie_watch*.go`)
```sh
movie watch add <media-id>    # Adds item to watchlist ('to-watch')
movie watch done <media-id>   # Marks item as watched
movie watch undo <media-id>   # Reverts watched item to 'to-watch'
movie watch rm <media-id>     # Removes from watchlist
movie watch ls                # Lists pending watchlist
movie watch ls --watched      # Lists completed items
movie watch ls --all          # Lists all watchlist entries
```
- Persisted in `Watchlist` table with `AddedAt` and `WatchedAt` timestamps.

### 6. Recommendations & Discovery (`cmd/movie_suggest.go`, `cmd/movie_discover.go`)
```sh
movie suggest [count]    # Recommends movies/shows based on top library genres
movie discover           # Explores trending content from TMDb
```
- Calculates frequency of genres across the local library (`TopGenres`).
- Queries TMDb recommendations based on top-rated local titles.
- Categories: Movies, TV Shows, or Random.

### 7. Playback (`cmd/movie_play.go`)
```sh
movie play <media-id>
```
- Resolves `CurrentFilePath` from `Media` table.
- Verifies physical file existence on disk before attempting to launch.
- Invokes host OS default player:
  - Windows: `cmd.exe /c start ""`
  - macOS: `open <filepath>`
  - Linux: `xdg-open <filepath>`
