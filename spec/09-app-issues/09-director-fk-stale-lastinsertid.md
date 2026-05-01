# Issue #09: Director Link FK Constraint Failure (787) — Stale LastInsertId in ensureDirector

## Issue Summary

1. **What happened**: During `movie scan`, many entries logged
   `[WARN] movie_scan_process.go:117: Director link error for '<Title>': constraint failed: FOREIGN KEY constraint failed (787)`.
2. **Where** (module + file paths):
   - `db/director.go` — `ensureDirector()` (the buggy helper)
   - `cmd/movie_scan_process.go:117` — call site logging the FK error
   - Schema: `db/schema_tables.go` — `MediaDirector(DirectorId)` FK → `Director(DirectorId)`
3. **Symptoms and impact**:
   - Director relationships silently dropped for many movies (not linked in `MediaDirector`).
   - `movie info`, suggest, and search queries that join through `MediaDirector` return incomplete/missing director data.
   - Log noise on every scan once the Director table has any entries.
4. **How discovered**: User screenshot of `movie scan` output showing the warning on `Requiem for a Dream` and other titles.

## Root Cause Analysis

1. **Direct cause**: `ensureDirector()` used `res.LastInsertId()` to short-circuit the lookup after `INSERT OR IGNORE`:

   ```go
   res, _ := d.Exec("INSERT OR IGNORE INTO Director (Name) VALUES (?)", name)
   id, _ := res.LastInsertId()
   if id > 0 { return id, nil }            // ← bug
   // fallback SELECT only runs when id == 0
   ```

   With `INSERT OR IGNORE`, when the UNIQUE(Name) constraint conflicts, no row is inserted, but **`LastInsertId()` in `mattn/go-sqlite3` returns the last successful insert ROWID on that connection** — which can be the rowid of an unrelated table that the same connection just inserted into (e.g., `Genre`, `MediaGenre`, `Media`). That stale ID is then inserted into `MediaDirector.DirectorId`, where it does not exist in `Director.DirectorId`, triggering SQLite extended error **787 (`SQLITE_CONSTRAINT_FOREIGNKEY`)**.

2. **Contributing factors**:
   - Sibling helper `EnsureGenre()` in `db/genre.go` does it correctly (always SELECT after INSERT OR IGNORE) — `ensureDirector` diverged from the established pattern.
   - Parallel scan worker pool (`NumCPU*2`) increases the rate at which one connection performs many inserts across tables, making the stale `LastInsertId` almost guaranteed for any duplicate director name.
3. **Triggering conditions**: Any second-or-later occurrence of an existing director name during a scan, on a connection that has performed any other INSERT since opening.
4. **Why spec did not prevent it**: No spec rule mandated the canonical `INSERT OR IGNORE … then SELECT` pattern for ensure-helpers. Genre followed it by convention; Director did not.

## Fix Description

1. **Spec changes**: Add a coding guideline that `ensure<Entity>` helpers must always perform a `SELECT` after `INSERT OR IGNORE` and **must not** rely on `LastInsertId()` to detect existing rows.
2. **New rules or constraints**: Forbid `LastInsertId()` as a "row exists?" signal after `INSERT OR IGNORE`. Use either `RowsAffected()` (to know if a row was inserted) or always `SELECT` to fetch the canonical PK.
3. **Why it resolves root cause**: Always reading the PK back from the table guarantees the returned ID points to a real `Director` row, eliminating the stale-rowid → FK-787 chain.
4. **Config changes**: None.
5. **Diagnostics required**: None — fix is local to `ensureDirector()`.

## Prevention and Non-Regression

1. **Prevention rule**: All `ensure<Entity>(name)` helpers across `db/` follow the Genre pattern: `INSERT OR IGNORE` then unconditional `SELECT … WHERE Name = ?`.
2. **Acceptance criteria**: `movie scan` over a library where multiple movies share a director produces zero `Director link error … FOREIGN KEY constraint failed` warnings, and `MediaDirector` rows exist for every linked director.
3. **Guardrails or linting**: Code review checklist item — flag any `res.LastInsertId()` used as an existence check after `INSERT OR IGNORE`.
4. **Spec references**: `spec/01-coding-guidelines/03-coding-guidelines-spec/`, `db/genre.go` (canonical pattern).

## Done Checklist

- [x] Issue write-up created under `/spec/09-app-issues/`
- [x] `db/director.go` `ensureDirector()` rewritten to match Genre pattern
- [x] Version bumped (code change)
- [ ] Real-OS verification on the user's library
