---
name: movie-cli-database-architecture
description: "SQLite database design, schema tables, relationships, migration pipelines, and CRUD patterns in movie-cli-v8."
---

# Movie CLI Database Architecture Skill

## Overview

The database subsystem of `movie-cli-v8` manages persistence for media items, collections, genres, directors, cast members, tags, watchlist entries, file moves, action logs, and system errors using a standalone SQLite 3 database operating in WAL mode.

## Database Location & Connection (`db/open.go`)

- **Primary Path:** `<binary-directory>/data/movie.db`
- **Fallback Resolution:** Reads `db_path` from user configuration or resolves relative to working directory.
- **SQLite Engine Pragmas:**
  - `PRAGMA journal_mode = WAL;` (Concurrent reads and atomic writes)
  - `PRAGMA foreign_keys = ON;` (Strict referential integrity and cascading deletes)
  - `PRAGMA busy_timeout = 5000;` (Prevents database locked errors during concurrent access)
  - `PRAGMA synchronous = NORMAL;` (High performance with write safety)

## Relational Schema & Table Architecture (`db/schema_tables.go`)

### Table Categories & Naming Standard
All database tables follow strict **PascalCase** naming. Primary keys MUST be `{TableName}Id` integer autoincrement columns. Booleans use integer flags (`0` or `1`) with strictly positive naming prefixes (`Is*`, `Has*`).

#### 1. Core Media Tables
- `Media`:
  - `MediaId` (INTEGER PRIMARY KEY AUTOINCREMENT)
  - `Title` (TEXT NOT NULL), `CleanTitle` (TEXT NOT NULL)
  - `Year` (SMALLINT), `Type` (TEXT CHECK(Type IN ('movie', 'tv')))
  - `TmdbId` (INTEGER UNIQUE), `ImdbId` (TEXT)
  - `Description` (TEXT), `Tagline` (TEXT), `TrailerUrl` (TEXT)
  - `ImdbRating` (REAL), `TmdbRating` (REAL), `Popularity` (REAL)
  - `LanguageId` (INTEGER REFERENCES Language(LanguageId))
  - `CollectionId` (INTEGER REFERENCES Collection(CollectionId))
  - `OriginalFileName` (TEXT), `OriginalFilePath` (TEXT), `CurrentFilePath` (TEXT)
  - `FileExtension` (TEXT), `FileSizeMb` (REAL), `Runtime` (INTEGER)
  - `ScannedAt` (TEXT), `UpdatedAt` (TEXT)
- `Collection`: `CollectionId`, `TmdbCollectionId`, `Name`, `Overview`, `PosterPath`, `BackdropPath`.

#### 2. Taxonomy & Entity Join Tables
- `Genre` & `MediaGenre`: `GenreId`, `Name` (UNIQUE); join: `MediaGenreId`, `MediaId`, `GenreId`.
- `Cast` & `MediaCast`: `CastId`, `Name`, `TmdbPersonId`; join: `MediaCastId`, `MediaId`, `CastId`, `Role`, `CastOrder`.
- `Director` & `MediaDirector`: `DirectorId`, `Name`, `TmdbPersonId`; join: `MediaDirectorId`, `MediaId`, `DirectorId`.
- `Tag` & `MediaTag`: User-defined tags (`TagId`, `Name`); join: `MediaTagId`, `MediaId`, `TagId`.
- `Language`: ISO language lookup (`LanguageId`, `Code`, `Name`).

#### 3. TV Series Hierarchies
- `Season`: `SeasonId`, `MediaId`, `SeasonNumber`, `TmdbSeasonId`, `Name`, `Overview`, `EpisodeCount`.
- `Episode`: `EpisodeId`, `SeasonId`, `EpisodeNumber`, `TmdbEpisodeId`, `Name`, `Overview`, `Runtime`, `IsWatched`, `WatchedAt`.

#### 4. Audit, History & Reversibility Tables
- `FileAction`: Lookup for action types (`FileActionId`, `Name`: `move`, `rename`, `scan`, `delete`, `popout`).
- `MoveHistory`: Tracks physical file relocations (`MoveHistoryId`, `MediaId`, `FileActionId`, `FromPath`, `ToPath`, `IsReverted`, `MovedAt`).
- `ActionHistory`: Full state audit log with JSON snapshot (`ActionHistoryId`, `FileActionId`, `MediaId`, `MediaSnapshot`, `Detail`, `BatchId`, `IsReverted`, `CreatedAt`).
- `Watchlist`: Personal watch queue (`WatchlistId`, `MediaId`, `TmdbId`, `Title`, `Year`, `Type`, `Status` ('to-watch'/'watched'), `AddedAt`, `WatchedAt`).
- `ScanFolder` & `ScanHistory`: Scanned folder paths and execution metrics (duration, files found, errors).

#### 5. Configuration & Diagnostics
- `Config`: Key-value configuration pairs (`ConfigKey` PRIMARY KEY, `ConfigValue`).
- `ErrorLog`: Structured diagnostic log (`ErrorLogId`, `Timestamp`, `Level`, `Source`, `Function`, `Command`, `WorkDir`, `Message`, `StackTrace`).

## Migration Pipeline (`db/migrate.go` & `db/migrate_v*.go`)

1. **Schema Versioning:** Tracked in table `SchemaVersion` (`Version INTEGER NOT NULL`).
2. **Version Sequence:**
   - `v1`: Initial tables (`Media`, `Genre`, `MoveHistory`, `ScanHistory`, `Config`).
   - `v2`: TV show support (`Season`, `Episode`), `Collection`, `Cast`, `Tag`.
   - `v3`: `ActionHistory` table, batch IDs, and JSON state snapshots.
   - `v4`: `Director`, `MediaDirector`, and IMDb ratings caching.
   - `v5`: `ErrorLog` system table and soft-deletion tracking.
   - `v6`: Performance indexes (`idx_media_clean_title`, `idx_media_current_path`, `idx_media_tmdb_id`).
   - `v7`: ImdbLookupCache table for DuckDuckGo/TMDb fallback resolution.
3. **Migration Rule:** Migrations run automatically on `db.Open()`. Every migration must be idempotent (`CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`).

## Query Conventions & Developer Rules

- **Return Types:** Always return typed model structs or slices; wrap database errors with `appfault.Wrap(msg, err)`.
- **Prepared Statements:** Parameterize all queries (`SELECT ... WHERE MediaId = ?`) to prevent SQL injection and enable statement caching.
- **Cascade Deletes:** Foreign keys must declare `ON DELETE CASCADE` for child/join tables (`MediaGenre`, `MediaCast`, `MediaTag`, `Season`, `Episode`) or `ON DELETE SET NULL` for non-destructive history.
