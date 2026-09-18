---
name: TV-show season/episode ingestion
description: v2.308.0 wires Season/Episode tables into live scan path. tmdb.GetTVSeason fetches /tv/{id}/season/{n}; cmd/movie_scan_tv_seasons.go::ingestTVSeasons walks 1..NumberOfSeasons and upserts via FK-safe db.InsertSeason / db.InsertEpisode. Triggered from linkScanMediaRelations when m.Type==TV && m.TmdbID>0. Per-season/episode failures isolated via errlog.Warn.
type: feature
---
**Trigger**: `cmd/movie_scan_process.go::linkScanMediaRelations` —
when the resolved Media is a TV show with a TMDb ID, calls
`ingestTVSeasons`.

**Flow**:
1. `resolveSeasonCount` → `client.GetTVDetails(tmdbID).Seasons`
2. Loop `n = 1..count`:
   - `client.GetTVSeason(tmdbID, n)` (TMDb `/tv/{id}/season/{n}`)
   - `db.InsertSeason` (FK-safe, returns canonical SeasonId)
   - For each episode: `db.InsertEpisode` (FK-safe)

**Files**:
- `tmdb/types.go` — added `TVSeason`, `TVEpisode`
- `tmdb/client.go` — added `GetTVSeason(tmdbID, seasonNumber int)`
- `cmd/movie_scan_tv_seasons.go` — new (≤120 lines, all funcs ≤15
  lines, params bundled in `tvIngestInput` struct to stay ≤3 args)
- `cmd/movie_scan_process.go` — wired into `linkScanMediaRelations`

**Side fix (v2.308.0)**:
`cmd/movie_contextmenu_telemetry.go` had a pre-existing build error
(`nullInt64Zero()` returned `sql.NullInt64` but `ActionSimpleInput.MediaID`
is `int64`). Replaced with literal `0`, removed the dead helper, dropped
the unused `database/sql` import.

**Verification**:
- `nix run nixpkgs#go -- build ./...` → clean
- `go test ./tmdb/... ./db/... ./cmd/...` → all pass
- `scripts/check-lastinsertid-anti-pattern.sh db` → ✅
- `scripts/check-boolean-naming.sh .` → ✅

**Cost**: 1 + N TMDb calls per TV show (1 details + N season fetches).
TMDb 40 req/s limiter already in place. Safe.
