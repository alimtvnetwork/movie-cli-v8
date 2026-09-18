---
name: Season/Episode stale LastInsertId audit
description: v2.305.0 audit fix — InsertSeason/InsertEpisode used ON CONFLICT DO UPDATE then trusted LastInsertId; latent FK/wrong-row bug under parallel scans, now SELECT canonical PK back
type: issue
---
Latent bug found while auditing `db/` for the same anti-pattern as
issue 05 (Director FK 787).

**Files**: `db/season.go` — `InsertSeason`, `InsertEpisode`
**Symptom (would have been)**: Episode rows attached to the wrong
SeasonId, or `Episode.SeasonId` FK error 787 under parallel TV scans.
**Why it never fired**: no live callers wired into the scan path yet.
**Fix (v2.305.0)**: Both helpers now SELECT the canonical PK back
using their natural unique key, mirroring `db/genre.go::EnsureGenre`.

**Rule**: Any `INSERT OR IGNORE` or `INSERT … ON CONFLICT DO UPDATE`
whose PK is used as a foreign key MUST SELECT the canonical PK back.
Never trust `LastInsertId()` in those cases. Reference impl:
`db/genre.go::EnsureGenre`.

Audited & confirmed clean: `db/genre.go`, `db/director.go` (v2.304.0),
`db/tags.go`, `db/seed.go`, `db/migrate_v4.go`, `db/migrate_v5.go`,
`db/history.go`, `db/media.go`, `db/action_history.go`,
`db/reconciliation_history.go`.

See `spec/09-app-issues/10-season-episode-stale-lastinsertid-audit.md`.
