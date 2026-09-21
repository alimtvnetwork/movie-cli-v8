---
name: movie-cli-configuration-and-admin
description: "Configuration management, database reset, error log audits, data export, and administrative maintenance in movie-cli-v8."
---

# Movie CLI Configuration and Admin Skill

## Overview

This subsystem governs configuration settings, database maintenance, diagnostics logs, library reset operations, data exports, and clean self-uninstallation for `movie-cli-v8`.

## Configuration Management (`cmd/movie_config.go`)

Configuration settings are persisted in the `Config` table (`ConfigKey`, `ConfigValue`) in `movie.db`.

### Supported Configuration Keys
- `movies_dir`: Default target directory for movie organization.
- `tv_dir`: Default target directory for TV shows.
- `archive_dir`: Default archive directory for selector bulk moves.
- `scan_dir`: Default scan directory when no path argument is provided.
- `tmdb_api_key`: TMDb v3 API key.
- `tmdb_token`: TMDb v4 Bearer Read Access Token.
- `page_size`: Pagination size for `movie ls` tabular outputs.

### CLI Commands
```sh
movie config                         # Displays all configured key-value pairs
movie config get <key>               # Retrieves value for specific key
movie config set <key> <value>       # Updates or inserts configuration value
```

## Library Reset & Cache Purging (`cmd/movie_reset*.go`)

Cleans up databases, thumbnails, and scan outputs across known folders without touching physical media files.
```sh
movie reset [flags]
```

### Safety Rules & Flags
- **Physical Media Protection:** Video files (`.mkv`, `.mp4`, etc.) are **NEVER** touched or deleted during a reset.
- `--dry-run`: Previews all files, tables, and folders that would be deleted.
- `--keep-config`: Preserves user settings in `Config` table (enabled by default).
- `--all`: Wipes global user caches in the home directory in addition to local databases.
- `--force`, `-f` / `--yes`, `-y`: Bypasses confirmation prompts for automated scripts.

## Diagnostic Logs (`cmd/movie_logs.go`, `db/errorlog.go`)

Errors, warnings, and unhandled exceptions are logged to the `ErrorLog` table in SQLite:
```sh
movie logs [flags]
```
- Filters logs by severity (`ERROR`, `WARN`, `INFO`), command context, source file, or timestamp.
- Captures function name, working directory, error message, and formatted stack trace.

## Data Export (`cmd/movie_export.go`)

Exports the full library metadata collection for external backup or spreadsheets:
```sh
movie export --format json > library.json
movie export --format csv > library.csv
```

## Database Maintenance (`cmd/movie_db*.go`)

Direct database administrative utilities:
```sh
movie db status        # Checks SQLite integrity, table counts, and size
movie db vacuum        # Rebuilds the database file, reclaiming unused space
movie db version       # Displays active schema migration version
```

## Self-Uninstall (`cmd/movie_self_uninstall.go`)

Performs clean removal of `movie-cli` from the host system:
```sh
movie self-uninstall
```
1. Unregisters OS context menus from the Windows Registry, macOS Services, or Linux desktop.
2. Removes deployment directory and PATH entries if configured.
3. Prompts whether to keep or wipe the `data/movie.db` library database.
