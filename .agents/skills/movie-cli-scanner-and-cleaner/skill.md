---
name: movie-cli-scanner-and-cleaner
description: "Architecture, algorithms, flags, and workflows for video scanning, filename cleaning, TV season extraction, and media reconciliation in movie-cli-v8."
---

# Movie CLI Scanner and Cleaner Skill

## Overview

The scanning and cleaning subsystem of `movie-cli-v8` discovers video files across local directories, parses messy filenames into clean metadata (title, release year, media type), fetches external TMDb details, generates thumbnails, builds interactive HTML reports, and reconciles disk changes against the local SQLite database.

## Core Packages and Source Files

- `cleaner/parse.go`: Video extension detection, 30+ regex junk patterns, release group stripping, year extraction, and TV pattern detection.
- `cmd/movie_scan.go`: Main `movie scan [folder]` command definition, CLI flags, output directory bootstrapping, and execution orchestration.
- `cmd/movie_scan_loop.go`: Scanning loop, directory traversal, file gathering, and concurrency dispatch.
- `cmd/movie_scan_collect.go`: Concurrency file collector and stat extraction.
- `cmd/movie_scan_hydrate.go`: Enriches scanned items with TMDb metadata, genres, cast, and thumbnails.
- `cmd/movie_scan_html.go`: Generates standalone interactive HTML reports (`.movie-output/report.html`).
- `cmd/movie_scan_thumb.go`: Downloads and caches poster/backdrop thumbnails locally.
- `cmd/movie_scan_tv_seasons.go`: Deep parsing of TV seasons (`S01E02`, `Season X`) and episode structures.
- `cmd/movie_scan_reconcile.go`: SmartRescan reconciliation detecting moved, renamed, or deleted files.
- `cmd/movie_scan_reverse.go`: DB-to-JSON reverse sync pass maintaining disk metadata consistency.
- `cmd/movie_rescan.go` & `cmd/movie_rescan_failed.go`: Re-evaluating previously failed or pending scan items.

## Filename Cleaner Architecture (`cleaner/parse.go`)

### Supported Video Extensions
```go
.mkv, .mp4, .avi, .mov, .wmv, .flv, .webm, .m4v, .ts, .vob, .ogv, .mpg, .mpeg, .3gp
```

### Cleaning Pipeline
1. **Extension Stripping:** Trims file extension and isolates base name.
2. **Media Type Detection:** Evaluates regex `(?i)S\d{1,2}E\d{1,2}|Season\s*\d+|Episode\s*\d+`. Matches classify as `"tv"`, otherwise `"movie"`.
3. **Year Extraction:**
   - Priority 1: Parenthesized 4-digit year `\((\d{4})\)` (e.g., `(2023)`).
   - Priority 2: Bare 4-digit year `\b((?:19|20)\d{2})\b` (1900–2099).
4. **Delimiter Normalization:** Converts dots (`.`) and underscores (`_`) to spaces.
5. **Dash Separation Splitting:** If filename has ` - `, cuts right side if identified as release group or audio/video spec junk.
6. **Multi-Pass Junk Removal:** Runs two sequential regex passes across 30+ patterns:
   - Resolutions: `1080p`, `720p`, `4k`, `2160p`, `uhd`, `fhd`, `hd`, `sd`.
   - Source/Rip: `bluray`, `bdrip`, `webrip`, `web-dl`, `dvdrip`, `remux`, `hdtv`.
   - Codecs: `x264`, `x265`, `hevc`, `avc`, `aac`, `ac3`, `dts`, `atmos`, `truehd`, `opus`.
   - Release Groups: `rarbg`, `yts`, `yify`, `eztv`, `sparks`, `fgt`, `ion10`, `psa`, `qxr`.
   - Editions: `extended`, `unrated`, `directors.cut`, `remastered`, `criterion`.
   - Audio/Subs: `dual`, `multi`, `eng`, `hindi`, `subbed`, `5.1`, `7.1`, `hdr10`, `dolby.vision`.
   - Enclosures: `[...]`, `(...)`, `{...}`.
7. **Whitespace Collapse:** Trims trailing junk dashes and collapses multiple spaces.

## Scanner Orchestration & CLI Usage

### Command Syntax & Flags
```sh
movie scan [folder] [flags]
```

- `--recursive`, `-r`: Recursively scans all subdirectories.
- `--depth`, `-d <N>`: Sets maximum subdirectory recursion depth (default `0` = unlimited).
- `--dry-run`: Previews discovered files and clean titles without writing to DB or disk.
- `--format <format>`: Output format: `default`, `table`, or `json`.
- `--rest`: Automatically boots the local REST API server and opens the browser report after scanning.
- `--port <port>`: Port for REST server (default `8086`).
- `--watch`, `-w`: Enters continuous polling mode watching for added/removed video files.
- `--interval <sec>`: Polling interval in seconds for `--watch` mode (default `10`).
- `--workers <N>`: Parallel scan worker concurrency (`0` = auto: `NumCPU * 2`, max `32`).
- `--keep-logs`: Preserves existing logs in `.movie-output/logs/` instead of truncating on start.
- `--no-open`: Prevents opening default web browser when report is generated.
- `--no-reconcile`: Disables SmartRescan file move/deletion reconciliation.
- `--no-reverse-sync`: Skips DB-to-JSON reverse sync pass.
- `--reverse-sync-only`: Skips forward scan and TMDb network calls, only refreshing disk JSON metadata from SQLite.

### Output Artifacts (`.movie-output/`)
During scan runs, a `.movie-output/` folder is maintained within the target root directory:
- `.movie-output/report.html`: Standalone HTML dashboard rendering media cards, stats, and search.
- `.movie-output/thumbnails/`: Downloaded TMDb poster and backdrop image files (`w500` format).
- `.movie-output/logs/`: Per-session execution logs (`scan.log`, `error.log`).

## SmartRescan & Reconciliation

1. **Missing File Detection:** Checks every existing `Media` record in SQLite against disk paths.
2. **Move Reconciliation:** Matches missing records against newly scanned files using file size, duration, and clean title hashes.
3. **Database Update:** Updates `CurrentFilePath` and logs an entry in `action_history` without losing metadata, genres, or user tags.
4. **Stale Records:** Stale entries whose physical files no longer exist can be inspected via `movie cleanup` and pruned with `movie cleanup --remove`.
