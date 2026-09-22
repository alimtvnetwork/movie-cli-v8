# Plan: CLI UI and Quick Installers Polish
Plan ID: 22-cli-ui-and-quick-installers-polish
Status: Completed
Date: 2026-09-22

## Overview
Comprehensive aesthetic and visual overhaul of quick installers, quick uninstallers, scan summary activity, library statistics dashboard, and interactive search/history tables following the GitMap terminal design system (Unicode box frames, Cyan accents, Split-DB indicators, star ratings, and aligned status chips).

## Tasks Executed

### Task-01: Quick Installers and Uninstallers Polish
- Target Files:
  - `uninstall-quick.ps1`
  - `uninstall-quick.sh`
- Accomplishments:
  - Upgraded PowerShell and Bash quick uninstallers with GitMap-inspired Unicode box banners, cyan accents, and clear status chips (`[ok]`, `[warn]`).
  - Added explicit SQLite Split-DB store removal reporting (`movie.db` Primary Library DB and `cache.db` Ephemeral Cache DB).
  - Preserved UTF-8 BOM encoding on `uninstall-quick.ps1` to ensure correct rendering in Windows PowerShell 5.1.
- Verification:
  - `bash -n install-quick.sh uninstall-quick.sh` passed cleanly.

### Task-02: Scan Activity & Post-Scan Summary UI
- Target Files:
  - `cmd/movie_scan_helpers_print.go`
- Accomplishments:
  - Overhauled `printScanCounts` with GitMap section headers (`── Scan Summary ──`), cyan bullet points, and aligned counts for Added, Updated, Unchanged, TV Series, and TMDB Hydrated.
  - Overhauled `printScanOutputFiles` with SQLite Split-DB store indicators for `movie.db` Primary Library DB and `cache.db` Ephemeral Cache DB.
  - Overhauled `printScanGuidanceCard` with GitMap Unicode box framing (`┌──────────────────────────────────────────────────────────┐`), cyan bold command highlights, and helpful next steps.
- Verification:
  - `golangci-lint run ./cmd/...` exited with code 0.

### Task-03: Stats Dashboard Modernization
- Target Files:
  - `cmd/movie_stats.go`
  - `cmd/movie_stats_table.go`
- Accomplishments:
  - Upgraded default `movie stats` to render styled metric cards:
    - Library Overview: Total Movies, Total TV Shows, Total Media.
    - Storage & Filesystem: Total Size, Largest File, Smallest File, Average File Size.
    - SQLite Split-DB Stores: Master and Cache disk footprints, WAL status.
    - Top Genres: Distribution bar chart rendered with cyan blocks.
    - Ratings Overview: Star glyphs (`⭐ 7.5 / 10`) and yellow accents.
  - Modernized `cmd/movie_stats_table.go` with Unicode table borders (`┌─┬─┐`, `├─┼─┤`, `└─┴─┘`), cyan headers, Split-DB rows, and star ratings.
- Verification:
  - `golangci-lint run ./cmd/...` exited with code 0.

### Task-04: Search and History Tables
- Target Files:
  - `cmd/movie_search_table.go`
  - `cmd/movie_history_table.go`
- Accomplishments:
  - In `cmd/movie_search_table.go`: Modernized table headers with cyan styling, added star ratings (`⭐ 7.5`), aligned TMDb IDs, and clean Unicode dividers.
  - In `cmd/movie_history_table.go`: Added cyan column headers, green `[ok]` and yellow `[rev]` status chips with exact visible padding, and clean dividers.
- Verification:
  - `golangci-lint run ./cmd/...` exited with code 0.

## Coding Guidelines Compliance
- Strict positive booleans (`isColor`, `isReverted`, `hasRatings`).
- Zero explicit boolean checks (`== true` total ban).
- Vertical blank lines before `if`, after `}`, before `return`.
- Strict relative git paths in all markdown and code documentation.
- Functions kept within 8–15 lines.
