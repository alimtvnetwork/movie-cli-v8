# Issue 10 — Season/Episode Latent Stale LastInsertId

**Status**: Fixed in `v2.305.0`
**Date (UTC+8)**: 01-May-2026
**Severity**: High (latent — would manifest the moment TV ingestion is wired up)
**Related**: spec/09-app-issues/09-director-fk-stale-lastinsertid.md

## Audit Trigger

After fixing the Director FK 787 bug we audited every `INSERT OR IGNORE`
and `ON CONFLICT DO UPDATE` in `db/` for the same anti-pattern: trusting
`res.LastInsertId()` after a statement that may not have inserted a new
row.

## Findings

### ✅ Already correct
- `db/genre.go` — `EnsureGenre` SELECTs the canonical `GenreId` after
  `INSERT OR IGNORE`. Reference implementation.
- `db/director.go` — fixed in v2.304.0 to follow the genre pattern.
- `db/tags.go`, `db/seed.go`, `db/migrate_v4.go`, `db/migrate_v5.go`,
  `db/history.go` — use `INSERT OR IGNORE` but **discard** the result.
  No FK risk.
- `db/media.go`, `db/action_history.go`,
  `db/reconciliation_history.go` — use plain `INSERT` (no `OR IGNORE`,
  no `ON CONFLICT`). `LastInsertId()` is reliable here.

### ❌ Latent bug — `db/season.go`
Both `InsertSeason` and `InsertEpisode` use `ON CONFLICT DO UPDATE` and
then return `res.LastInsertId()`:

- `InsertSeason` had a fallback SELECT only when `id == 0`. In a
  parallel-scan scenario, `LastInsertId()` after the UPDATE branch can
  return a non-zero rowid that belongs to a **different** table on the
  same connection (e.g. the most recent `MediaGenre` insert). The
  function would then return that stale rowid as if it were the
  `SeasonId`, and any subsequent `InsertEpisode(SeasonId=stale)` call
  would either FK-fail (787) or — worse — silently attach episodes to
  the wrong season.
- `InsertEpisode` had **no** SELECT fallback at all. Same hazard,
  guaranteed to bite the moment two goroutines call it concurrently.

These functions are not wired into the live scan path yet (no callers
in `cmd/`), so the bug never manifested in production. We fixed it
proactively while the lesson from issue 09 is fresh.

## Fix (v2.305.0)

Both `InsertSeason` and `InsertEpisode` now follow the genre /
director pattern unconditionally:

1. Run the `INSERT … ON CONFLICT DO UPDATE` purely for its side
   effect.
2. Discard `res.LastInsertId()`.
3. `SELECT` the canonical PK back using the natural unique key
   (`(MediaId, SeasonNumber)` for seasons, `(SeasonId, EpisodeNumber)`
   for episodes).

This makes the helpers safe under parallel scans and removes the
last instance of the stale-rowid anti-pattern in `db/`.

## Rule (added to project conventions)

> Any helper that runs `INSERT OR IGNORE` or
> `INSERT … ON CONFLICT DO UPDATE` against a table whose PK is later
> used as a foreign key MUST `SELECT` the canonical PK back using the
> natural unique key. Never trust `res.LastInsertId()` in those cases.
> Reference implementation: `db/genre.go::EnsureGenre`.
