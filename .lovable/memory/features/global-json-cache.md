---
name: Global JSON cache
description: Per-user cache mirrors per-folder sidecars by TmdbID; survives folder moves and OS reinstalls
type: feature
---

Cross-folder JSON cache implemented in `cmd/movie_global_cache.go` (v2.302.0).

**Layout:** `~/.movie/cache/json/{movie|tv}/<tmdbid>.json` (uses `os.UserHomeDir()`).

**Write path:** `writeMediaJSON` (cmd/movie_scan_json.go) calls
`MirrorToGlobalCache(data)` after every successful per-folder sidecar write.
Idempotent overwrite. Best-effort — silently no-ops on FS errors.

**Read path:** SmartRescan hydration (`readSidecarFile` →
`enrichFromGlobalCache`) overlays cached metadata fields (Description,
Genre, Director, CastList, ImdbID, ImdbRating, TmdbRating) when the local
sidecar lacks them. Path/file fields stay local.

**Disable:** `MOVIE_NO_GLOBAL_CACHE=1` (env). Both `MirrorToGlobalCache`
and `LookupGlobalCache` no-op when set or when TmdbID ≤ 0.

**Why TmdbID-keyed:** stable identifier across folder moves, file renames,
and OS reinstalls. A movie discovered in `~/Movies` enriches a re-scan in
`~/Drive/Films` even with zero TMDb calls.

**Future expansion ideas:** `movie cache global ls/clear/import` admin
subcommand; filename → TmdbID guess to skip TMDb when DB miss + cache hit.
