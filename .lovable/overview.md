# Project Overview

**Name:** Movie CLI
**Type:** Go 1.22 CLI application (NOT a web app)
**Binary:** `movie` (`movie.exe` on Windows)
**Repo folder:** `movie-cli`
**Module:** `github.com/alimtvnetwork/movie-cli-v8`
**Updated:** 2026-04-20

## Purpose

Cross-platform CLI tool for managing a personal movie and TV show library. Scans local folders, cleans filenames, fetches TMDb metadata, stores in SQLite, organizes files.

## Key Architecture

- Pure-Go SQLite (`modernc.org/sqlite`), WAL mode, single `movie.db`
- TMDb API for metadata
- Cobra CLI framework
- Console-safe self-update via gitmap handoff pattern
- Data folder at `<binary-dir>/data/` (resolved via `os.Executable()`) containing `movie.db`, `log/`, `config/`, `thumbnails/`, `json/`

## Current Version

v2.135.0

## Quick Links

- [Plan](plan.md) — project roadmap
- [Memory Index](.lovable/memory/index.md) — all institutional knowledge
- [Strictly Avoid](.lovable/strictly-avoid.md) — forbidden patterns
- [Prompts](.lovable/prompt.md) — reusable AI prompts
