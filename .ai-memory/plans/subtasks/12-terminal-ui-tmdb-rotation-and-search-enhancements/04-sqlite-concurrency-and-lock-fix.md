# Subtask 04: SQLite Concurrency & SQLITE_BUSY Lock Elimination

> **Parent Plan:** `.ai-memory/plans/pending/12-terminal-ui-tmdb-rotation-and-search-enhancements.md`  
> **Status:** Complete  
> **Target Files:**
> - `db/open.go`
> - `db/errorlog.go`
> - `cmd/movie_scan.go`

---

## 1. Problem Statement

1. During parallel scan execution (up to 32 worker threads), concurrent goroutines log warnings/errors via `errlog.Error` and `errlog.Warn`.
2. This triggers the error log database writer (`database.InsertErrorLog`) simultaneously from multiple goroutines.
3. Because SQLite only allows one write transaction at a time and `modernc.org/sqlite` with multiple pool connections can experience write conflicts, this produces:
   `⚠️ Could not write error to DB: insert error log: database is locked (5) (SQLITE_BUSY)`
4. The user specifically raised concerns about database integrity: "are you populating the SQLite database? You have to confirm because this is how we are going to save later on." All scanned items must be reliably inserted and indexed in `movie.db`.

---

## 2. Proposed Changes

### A. Connection Pool Configuration (`db/open.go`)
- In `openAndConfigureDB`:
  - Set `conn.SetMaxOpenConns(1)`: Serializes database access through a single connection pool handle, which is the recommended practice for SQLite in Go to completely prevent internal multi-connection lock contention.
  - Set `PRAGMA busy_timeout = 10000;` (10 seconds timeout).

### B. Exponential Backoff Retry on Busy (`db/errorlog.go`)
- In `InsertErrorLog`:
  - Wrap write execution in a retry loop (up to 5 attempts) with exponential backoff (25ms, 50ms, 100ms, 200ms).
  - Check error for "locked" or "busy" and retry safely.

### C. Safe Asynchronous DB Logger Guard (`cmd/movie_scan.go`)
- In `initScanLogger`:
  - Guard the DB logger callback with a mutex or non-blocking buffer to avoid worker thread deadlocks.
  - Suppress duplicate console noise if a DB error log write fails after retries.

### D. SQLite Media Population Integrity Check
- Verify that both TMDb-matched items and local-only fallback items are saved to the `Media` table with all metadata, file sizes, and paths populated.

---

## 3. Verification Criteria
- Multi-threaded scanning runs without producing any `database is locked (5) (SQLITE_BUSY)` warnings.
- All scanned media entries are present and verified in `movie.db`.
