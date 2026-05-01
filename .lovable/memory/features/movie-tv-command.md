---
name: movie tv subcommand surface
description: v2.309.0 adds `movie tv [seasons|episodes|mark|unmark]` so the Season/Episode data populated by scan (v2.308.0) is queryable from the CLI. Also renames db.MarkEpisodeUnwatched→MarkEpisodePending per boolean-naming rule, adds db.FindEpisodeByMediaAndCode helper.
type: feature
---
**Surface**:
- `movie tv seasons <id-or-title>` — list all seasons
- `movie tv episodes <id-or-title> <seasonNumber>` — list episodes
- `movie tv mark <id-or-title> SxxExx` — mark watched
- `movie tv unmark <id-or-title> SxxExx` — mark pending

**Files**:
- `cmd/movie_tv.go` — parent + 4 subcommand definitions, helpers
  (`resolveTvMedia`, `parseEpisodeCode`, `formatYearOrDash`)
- `cmd/movie_tv_run.go` — Run handlers split out for ≤200-line cap
- `db/season.go` — renamed `MarkEpisodeUnwatched` → `MarkEpisodePending`
  (boolean-naming rule); added `FindEpisodeByMediaAndCode`
- `cmd/root.go` — added `movie tv` examples to help text

**Resolver**: reuses `resolveMediaByQuery` (id-or-title fuzzy match);
fails fast if resolved Media is not type=tv.

**Vocabulary**: chose `mark` / `unmark` instead of `watched` /
`unwatched` so the CLI verb itself avoids the negative prefix
pattern, even though function names aren't booleans.

**Verification**:
- `nix run nixpkgs#go -- build ./...` → clean
- `go test ./db/... ./cmd/...` → all pass
- `movie tv --help` → renders all 4 subcommands
- Both lint guards (`check-lastinsertid-anti-pattern.sh`,
  `check-boolean-naming.sh`) → ✅
