# Plan: Installer Post-Install & CLI UI Modernization Following GitMap
Plan ID: 23-installer-postinstall-and-cli-ui-modernization
Status: Completed
Date: 2026-09-22
Loops/Steps Budget: N=200 Steps

## Overview
Comprehensive overhaul of installer post-installation screens, movie detail cards, cleanup/move preview outputs, configuration key-value tables, and database reset safety diagnostics following the GitMap terminal design system (Unicode box frames, Cyan accents, Split-DB indicators, masked secrets, and aligned status chips).

## Tasks Executed

### Task-01: Post-Install and Quick Installer UI
- Target Files:
  - `install.ps1`
  - `install.sh`
  - `install-quick.sh`
- Accomplishments:
  - Added GitMap Unicode box banners (`┌──────────────────────────────────────────────────────────┐`) to post-installation summaries.
  - Added explicit SQLite Split-DB store locations (`movie.db` Primary Library DB and `cache.db` Ephemeral Cache DB in WAL mode).
  - Streamlined System Diagnostics with `[ok]` and `[warn]` status chips.
  - Enhanced Quick Start Commands guide with cyan bold headers, yellow command keys, and clear descriptions.
  - Verified `install.ps1` with PowerShell 5.1 UTF-8 BOM encoding and `install.sh`/`install-quick.sh` with `bash -n`.

### Task-02: Movie Info Card Modernization
- Target Files:
  - `cmd/movie_info.go`
  - `cmd/movie_info_table.go`
- Accomplishments:
  - Modernized default `movie info` output with `printMediaDetailCard` featuring GitMap Unicode card framing (`┌──────────────────────────────────────────────────────────┐`).
  - Added media type badges (`[Movie]`, `[TV Series]`), dual ratings with yellow star glyphs (`⭐ 8.8 (IMDb)  ⭐ 8.4 (TMDb)`), and release year.
  - Integrated storage and SQLite Split-DB indicators displaying `movie.db (Primary Library)`, file path, and formatted file size.
  - Modernized `movie info --format table` with `Store Tier` row and clean borders.

### Task-03: Clean and Move Preview UI
- Target Files:
  - `cmd/movie_cleanup.go`
  - `cmd/movie_move_selector.go`
- Accomplishments:
  - Overhauled `movie cleanup` preview with GitMap section headers (`── Stale Database Entries ──`), missing file paths, and `movie.db` store attribution.
  - Added dry-run preview guidance card (`── Preview Mode ──`) and prominent warning confirmation prompt.
  - Overhauled `movie move` preview with GitMap operation card, cyan arrows (`→`), and green `[ok]` success chips.

### Task-04: Config and Reset UI
- Target Files:
  - `cmd/movie_config.go`
  - `cmd/movie_reset_helpers.go`
- Accomplishments:
  - Modernized `movie config` output into a GitMap property card with cyan keys, dim labels, and masked API secrets (`••••••••`).
  - Overhauled `movie reset` with a prominent warning box (`┌── WARNING: SYSTEM RESET ──┐`) detailing Split-DB stores to be wiped (`movie.db` + `cache.db`).
  - Included `cache.db`, `cache.db-wal`, and `cache.db-shm` in reset targets to ensure complete Split-DB cleanup.
  - Modernized reset dry-run preview and post-reset summary outputs.

## Coding Guidelines Compliance
- Strict positive booleans (`isColor`, `hasTarget`, `hasRatings`, `hasCredits`, `hasSynopsis`, `isKeepConfig`).
- Zero explicit boolean checks (`== true` total ban).
- Vertical blank lines before `if`, after `}`, before `return`.
- Strict relative git paths in all markdown and code documentation.
- Functions kept within 8–15 lines.
