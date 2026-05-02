# Changelog

All notable changes to this project will be documented in this file.

## v2.320.0

### Changed
- **Version bump** — minor version bumped to v2.320.0 to roll up recent lint, formatting, and documentation fixes (golangci-lint cleanup in v2.319.0, README screenshot redaction in v2.318.0, legacy-name purge from temp/log paths in v2.317.0).

## v2.230.0

### Changed
- **Module path renamed** — `github.com/alimtvnetwork/movie-cli-v8` → `github.com/alimtvnetwork/movie-cli-v8` across the entire project (147 files: Go imports, `go.mod`, README, CI workflows, install scripts, specs, memory).
- All GitHub URLs (`github.com/alimtvnetwork/movie-cli-v8` → `…/movie-cli-v8`) updated in README badges, install one-liners, release-asset URLs, and bootstrap scripts.
- CI old-module-path guard widened to ban `movie-cli-v[23456]` (was `v[2345]`).

### Migration
- Run `go mod tidy` after pulling to refresh the module cache.
- Local clones tracking the old remote should update their `origin` URL: `git remote set-url origin https://github.com/alimtvnetwork/movie-cli-v8.git`.

### Documentation
- Restored historical accuracy in earlier CHANGELOG entries (v2.93.0, v2.96.0, v2.117.0, v2.128.4, v2.106.0, v2.0.0) and `spec/12-ci-cd-pipeline/05-ci-cd-issues/06-release-missing-asset-404.md`: prior renames now correctly read v1→…→v5→v6 instead of being collapsed into "v7→v7".

## v2.135.0

### Removed
- **All legacy-name references purged from the entire codebase.** Stripped the legacy DB cleanup (`legacyDBFiles` and `removeLegacyDB` removed from `db/open.go`) and the legacy binary sweep (`legacyBaseNames` slice and its `cleanDir` loop removed from `updater/cleanup.go`). Old database files and handoff binaries from the previous project name are no longer auto-deleted — users with leftover artifacts must remove them manually.
- Updated CI guard in `.github/workflows/ci.yml` to ban the legacy term across the **entire repo with no allowlist** (previously db/open.go, updater/cleanup.go, CHANGELOG.md, and .lovable/* were exempt). Any reintroduction now fails the build immediately.
- Cleaned all historical CHANGELOG entries, `.lovable/overview.md`, and the CI guard's own help text of the legacy term.

## v2.134.0

### Added
- **CI guard against old module paths (v2/v3/v4)** — new `Old module path regression guard` step in `.github/workflows/ci.yml`. Greps the entire repo for any `movie-cli-v8`, `movie-cli-v8`, or `movie-cli-v8` reference and hard-fails the build if found. Current module is `github.com/alimtvnetwork/movie-cli-v8`; any stale import or documentation reference to the old iterations is a regression.

## v2.133.0

### Added
- **CI guard against legacy-name regression** — new guard step in `.github/workflows/ci.yml` (runs in the `lint` job, right after the Acronym MixedCaps guard). Hard-fails the build if any reference to the previous project name reappears.

## v2.132.0

### Changed
- **Reverted branding back to "movie"** — binary name is now `movie` (`movie.exe` on Windows), database file is `movie.db`. Repository folder name remains `movie-cli`. Data folder layout unchanged: `<binary-dir>/data/{movie.db,log/,config/,thumbnails/,json/}`.
  - `db/open.go`: `dbFile = "movie.db"`.
  - `Makefile`: `BINARY_NAME=movie`.
  - All `spec/` files, mermaid diagrams, `.lovable/memory/*`, `.lovable/overview.md`, `.lovable/strictly-avoid.md`, and `CONTRIBUTING.md` updated to use `movie` and `movie.db` everywhere.

## v2.130.1

### Changed
- **README Quick Start and Installation sections** — both now show **two** clearly separated install variants: a "latest release" one-liner using `/releases/latest/download/...` (auto-tracks newest) and a "pinned to vX.Y.Z" one-liner using `/releases/download/v2.130.0/...` (frozen, exact version forever). Each variant has its own header, Windows + Unix one-liners, and an explanatory note. Added a "When to use which" callout in both sections steering users to **latest** for personal machines and **pinned** for CI / Dockerfiles / onboarding docs / rollbacks. The pinned section explicitly notes that `PINNED_VERSION` is baked into the script and links to the contract at `spec/12-ci-cd-pipeline/06-version-pinned-install-scripts.md`.

## v2.130.0

### Added
- **CI enforcement of the version-pinning contract** — new `Enforce version-pinning contract on install scripts` step in `.github/workflows/release.yml`, runs immediately after `install.ps1` and `install.sh` are generated and BEFORE the release upload. It hard-fails the workflow if either generated script contains any forbidden string (`releases/latest/`, `bootstrap.sh`, `bootstrap.ps1`) or if the `$PinnedVersion` / `PINNED_VERSION` literal is missing or does not equal the resolved release version. Errors are surfaced as `::error file=...::` annotations with the offending line numbers and a pointer to spec 06. This makes spec/12-ci-cd-pipeline/06-version-pinned-install-scripts.md unbreakable in CI — the contract can no longer be silently regressed by future edits to the script generators.

## v2.129.0

### Documentation
- **New `spec/12-ci-cd-pipeline/06-version-pinned-install-scripts.md`** — formal contract spec stating that `install.ps1` and `install.sh` attached to a GitHub Release MUST install that exact release version. Documents the `VERSION_PLACEHOLDER` → literal `$PinnedVersion` / `$PINNED_VERSION` substitution, forbids `releases/latest/`, forbids any sibling-repo probing or `bootstrap.*` delegation inside the per-release scripts, and lists acceptance criteria + prevention rules so future AI edits cannot regress this guarantee.
- **`spec/12-ci-cd-pipeline/README.md`** — added the new doc to the documents table; corrected the "Release trigger" row to reflect that `v*` tag triggers were removed in v2.128.5.

### Verified (no code change needed)
- `.github/workflows/release.yml` already generates both install scripts with `PINNED_VERSION`/`$PinnedVersion` baked in via `sed` substitution at release time, and the release page one-liner already references `/releases/download/$VERSION/install.{ps1,sh}` (NOT `/latest/`). The pinning contract was already implemented; this version formalises it as a spec so it cannot be silently regressed.

## v2.128.7

### Changed
- **README Quick Start install blocks** — moved OS labels out of the code blocks (no more `# Windows (PowerShell)` / `# Linux / macOS` comments inside copyable snippets) and into proper `###` section headers. Each install command is now a clean one-liner that copy-pastes without a leading comment.

## v2.128.6

### Added
- **MIT License** — project now licensed under MIT. Added `LICENSE` file with full MIT text (Copyright (c) 2024-2025 Alim TV Network).
- **Go Report Card badge** — added to README alongside existing CI and version badges.

### Changed
- **README license badge** — updated from `license-Private-red` to `license-MIT-blue` linking to the LICENSE file.

## v2.128.5

### Fixed
- **Release dual-trigger race that produced partial v2.97.0 release** — investigation of the v2.97.0 missing-archives incident found TWO workflow runs fired for the same release: branch push `release/v2.97.0` (Run 24534322512, succeeded with 6 archives) and tag push `v2.97.0` (Run 24534323295, fired ~2s later because `softprops/action-gh-release` creates the tag, then **failed at the publish step after partially overwriting** the already-published assets, leaving 3 of 6 archives missing). Concurrency group `release-${{ github.ref }}` did not protect because the two refs (`refs/heads/release/v2.97.0` vs `refs/tags/v2.97.0`) hashed to different groups.
- **`.github/workflows/release.yml`** — removed the `tags: ["v*"]` trigger entirely. The release tag is **created by** `softprops/action-gh-release` at publish time, so triggering on it produces a self-inflicted second run. Workflow now triggers ONLY on `release/**` branch pushes. Updated the `Resolve version` step to reject any non-`release/**` ref with a clear error.

### Documentation
- **New `spec/12-ci-cd-pipeline/05-ci-cd-issues/07-release-dual-trigger-race.md`** — full RCA with workflow run IDs, evidence table, root cause walkthrough, the fix, prevention rules, and recovery procedure for partially-uploaded releases.

## v2.128.4

### Fixed
- **Install script 404 on `irm | iex` for v2.97.0** — user reported `Download failed: Not Found` when running the documented one-liner. Root cause: the v2.97.0 GitHub Release was published with a partial asset set (Windows amd64/arm64 + linux-amd64 archives never uploaded; `install.ps1` and `checksums.txt` were uploaded). The release workflow had no guard against partial uploads.
- **Repo-root `install.ps1`** — `RepoUrl` was pointing at the deleted `movie-cli-v8.git`. Updated to `movie-cli-v8.git` so the build-from-source fallback actually works.

### Changed
- **`.github/workflows/release.yml`** — added `Verify all 6 archives are present` step after compress/checksum and BEFORE the GitHub Release upload. Enumerates the 6 expected filenames and exits non-zero with `::error::` annotations if any are missing. Prevents future partial releases.
- **Generated `install.ps1`** — 404-aware error handler. When the binary archive returns HTTP 404, prints the exact missing URL, an explanation that this is a publisher-side issue, and two recovery commands (try a different release tag; build from source via `git clone … && ./run.ps1`).

### Documentation
- **New `spec/12-ci-cd-pipeline/05-ci-cd-issues/06-release-missing-asset-404.md`** — full RCA, prevention rule, and acceptance criteria for the partial-release class of issues.

## v2.128.3

### Fixed
- **Acronym MixedCaps regression in `doctor/json.go` + `cmd/doctor.go`** — renamed `JSONReport`→`JsonReport`, `JSONFinding`→`JsonFinding`, `PrintJSON`→`PrintJson`, `toJSON`→`toJson`, `toJSONFindings`→`toJsonFindings`, `emitJSON`→`emitJson`, `doctorJSON`→`doctorJson` to comply with the project-specific MixedCaps rule (`spec/01-coding-guidelines/03-coding-guidelines-spec/03-golang/09-acronym-naming.md`). The user-facing `--json` flag and JSON wire format are unchanged.

### Documentation
- **New `spec/12-ci-cd-pipeline/05-ci-cd-issues/` folder** — granular per-incident log of every CI lint failure encountered, one file per issue. Each file documents symptom, trigger, root cause, fix pattern, prevention rule, and version history. AI/contributors must read this folder before fixing similar errors.
  - `00-overview.md` — folder convention + index table
  - `01-misspell-british-american.md` — misspell linter (US locale), 22-entry British→American table
  - `02-gofmt-struct-tag-padding.md` — gofmt rejects hand-padded struct tags
  - `03-gofmt-doc-list-indent.md` — gofmt requires 2-space indent for numbered lists in doc comments
  - `04-govet-fieldalignment.md` — struct field order rules to minimise padding
  - `05-acronym-mixedcaps.md` — project-specific MixedCaps rule (inverts Effective Go), exemption process for stdlib interface overrides
- **Memory updates**: `mem://index.md` Core gained two new rules (American English only, MixedCaps acronyms) plus a pointer to the new ci-cd-issues folder. `mem://constraints/acronym-mixedcaps` history extended with v2.128.3 regression entry.

## v2.128.2

### Fixed
- **golangci-lint clean (round 2)** — three remaining lint failures from v2.127.0/v2.128.1:
  - `doctor/json.go`: gofmt — collapsed extra padding in struct tag column.
  - `doctor/json.go`: govet fieldalignment — reordered `JSONReport` (strings, slice, bools last) to drop padding from 80 → 72 pointer bytes.
  - `doctor/diagnose.go`: misspell — `optimised` → `optimized`.

## v2.128.1

### Fixed
- **golangci-lint clean** — fixed five lint failures introduced across v2.125.0–v2.128.0:
  - `doctor/checks.go`: gofmt — collapsed double-spaces in const block alignment.
  - `doctor/diagnose.go`: gofmt — normalised numbered-list indent in package doc comment (3-space → 2-space) and converted the file-prefix doc comment to a proper `Package doctor` comment.
  - `doctor/diagnose.go`: govet fieldalignment — reordered `Report` struct fields (strings first, slice last) to drop padding from 64 → 56 pointer bytes.
  - `cmd/doctor.go`: misspell — `catalogued` → `cataloged`.
  - `updater/self_replace.go`: misspell — `Behaviour` → `Behavior`.

## v2.128.0

### Added
- **`movie update` now auto-runs `movie doctor --fix` when preflight detected a fixable mismatch.** After a successful update, if the up-front preflight pass found a fixable issue (PATH/deploy mismatch, version drift, stale workers), the updater immediately re-diagnoses and runs the standard fix pipeline — calling `self-replace` automatically so users no longer have to bootstrap manually after a stuck-handoff scenario. If the post-update state is already clean (the update happened to land on the right binary), the auto-fix is skipped with a one-line note.

### Fixed
- **Removed `updater → doctor` import** introduced in v2.126.0 that would have caused a compile-time import cycle (`doctor` already imports `updater` for `SelfReplace`). Preflight orchestration moved to `cmd/update.go`, which already sits above both packages. v2.126.0/v2.127.0 functionality is preserved; only the wiring layer changed.

## v2.127.0

### Added
- **`movie doctor --json`** — emits the doctor report as a stable JSON document (`schema: "movie-doctor/v1"`) for scripting and CI pipelines. Includes resolved paths (`deploy_source`, `active_binary`, `deploy_dir`), aggregate flags (`has_errors`, `has_fixable`), and a `findings[]` array with `id`, `title`, `severity`, `detail`, `fix_hint`, `is_fixable`.
- **Documented exit codes for `movie doctor`**: `0` = all OK, `2` = errors found, `3` = fixable warnings only (no errors). Applies to both human and JSON output modes. Previously fixable warnings exited 0 — CI pipelines can now branch on the warning state.
- New `doctor/json.go` with `Report.PrintJSON()` and stable `JSONReport`/`JSONFinding` wire types (snake_case fields).

## v2.126.0

### Added
- **`movie update` now runs `movie doctor`'s checks up front** — before the handoff copy is created and before `run.ps1` is invoked, the updater calls `doctor.Preflight()` and prints any PATH/deploy mismatches, missing PATH entries, stale `*-update-*` workers, or version drift. Users see the warnings *before* a full build cycle, with a hint to run `movie doctor --fix` afterwards. Preflight failures are non-fatal (the update still proceeds) but they make the v2.97.0 → v2.121.0 stuck-binary class of problems impossible to miss.
- New `doctor/preflight.go` with a compact banner printer that reuses the standard finding renderer; OK passes print a single `[ OK ] Preflight checks passed` line.

## v2.125.0

### Added
- **`movie doctor` command** — diagnostic that surfaces the exact failure modes from the v2.97.0 → v2.121.0 stale-handoff loop, up front and in one place:
  - Active PATH binary vs `powershell.json` `deployPath` mismatch
  - Deploy directory missing from `$PATH`
  - Stale `*-update-*` handoff workers left on disk
  - Version drift between the active and the deployed binary
- **`movie doctor --fix`** — auto-repairs fixable findings: calls `self-replace` for binary mismatches/version drift, sweeps stale workers (best-effort, locked files are skipped silently), and prints (never auto-applies) a PowerShell one-liner to add the deploy directory to the User PATH. Re-runs the diagnose pass after fixing so the user sees the post-repair state immediately.
- New `doctor/` package: `diagnose.go`, `checks.go`, `paths.go`, `workers.go`, `report.go`, `fix.go`. All functions ≤15 lines, files ≤200 lines, errors via `apperror.Wrap`.
- Output uses the v2.123.0 indent scheme (0/1/2/3 spaces, `[ OK ]`/`[WARN]`/`[ERR ]`/`[FIX ]` tags).

## v2.124.0

### Documentation
- **Full failure-chain RCA published** for the v2.97.0 → v2.121.0 stale-handoff loop. Documents all 5 independent bugs (dual `bin-run` dirs, deployPath trust, stale-worker `-TargetBinaryPath` gap, PATH-sync loop attacking the live parent, stderr-as-NativeCommandError), why v2.118.0/v2.119.0/v2.120.0 fixes were invisible in production, the v2.121.0 root-cause fix, the v2.122.0 `self-replace` bootstrap, and the v2.123.0 output polish. Added as `spec/09-app-issues/08-updater-stale-handoff-loop-full-rca.md` and mirrored to `.lovable/memory/issues/07-updater-stale-handoff-loop-full-rca.md`. Issue tracker overview updated.

## v2.122.0

### Added
- **`movie self-replace` command** — one-shot bootstrap that atomically copies the freshly deployed binary over the active PATH `movie` using rename-first semantics (which works even while the active binary is loaded by another process on Windows). Defaults to `--from <deployPath>/<binaryName>` from `powershell.json` and `--to <active 'movie' on PATH>`. Use this to break out of any stuck-update loop where deployPath and the PATH-resolved binary live in different directories: `movie self-replace`. See `spec/09-app-issues/07-updater-deploypath-vs-path-mismatch.md`.

## v2.121.0

### Fixed
- **Updater finally targets the active PATH binary in update mode (root-cause fix for the v2.97.0 → v2.119.0 stuck-binary loop).** When `-Update` is set and `-TargetBinaryPath` is missing (legacy worker handoffs from older versions), `run.ps1` now resolves the active `movie` from PATH and uses *that* as the deploy target — bypassing whatever `powershell.json`'s `deployPath` is set to. This breaks the failure mode where `deployPath = D:\bin-run` but PATH points at `E:\bin-run\movie.exe`, which had been freezing the active binary on an old version forever and causing every subsequent update to re-spawn an old worker.
- **Update mode now unconditionally skips the post-deploy PATH-sync retry loop.** The "Active PATH binary is in use; retrying (1/5)..." spam, the "Could not sync active PATH binary after retries" warning, and the manual `Copy-Item` hint are now gone from `movie update` — that loop only ever produced false alarms because the deploy already replaced the right file.
- **Cleanup permission errors no longer surface as PowerShell `NativeCommandError` red blocks.** `updater/cleanup.go` now silently skips any locked `*-update-*` worker leftover (it is always swept on a later run) and routes the remaining best-effort warnings to stdout instead of stderr, so the worker's final lines stay clean even when the cleanup pass races a still-running sibling worker.

## v2.119.0

### Fixed
- **Windows updater handoff now stops re-touching the active PATH binary during update mode** — `run.ps1` now treats `-TargetBinaryPath` as the single source of truth and skips the old post-deploy PATH sync/copy loop when running under `-Update`, so the worker no longer tries to overwrite a still-running `movie.exe` in a second location.
- **Worker post-update checks now stay pinned to the handed-off target binary** — the temp script no longer falls back to `Get-Command movie`, which prevented stale PATH resolution from reintroducing wrong-binary version checks or cleanup calls.
- **Cleanup error output no longer leaks PowerShell native-command gibberish** — the worker now fully suppresses cleanup command stderr/stdout during the best-effort sweep, leaving only the updater's own ASCII-safe messages.

## v2.118.0

### Fixed
- **Updater handoff target-path regression** — the worker now passes the original executable path to `run.ps1` as one authoritative `-TargetBinaryPath` value instead of reconstructing the destination from split pieces. This keeps update-mode deploys pinned to the exact binary the user launched and prevents fallback to the wrong config or PATH-resolved location.
- **Worker cleanup preservation** — `update-cleanup --skip-path` now trims quotes and whitespace before path comparison, so the live handoff worker is reliably preserved and is no longer treated as a removable artifact.
- **Garbled Windows updater output** — the temp PowerShell update script now sets UTF-8 console encodings and updater status/error prefixes were normalized to ASCII-safe labels (`[OK]`, `[WARN]`, `[ERR]`) so warning lines no longer degrade into gibberish in PowerShell.

## v2.117.1

### Fixed
- **`updater/repo.go`: `normalizeRepoPath` quote stripping** — `TestNormalizeRepoPathTrimsQuotes` failed because `strings.Trim(raw, "\"'")` was applied **before** `TrimSpace`. Input `  "/tmp/x"  ` had the outer spaces protecting the quotes from being stripped, so the result still contained `"`. Reordered to: TrimSpace → Trim quotes → TrimSpace again. Quotes are now stripped from both bare and space-padded paste forms.

## v2.117.0

### Added
- **bootstrap.ps1 / bootstrap.sh** — version-discovery installers per `spec/03-general/05-install-latest-sibling-repo.md`. Probes sibling repos (`-v<N+k>` for `k = 25..0`, highest first) on the same GitHub owner and delegates install to the highest existing one. Auto-upgrades stale install URLs (e.g. user pastes `…/movie-cli-v8` but v5 exists → bootstrap picks v5).
  - Algorithm: parse URL → probe `raw.githubusercontent.com/<owner>/<base>-v<N+k>/main/install.{ps1,sh}` with 5s timeout, no retries → first HTTP 200 wins → `irm | iex` (PS) or `curl | bash` (Bash) the winner.
  - Edge cases: trailing `.git` stripped; URLs without `-v<N>` suffix install as-is; all-misses fallback to starting URL; `/tree/` and `/blob/` URLs rejected.
  - Logs every probe with verdict (`miss (404)` / `miss (timeout)` / `HIT`), final selection, and any auto-upgrade jump. Persistent log at `$env:TEMP/movie-bootstrap.log` (Win) or `/tmp/movie-bootstrap.log` (Unix).
  - End-to-end tested against `alimtvnetwork/movie-cli-v8` — probes v30→v5, hits v5, delegates to its install.sh.
  - New public install one-liners (paste forever, auto-upgrade silently):
    - PS: `irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/bootstrap.ps1 | iex`
    - Bash: `curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/bootstrap.sh | bash`

## v2.116.0

### Added
- **CI guard for acronym MixedCaps convention.** New step in `.github/workflows/ci.yml` (lint job) greps every `*.go` file for the forbidden pattern `\b(IMDb|TMDb|API|HTTP|URL|JSON|SQL|HTML|XML)[A-Z]`. Trailing-initialism short locals (`imdbID`, `tmdbID`, `*URL`) are allowlisted per spec section 2.1. On match the step fails the build with a descriptive `::error` pointing at the spec.

### Changed
- **JSON acronyms normalized** in 7 files: `JSONItems` → `JsonItems`, `JSONSubDir` → `JsonSubDir`, `printScanJSON` → `printScanJson`, `UseJSON`/`useJSON` → `UseJson`/`useJson`, `scanJSONItem` → `scanJsonItem`, `buildMediaJSONItem` → `buildMediaJsonItem`. Required to satisfy the new CI guard.

### Docs
- **spec/12-ci-cd-pipeline/01-ci-pipeline.md** — documented the acronym guard step alongside `golangci-lint`.

## v2.115.0

### Changed (BREAKING — internal API rename)
- **Acronym identifiers normalized to MixedCaps across all Go code.** Per the project's "Imdb not IMDb" convention (already applied to `attachImdbCacheUnless` in v2.107.0), every remaining acronym-cased identifier was renamed:
  - `tmdb.Client.HTTPClient` → `HttpClient`
  - `tmdb.Client.IMDbCache` → `ImdbCache`
  - `tmdb.Client.APIKey` → `ApiKey`
  - `tmdb.IMDbCache` (interface) → `tmdb.ImdbCache`
  - `tmdb.Client.SetIMDbCache` → `SetImdbCache`
  - `tmdb.Client.LookupByIMDbID` → `LookupByImdbId`
  - `tmdb.lookupByIMDbID` / `fetchIMDbIDFromDuckDuckGo` / `lookupIMDbCache` / `storeIMDbCache` / `tryIMDbViaWeb` / `imdbIDPattern` → `Imdb`/`Id` casing.
  - `cmd.tmdbCredentials.APIKey` → `ApiKey`
  - `cmd.initTMDbClient` / `resolveInfoTMDbClient` / `resolveSearchTMDbClient` / `resolveScanTMDbCredentials` / `readTMDbCredentials` / `readTMDbConfigValue` → `Tmdb` casing.
  - `cmd.newIMDbCacheAdapter` → `newImdbCacheAdapter`.
  - `db.TestInsertMediaAllowsMissingTMDbID` → `TestInsertMediaAllowsMissingTmdbId`.
- 17 files touched in a single sed sweep. No behaviour changes — pure rename. Local variables that follow Go's *trailing initialism* convention (`imdbID`, `tmdbID`, `imgURL`, `reqURL`) and database column names (`ImdbId`, `TmdbId`, already correctly cased) were intentionally left alone.

## v2.114.0

### Fixed
- **tmdb/client.go fieldalignment (real fix)** — v2.113.0 moved `HTTPClient` to the end, but govet still reported 56→48 because every field contained pointers, so the GC pointer-scan prefix still ran to offset 48. Reordered to `HTTPClient` (8 ptr bytes) → `IMDbCache` (16 ptr bytes) → `APIKey` (string) → `AccessToken` (string). The trailing string `len` words are non-pointer, so the pointer scan now stops at offset 40, giving the required 48 pointer bytes.

### Added
- **spec/03-general/05-install-latest-sibling-repo.md** — generic "find latest sibling repo + delegate to its installer" bootstrap spec. Parses `…-v<N>` URLs, probes `v(N+25)…v(N+0)` highest-first via raw.githubusercontent.com `install.ps1` (5s timeout, fail-fast, no retries), delegates to the winner via `irm | iex`, falls back to the starting URL if all candidates miss. Includes full PowerShell sketch, Bash notes, logging format, and acceptance criteria.

## v2.113.0

### Fixed
- **tmdb/client.go struct alignment** — reordered `Client` struct fields (`HTTPClient` moved after `IMDbCache`) to eliminate 8 bytes of padding, reducing size from 56 to 48 bytes (govet fieldalignment).

## v2.112.0

### Added
- **`movie cache imdb forget <cleanTitle> [year]` subcommand** — deletes a single row from the `ImdbLookupCache` so the next scan re-resolves that one title from scratch (DuckDuckGo + TMDb `/find`) without nuking the entire cache or having to pass `--no-cache` for every other title in the run. Year defaults to `0`. Prints a friendly warning when no matching row is found and points the user at `movie cache imdb list` for exact spellings.
- **`db.ForgetImdbLookup(cleanTitle, year)`** — new admin helper on the `DB` type that runs `DELETE FROM ImdbLookupCache WHERE LookupKey = ?` using the same `imdbLookupKey` normalization as the read/write paths.

## v2.111.0

### Fixed
- **CI lint errors** — 14 golangci-lint findings resolved across the codebase:
  - `updater/gitmap.go`: exported `gitmapDir` → `GitmapDir` and `readGitMapLatest` → `ReadGitMapLatest` (unused symbols)
  - `updater/repo_config.go`: changed `err == sql.ErrNoRows` to `errors.Is(err, sql.ErrNoRows)` (errorlint)
  - `updater/run.go`: renamed inner `err` → `saveErr` to avoid shadowing (govet)
  - `tmdb/client.go`: reordered struct fields for optimal alignment (fieldalignment)
  - `db/imdb_lookup_cache_admin.go`: reordered `ImdbCacheEntry` fields (fieldalignment)
  - `cmd/log_init_helper.go`: fixed British spelling `initialises` → `initializes` (misspell)
  - `db/migrate_v3.go`: fixed British spelling `behaviour` → `behavior` (misspell)
  - `cleaner/parse.go`: removed unsupported backreference `\1` from `duplicateYear` regex (staticcheck)
  - `version/info.go`, `updater/cleanup.go`, `updater/repo_config.go`, `updater/repo_test.go`: fixed gofmt alignment and trailing newlines

## v2.110.0

### Added
- **`--keep-logs` flag on `movie scan`, `movie rescan`, and `movie rescan-failed`** — opts out of the v2.109.0 logs-folder wipe so the previous run's `error.txt` is preserved. Useful for diffing two consecutive runs or keeping a longer error history. Prints `📝 --keep-logs: appending to existing <command> log` when active.
- **`cmd/log_init_helper.go`** — new shared `initRunLogger(outputDir, command, keepLogs)` helper that picks `errlog.Init` (append) vs `errlog.InitFresh` (wipe). Replaces the duplicated init-and-warn block in all three commands.

## v2.109.0

### Changed
- **Logs folder is wiped at the start of every `movie scan`, `movie rescan`, and `movie rescan-failed` run.** Each run now starts with a clean `.movie-output/logs/` directory instead of appending to the previous run's `error.txt`, so the file always reflects only the current run.
- **Renamed `attachIMDbCacheUnless` → `attachImdbCacheUnless`** to follow Go MixedCaps (Imdb, not IMDb) and stay consistent with the rest of the codebase.

### Added
- **`errlog.InitFresh(outputDir, command)`** — sibling of `errlog.Init` that `os.RemoveAll`s the logs dir before recreating it. Used by the three long-running commands above; one-shot commands (`rest`, `logs`, etc.) still use `Init` so their history is preserved.

## v2.107.0

### Added
- **`--no-cache` flag on `movie rescan` and `movie rescan-failed`** — detaches the persistent `ImdbLookupCache` for that single run so the search-fallback chain re-hits DuckDuckGo and TMDb `/find` from scratch. Lets users force-refresh a stale or wrong cached resolution without clearing the whole cache (and losing every other warm hit).
- **`cmd/imdb_cache_attach.go`** — new shared `attachImdbCacheUnless(client, db, noCache, commandName)` helper that either wires the cache or prints a one-line `--no-cache: bypassing IMDb cache for this <command> run` notice. Used by both rescan commands so the bypass behaviour stays in one place.

## v2.106.0

### Added
- **`movie cache imdb backfill` subcommand** — walks every cached HIT row whose `TmdbId` is 0 (legacy v2 rows or partial hits where `/find` never resolved) and re-runs TMDb `/find?external_source=imdb_id` for each one. Successful resolutions are written back into the cache so the next normal scan becomes a fully warm hit and skips both DuckDuckGo AND `/find`.
  - `--limit N` caps how many rows are processed in a single run.
  - `--dry-run` prints what would be resolved without writing to the cache.
  - A 250 ms pause between requests keeps the run under TMDb's rate limit.
- **`db.ListImdbLookupsUnresolved`** — returns every cached HIT with `TmdbId = 0 AND ImdbId != ''`, ordered oldest first so the longest-stale rows are backfilled first.
- **`tmdb.Client.LookupByIMDbID`** — exported wrapper around the existing private `lookupByIMDbID` so the cmd layer can call `/find` directly without touching the fallback chain.

## v2.105.0

### Added
- **TMDb `/find` is now cached too.** `ImdbLookupCache` v3 migration adds `TmdbId INTEGER` and `MediaType TEXT` columns. On a fully warm cache hit `tmdb.SearchWithFallback` returns a synthetic `SearchResult{ID, MediaType}` immediately, skipping both the DuckDuckGo HTML scrape AND the TMDb `/find?external_source=imdb_id` round-trip. Partial hits (IMDb id known but TmdbId never resolved) still call `/find` and back-fill the cache so the next run is fully warm.
- **`db.ImdbLookupResult` extended** with `TmdbID` and `MediaType`. `db.SetImdbLookup` signature widened to `(cleanTitle, year, imdbID, tmdbID, mediaType)`.
- **`movie cache imdb list`** now prints the resolved `TMDb: <id> (movie|tv)` line per entry, or `(unresolved)` for legacy/partial rows.

### Changed
- **`tmdb.IMDbCache` interface widened** to surface the cached TmdbId + MediaType (`Look` now returns 5 values; `Store` takes 5 args). The `cmd.imdbCacheAdapter` was updated accordingly. There are no other implementations in-repo.
- **`applyMovieDetails` / `applyTVDetails` now also populate `Description`, `TmdbRating`, and `Popularity`** from the `/movie/{id}` and `/tv/{id}` payloads. Previously these came only from the search result; with `/find` skipped on a warm hit they would have been lost otherwise.

### Migrations
- **v3** `ImdbLookupCache: add TmdbId + MediaType to skip /find on hit` — `ALTER TABLE ImdbLookupCache ADD COLUMN TmdbId INTEGER NOT NULL DEFAULT 0` and `... ADD COLUMN MediaType TEXT NOT NULL DEFAULT ''`. Existing rows keep working under the v2 behaviour (one `/find` call per hit) until they are re-resolved.

## v2.104.0

### Added
- **`movie cache imdb` command** with subcommands:
  - `movie cache imdb` — prints a summary (total entries, hits, misses).
  - `movie cache imdb list [--limit N]` — lists cached lookups most recent first; use `--limit 0` for all rows.
  - `movie cache imdb clear` — deletes every row from `ImdbLookupCache` (hits + misses).
  - `movie cache imdb clear-misses` — deletes only cached misses so previously-failed titles are retried on the next scan, while preserving long-lived hits.
- **`db.ListImdbLookups`, `db.CountImdbLookups`, `db.ClearImdbLookups`, `db.ClearImdbLookupMisses`** — new helpers in `db/imdb_lookup_cache_admin.go` backing the command.

## v2.103.0

### Added
- **`ImdbLookupCache` table (migration v2)** — persists every DuckDuckGo→IMDb id lookup keyed by lowercase clean title + year. Hits are valid for 180 days; misses for 7 days (so titles eventually retry as IMDb/TMDb improve).
- **`db.GetImdbLookup` / `db.SetImdbLookup`** — TTL-aware read/write helpers returning hit / miss / expired so callers can short-circuit the web call.
- **`tmdb.IMDbCache` interface + `Client.SetIMDbCache`** — optional persistent cache plug-in for the search fallback chain. When attached, `tmdb.Client.findIMDbIDViaWeb` consults the cache first and writes the result back (hit OR miss) so repeated `movie scan`, `movie rescan`, and `movie rescan-failed` runs no longer re-hit DuckDuckGo for the same `(clean title, year)` pair.
- **`cmd/imdb_cache_adapter.go`** — adapts `*db.DB` to `tmdb.IMDbCache`, swallowing DB errors so a broken cache degrades to a fresh web call instead of failing the search.

### Changed
- **All command sites that build a TMDb client now attach the cache**: `movie scan`, `movie scan --watch`, `movie rescan`, `movie rescan-failed`, and the REST `similar` handler. Commands that have no DB context (one-off `movie info`, `movie search`, etc.) continue to work without a cache.

## v2.102.0

### Fixed
- **`movie update` handoff is now correctly DETACHED again.** The previous "fix" (iteration 3) made the parent block on the worker with `cmd.Run()` so the console stayed attached, but that kept the parent process alive — which kept the OS file lock on the active `movie.exe`. The visible symptoms were `run.ps1` printing *"Active PATH binary is in use; retrying (1..5/5)"* and the cleanup step dying with *"Could not remove movie-update-<pid>.exe: Access is denied"* (because that file was the still-running worker). The parent now spawns the worker in its own console window (`CREATE_NEW_CONSOLE` on Windows, `setsid` on Unix) and exits 0 immediately, so the lock is released before `run.ps1` deploys.
- **Worker copy now self-deletes after the update finishes.** The temp PowerShell script ends with a hidden `cmd /c ping 127.0.0.1 -n 3 & del "<worker>"` so the `movie-update-<pid>.exe` copy is removed ~2 s after the worker exits. `update-cleanup` is still kept as a belt-and-braces sweeper for crashed/Ctrl-C runs and continues to honour `--skip-path`.

### Changed
- **Update worker output is indented and reformatted.** Every line in the worker's temp script now uses a `Say`/`SayOk`/`SayWarn`/`SayErr` helper with a consistent 4-space prefix, the closing pipe of the completion banner is aligned, and the banner uses `+======+` corners that line up. Easier to scan, no more ragged right edge.

### Docs
- **`spec/13-self-update-app-update/03-copy-and-handoff.md` rewritten** with the authoritative detached flow, a regression note explaining what went wrong in iteration 3, and an explicit "do not reintroduce `cmd.Run()`" rule.
- **`HANDOFF-LESSONS.md` added at the repo root** — a short, AI-shareable doc explaining why the obvious blocking-handoff "fix" is wrong and what to do instead. Hand this file to any AI/contributor before they touch the updater.
- **`.lovable/memory/issues/05-updater-async-console.md` updated** with the full 4-iteration history; the iteration-3 entry is now flagged as the cause of the bug, not the fix.

## v2.101.0

### Added
- **`movie rescan-failed` command** — selects every Media row where `TmdbId` is `NULL` or `0` and re-runs the lookup using the new `SearchWithFallback` chain (progressive trim → DuckDuckGo → IMDb id → TMDb `/find`). Supports `--limit N` to cap the batch size. After a successful pass it regenerates `report.html` and `summary.json` for affected scan dirs.
- **`db.GetMediaWithMissingTmdbID`** query helper used by the new command.

### Changed
- **`rescanMediaEntry` (used by both `movie rescan` and `movie rescan-failed`) now calls `SearchWithFallback`** instead of the bare `SearchMulti`, so the existing `movie rescan` also benefits from the trim + IMDb fallback for entries with weak metadata.

## v2.100.0

### Fixed
- **TMDb search no longer fails on filenames padded with release tags** like `Aan Paavam Pollathathu 10bit DS4K JHS ESub - Immortal 2025`. The cleaner now strips a much larger codec/release-group/streaming-platform vocabulary (DS4K, JHS, HEVC, AVC, EAC3, DDP5.1, AMZN, DSNP, etc.), cuts at `" - ReleaseGroup"` separators when the right side is pure junk, collapses duplicate years (`2025 2025` → `2025`), and runs junk removal in two passes so tokens revealed after the first pass also get stripped.

### Added
- **`tmdb.Client.SearchWithFallback`** — three-tier search chain used by `movie scan`:
  1. `SearchMulti(title + year)` (existing behavior)
  2. **Progressive trim**: drop the trailing word of the title repeatedly and retry, with and without the year, until a match is found or the title is too short.
  3. **IMDb-via-web fallback**: when TMDb still returns nothing, query DuckDuckGo HTML for `<title> <year> imdb`, extract the first `tt\d{7,10}` IMDb id, and resolve it via TMDb `/find/{imdb_id}?external_source=imdb_id`.
- `enrichFromTMDb` now reports `(year %d) after fallback chain` in its warning so it is obvious when even the IMDb fallback failed.

## v2.99.0


### Fixed
- **The update worker now calls `run.ps1` with explicit named arguments** instead of relying on a splatted hashtable, so `-DeployPath` and `-BinaryNameOverride` always bind and the update redeploys to the exact original binary path that launched `movie update`.
- **Cleanup skip-path comparison now normalizes both sides before matching**, so `update-cleanup` reliably preserves the still-running `movie-update-*.exe` handoff worker and no longer ends with `Access is denied` after a successful update.

## v2.98.0

### Fixed
- **`movie update` now passes `-Update`, `-DeployPath`, and `-BinaryNameOverride` into `run.ps1` via a named-parameter splat** instead of a raw string array. The handoff worker was launching correctly, but PowerShell was not binding those override flags, so `run.ps1` silently fell back to `powershell.json` and deployed to the wrong directory.
- **Auto-cleanup now skips the currently running handoff worker binary** via `update-cleanup --skip-path <worker>`, so post-update cleanup no longer tries to delete the still-running `movie-update-*.exe` copy and throw `Access is denied` at the end of a successful update.

## v2.97.0

### Fixed
- **`run.ps1` parser failure on Windows PowerShell 5.1** — the script contained UTF-8 em-dashes inside double-quoted strings. Without a UTF-8 BOM, Windows PowerShell Desktop reads `.ps1` as Windows-1252, mis-decodes the bytes, and reports cascading errors like `Missing closing '}'` and `The Try statement is missing its Catch or Finally block.` All em-/en-dashes in `run.ps1`, `build.ps1`, and `install.ps1` are now ASCII `--`, and each script starts with a UTF-8 BOM.
- **Removed misleading `📋 Gitmap` line from `movie update` preflight** in `updater/run.go`. The value referred to the previous release branch and was being printed before the new pull, looking like stale active state.

## v2.96.0

### Fixed
- **`movie update` now persists the resolved repo path into the local DB** using `RepoPath`, so future update runs can reuse the actual machine repo location instead of relying only on binary-dir heuristics.
- Added `movie update --repo-path <path>` as an explicit override, with trimming, quote cleanup, `~` expansion, absolute-path resolution, and validation against `.git` plus `go.mod` module `github.com/alimtvnetwork/movie-cli-v8`.
- Removed leftover updater references to `movie-cli-v8` during sibling/bootstrap repo resolution; bootstrap clone now targets `movie-cli-v8` and the canonical v5 GitHub URL.
- **Update handoff now runs only `run.ps1 -Update`** for git pull, build, and deploy; the generated worker script no longer performs its own `git pull`, keeping all mutable git/build logic inside `run.ps1` as intended.

## v2.95.0

### Fixed
- **install.sh now exports PATH for the running script and prints a copy-paste refresh hint** — when invoked via `curl … | bash`, the script runs in a subshell and cannot mutate the parent interactive shell's environment (Unix process isolation). Previously the rc-file write happened silently and users were left wondering why `movie` was "not found" until they opened a new terminal.
- The installer now ends with a clear one-liner (e.g. `export PATH="$HOME/.local/bin:$PATH"` or `fish_add_path …`) the user can paste to refresh their current shell immediately.
- Messaging updated: "Added to ~/.zshrc (new shells will pick it up)" instead of just "Added to ~/.zshrc", removing ambiguity about why the current shell still doesn't see the binary.

## v2.94.0

### Fixed
- **Installer now updates PATH for the current PowerShell session** — previously `[Environment]::SetEnvironmentVariable("PATH", ..., "User")` only affected *future* sessions, so users got "The term 'movie' is not recognized" immediately after install and had to open a new terminal. The installer now also refreshes `$env:PATH` in the running session so `movie` works right away.

## v2.93.0

### Changed
- **Module path renamed** — `github.com/alimtvnetwork/movie-cli-v8` → `github.com/alimtvnetwork/movie-cli-v8` across the entire project (104 files: Go imports, `go.mod`, README, CI workflows, install scripts, docs).
- All GitHub URLs (`github.com/alimtvnetwork/movie-cli-v8` → `…/movie-cli-v8`) updated in README badges, install one-liners, and release-asset URLs.

### Migration
- Run `go mod tidy` after pulling to refresh the module cache.
- Local clones tracking the old remote should update their `origin` URL: `git remote set-url origin https://github.com/alimtvnetwork/movie-cli-v8.git`.

## v2.92.0

### Changed
- **Updater scope locked down** — Go `updater/` package no longer runs `git checkout`/`pull`/`fetch` or build commands. All git mutations and build/deploy steps are delegated to `run.ps1 -Update`.
- Replaced `prepareRepoBranch` with `preflightRepo` in `updater/run.go` — only validates clean working tree before handoff.
- Eliminates "pathspec did not match" errors caused by attempting to check out gitmap release labels (e.g. `release/v1.31.0`) that aren't real branches.

### Added
- New memory constraint `mem://constraints/updater-scope.md` — documents the Go-vs-PowerShell scope split so future contributors don't reintroduce git logic in Go.

## v2.91.0

### Fixed
- Removed branch-switching logic from Go updater that caused `pathspec 'release/v1.31.0' did not match` failures on `movie update`.

## v2.16.0

### Changed
- **Extracted helpers from 12 oversized functions** — all command functions now comply with the ≤50-line guideline
  - `runMovieLsTable` (81→38) — extracted `printLsTableHeader`, `printLsTableRow`, `printLsTableDivider`, `formatRating`
  - `printMediaDetailTable` (75→30) — extracted `buildDetailTableRows` with declarative optional field list
  - `printMediaDetail` (69→33) — extracted `printDetailHeader`, `printDetailIdentifiers`, `printDetailRatings`, `printDetailCredits`, `printDetailFinancials`, `printDetailDescription`, `printDetailFiles`
  - `runMovieScan` (87→37) — extracted `createScanContext`, `executeScan`, `finalizeScan`
  - `runMovieRescan` (71→38) — extracted `fetchRescanEntries`, `processRescanEntries`, `printRescanResult`
  - `runMovieCd` (67→25) — extracted `listScanFolders`, `matchScanFolder`
  - `runWatchLoop` (64→30) — extracted `seedWatchSeen`, `processWatchCycle`, `logWatchScanHistory`
  - `writeHTMLReport` (64→33) — extracted `buildHTMLReportItems`, `splitGenreList`
  - `writeScanSummary` (62→27) — extracted `buildSummaryItems`, `categorizeByGenre`
  - `promptDestination` (62→38) — extracted `loadDestinationDirs`, `loadConfigDir`
  - `runMovieDuplicates` (60→24) — extracted `findDuplicateGroups`, `printDuplicateGroups`, `resolveDuplicatePath`
  - `runMoviePopout` (64) — already well-structured, no change needed

## v2.15.0

### Fixed
- **Update handoff now blocks (foreground)** — changed `cmd.Start()` + `Process.Release()` to `cmd.Run()` so the terminal stays stable and the user sees all worker output
- **Reads `.gitmap/release/latest.json`** — the update command now reads gitmap to determine the correct release branch, and checks out that branch before pulling
- New `updater/gitmap.go` with `GitMapRelease` struct and `readGitMapLatest()` reader

## v2.13.0

### Changed
- **Added 11 option structs for 39 functions with >3 params** — all functions now comply with the ≤3 parameter guideline
  - `StatsCounts` — groups totalMovies/totalTV/total for stats rendering
  - `MoveContext` — groups database/scanner/sourceDir/files/home for move flows
  - `CleanupContext` — groups scanner/database/batchID for popout cleanup
  - `ScanServiceConfig` — groups scanDir/outputDir/database/creds for post-scan services
  - `ScanLoopConfig` — groups client/scanDir/batchID/format flags for main scan loop
  - `ScanOutputOpts` — groups useTable/useJSON flags for scan output
  - `SuggestCollector` — groups client/existingIDs/count for suggestion helpers
  - `LsPage` — groups offset/pageSize/total for list pagination
  - `RecursiveWalkOpts` — groups baseParts/maxDepth for recursive directory walks
  - `ThumbnailInput` — groups client/database/media/posterPath/outputDir for thumbnail downloads
  - `HistoryLogInput` — groups basePath/title/year/fromPath/toPath for history logging
- New `cmd/types.go` with all option struct definitions
- Updated `Test-SourceFiles` in `run.ps1` (88 total)

## v2.12.0

### Changed
- **Split 3 oversized files to under 300 lines** — zero files now exceed the 300-line limit
  - `tmdb/client.go` (323→212) + new `tmdb/http.go` (125) — HTTP retry/response logic extracted
  - `db/schema.go` (314→48) + new `db/schema_tables.go` (263) + `db/schema_indexes.go` (40) — DDL split by table group
  - `cmd/movie_suggest.go` (314→154) + new `cmd/movie_suggest_helpers.go` (169) — genre analysis, discovery, and printing extracted
- Updated `Test-SourceFiles` in `run.ps1` with 4 new files (87 total)

## v2.11.0

### Fixed
- **Self-update now targets the exact original executable path** — the worker passes both deploy directory and binary filename into `run.ps1`, so rebuild/redeploy lands on the same binary that launched `movie update`
- **`run.ps1` deploy override completed** — added `-BinaryNameOverride` alongside `-DeployPath` so update mode no longer depends on `powershell.json` filename defaults

## v2.10.0

### Fixed
- **Self-update now redeploys the exact binary that launched `movie update`** — the original executable path is passed into the handoff worker and forwarded into `run.ps1` as a deploy-path override
- **True handoff flow** — the parent process now starts the copied worker and exits immediately so the original binary can release its file lock before rebuild/deploy
- **Repo-driven rebuild path** — the worker still runs `run.ps1` from the cloned/local GitHub repo, but now targets the original executable directory instead of only the default `powershell.json` deploy path
- **Hidden worker contract tightened** — `update-runner` now requires both `--repo-path` and `--target-binary`

## v2.9.0

### Changed
- **Eliminated all 41 `} else {` violations in Go code** — converted to early-return guard clauses, `continue` in loops, and extracted helpers across 30 files
- Remaining 2 `} else {` are in PowerShell template strings (`updater/script.go`) — not Go logic
- Extracted `runDryRunPlainOutput` and `incrementTypeCountPtr` helpers from oversized `runDryRunScan`
- Simplified `CountMedia` and `ListWatchlist` with early-return patterns
- Files changed: `movie_config.go`, `movie_discover.go`, `movie_history.go`, `movie_info.go`, `movie_logs.go`, `movie_move_batch.go`, `movie_play.go`, `movie_popout.go`, `movie_popout_cleanup.go`, `movie_popout_discover.go`, `movie_redo_exec.go`, `movie_redo_handlers.go`, `movie_rescan.go`, `movie_rest.go`, `movie_rest_report.go`, `movie_scan.go`, `movie_scan_json_output.go`, `movie_scan_loop.go`, `movie_scan_process.go`, `movie_scan_process_helpers.go`, `movie_scan_table.go`, `movie_search.go`, `movie_stats.go`, `movie_tmdb.go`, `movie_undo_handlers.go`, `update.go`, `db/cleanup.go`, `db/media_query.go`, `db/watchlist.go`

## v2.8.0

### Added
- **Pre-build source file validation** — `Test-SourceFiles` function in `run.ps1` validates 83 critical source files exist before compilation, catching missing files early (ported from gitmap-v2 pattern)

## v2.7.0

### Fixed
- **Updater: wrong GitHub repo URL** — `repoURL` used `movie-cli-v8.git` but actual GitHub repo is `movie-cli-v8`; sibling dir search also looked for wrong name
- **run.ps1: stale version file path** — referenced `version/version.go` (renamed to `version/info.go`), causing version detection to fail
- **run.ps1: wrong ldflags module path** — used `movie-cli-v8` instead of `movie-cli-v8` Go module path in build ldflags

### Added
- **run.ps1: `-Deploy` and `-Update` flags** — matches gitmap-v2 pattern; `-Deploy` forces deploy, `-Update` enables rename-first PATH sync
- **run.ps1: PATH binary sync** — when deployed binary differs from PATH binary, auto-syncs with retry and rename-first fallback (ported from gitmap-v2)
- **Updater: passes `-Update` flag to run.ps1** — enables PATH sync during `movie update` flow

## v2.6.0

### Changed
- **P4: Option structs for >3 params** — introduced 6 new input structs (`ErrorLogEntry`, `MoveInput`, `ScanHistoryInput`, `ActionInput`, `WatchlistInput`, `ScanStats`) to replace functions with 4–9 positional parameters; reduced violations from 58 → 47 across 18 files

## v2.5.0

### Changed
- **P3: Replaced all `fmt.Errorf` with `apperror.Wrap()`** — eliminated all 106 `fmt.Errorf` calls across the codebase; all errors now use `apperror.Wrap`, `Wrapf`, or `New` for consistent structured error handling

## v2.4.0

### Changed
- **P2: Eliminated nested ifs** — refactored top 10 worst files using early returns and guard clauses; flattened deeply nested conditionals across scan, move, rename, popout, suggest, rest, and undo commands

## v2.3.0

### Changed
- **Schema fix** — `db/schema.go` multi-value `d.Exec()` error fixed (single-value context)

## v2.2.0

### Changed
- **File splits** — extracted `movie_popout_discover.go`, `movie_popout_cleanup.go`, `movie_scan_loop.go` to keep files under 200 lines; removed duplicate function declarations

## v1.31.0

### Added
- **Version in CLI header box** — scan output now shows `🎬  Movie CLI v1.31.0` centered in the banner (matches gitmap style)

### Changed
- **Spec v1.1** (`spec/10-cli-output-spec.md`) — added flag reference table, JSON item schema, table column definitions, exit codes, flag interaction edge cases, metadata line priority order

## v1.30.0

### Added
- **`--rest` flag for `movie scan`** — starts REST server and opens HTML report in browser after scan completes
- **`--port` flag for `movie scan`** — customize REST server port when using `--rest`
- **REST API request logging** — every HTTP request logged via `errlog.Info` with method, path, status, duration
- **Thumbnails in output folder** — saved to `.movie-output/thumbnails/{slug}-{id}.jpg` with relative paths
- **Thumbnails served via REST** — `/thumbnails/` route serves poster images for the HTML report
- **Gitmap-style CLI output** — box header, numbered items with type icons (🎬/📺), ratings, tree-style output files
- **CLI output spec** — `spec/10-cli-output-spec.md` documents the full output format

### Changed
- Thumbnail naming: `{slug}-{tmdbID}.jpg` flat in `thumbnails/` dir (was nested subdirectories)
- Thumbnail path stored as relative (`thumbnails/xxx.jpg`) for portability
- REST HTML report uses `/thumbnails/` route for images instead of absolute file paths
- Scan output modernized: numbered items, category icons, structured sections

## v1.28.0

### Added
- **Centralized error logging system** (`errlog/logger.go`) — all errors are now logged to:
  - `.movie-output/logs/error.txt` (file-based, append-only, with timestamp/source/stack trace)
  - `error_logs` DB table (queryable, with level/source/function/command/workdir/stack trace)
- **`error_logs` table** (`db/errorlog.go`) — new table with columns: timestamp, level (ERROR/WARN/INFO), source, function, command, work_dir, message, stack_trace; includes `RecentErrorLogs()` query
- **`errlog` package** — `Error()`, `Warn()`, `Info()` functions with automatic caller detection, stack trace capture (errors only), and dual output (file + DB)
- **DB writer injection** — `errlog.SetDBWriter()` allows wiring DB logging without circular imports

### Changed
- **`movie scan` errors** — DB search, stat, insert, update, JSON write, TMDb, and thumbnail errors now use `errlog` instead of raw `fmt.Fprintf(os.Stderr)`
- **`movie rest` errors** — JSON encode, template render, watchlist update, tag add, config read errors now use `errlog`
- **Error entries include**: timestamp, severity, source file:line, function name, CLI command, working directory, message, and full Go stack trace

## v1.27.0

### Changed
- **Modernized HTML report** — complete UI overhaul: sticky toolbar with inline search, genre/rating/sort dropdowns, type filter pills, dark zinc theme, result count, empty state, keyboard shortcut (`/` to search, `Esc` to close modal), responsive layout
- **Search now searches titles, directors, and cast** — not just titles
- **Genre filter dropdown** — auto-populated from scan data
- **Rating filter dropdown** — filter by minimum rating (5+ through 9+)
- **Sort options** — sort by title, rating, or year (ascending/descending)
- **Connected REST indicator** — banner shows green dot when REST server is detected

### Fixed
- **`writeJSON` error swallowed** — `json.Encoder.Encode` error now logged to stderr
- **`tmpl.Execute` error swallowed** — template render error now logged to stderr
- **`GetConfig` errors swallowed** — `tmdb_api_key` and `tmdb_token` config read errors now logged
- **`database.Exec` watchlist update error swallowed** — now logged to stderr
- **`database.AddTag` watched tag error swallowed** — now logged to stderr
- **JS error handling** — all `catch(e)` blocks now show specific error messages; `fetch` non-ok responses show HTTP status/body

## v1.26.0

### Added
- **`GET /` on REST server** — serves a live HTML library report rendered from the database; always up-to-date, no need to open a static file

## v1.25.0

### Added
- **HTML report: tag management** — add/remove tags per card with inline input; tags shown as purple pills with ✕ to remove
- **HTML report: mark watched** — 👁 button marks a movie as watched via REST API; card gets green border and "watched" tag
- **HTML report: similar movies** — 🔍 button opens a modal with TMDb recommendations (poster, title, year, rating, description)
- **HTML report: watched filter** — new "✅ Watched" filter button in the toolbar
- **HTML report: tags auto-load** — when REST server is detected, all tags load automatically on page open

## v1.24.0

### Added
- **`GET/POST/DELETE /api/tags`** — full tag management via REST: list all tags with counts, list tags per media, add tag, remove tag
- **`GET /api/media/{id}/similar`** — fetches TMDb recommendations for a media item
- **`PATCH /api/media/{id}/watched`** — marks a media item as watched (updates watchlist + adds "watched" tag)
- **Refactored REST handlers** — new endpoints in `cmd/movie_rest_handlers.go` to keep files under 200 lines

## v1.23.0

### Added
- **`movie rest --open`** — automatically opens the HTML report in the default browser when the REST server starts; supports macOS (`open`), Windows (`rundll32`), and Linux (`xdg-open`)

## v1.22.0

### Added
- **`movie rest`** — starts a local REST API server (default port 8086, `--port` to override) exposing library endpoints: `GET /api/media`, `GET/DELETE/PATCH /api/media/{id}`, `GET /api/stats`; enables interactive features in the HTML report
- **HTML report** — `movie scan` now generates `report.html` in `.movie-output/` with responsive card layout showing thumbnail, title, year, rating, genre, director, cast, description, and tagline; includes search, filter, and delete via REST API
- **`templates/report.html`** — external HTML template file (not embedded in Go code); bundled via Go `embed` at compile time through `templates/embed.go`

## v1.21.0

### Added
- **`movie db`** — prints the resolved database path, data directory, and record counts for debugging

## v1.20.0

### Changed
- **Renamed `<package>/<package>.go` files** — `db/db.go` → `db/open.go`, `cleaner/cleaner.go` → `cleaner/parse.go`, `updater/updater.go` → `updater/run.go`, `version/version.go` → `version/info.go`; enforced as a permanent naming convention

## v1.19.0

### Added
- **`movie history --format table`** — output move history as a formatted table with columns: #, Title, From, To, Date, Status

## v1.18.0

### Added
- **Binary-relative data storage** — all data (database, thumbnails, JSON metadata) is now stored in `data/` next to the CLI binary, not the working directory
- **`run.ps1` deploys data folder** — build script copies data directory alongside the deployed binary

## v1.17.0

### Added
- **`movie ls --format table`** — output library listing as a formatted table with columns: #, Title, Year, Type, Rating, Genre, Director (no interactive pager)

### Changed
- **Refactored `movie_ls.go`** — split 313-line file into `movie_ls.go` (196), `movie_ls_table.go` (99), and `movie_ls_detail.go` (120)

## v1.16.0

### Changed
- **Refactored `movie_search.go`** — extracted save-and-print logic into `cmd/movie_search_save.go` (135 lines); `movie_search.go` reduced from 240 to 135 lines

## v1.15.0

### Added
- **`movie stats --format table`** — output library statistics as a formatted key-value table with sections for counts, storage, genres, and ratings

## v1.14.0

### Changed
- **Refactored `movie_info.go`** — extracted `fetchMovieDetails` and `fetchTVDetails` into `cmd/movie_fetch_details.go`

## v1.13.0

### Fixed
- **`movie update` fresh-clone flow** — when no local repo exists, a new clone is now reported as bootstrap success instead of incorrectly saying "Already up to date"
- **Self-update specs** — documented repo bootstrap vs existing-repo pull behavior using the GitMap-aligned update flow

## v1.12.0

### Added
- **`movie search --format table`** — output TMDb search results as a formatted table (no interactive prompt); columns: #, Title, Year, Type, Rating, TMDb ID
- **`movie info --format table`** — output media detail as a key-value formatted table; shows all metadata fields dynamically

## v1.11.0

### Added
- **`movie search --format json`** — output TMDb search results as a JSON array to stdout (no interactive prompt); pipeable to `jq` and scripts
- **`movie info --format json`** — output media detail as a JSON object to stdout; includes source field ("local" or "tmdb")

## v1.10.0

### Added
- **`movie ls --format json`** — output entire library as a JSON array to stdout; includes id, title, year, type, ratings, genre, file path, and file size per item
- **`movie stats --format json`** — output library statistics as a JSON object to stdout; includes counts, storage, top genres, and average ratings

## v1.9.0

### Added
- **`movie scan --format json`** — output scan results as structured JSON to stdout for piping to `jq`, scripts, or other tools; includes metadata, counts, and per-item details; works with `--dry-run` too

## v1.8.0

### Fixed
- **`movie scan` no longer fails when TMDb is unset** — media with no TMDb match/key now store `tmdb_id` as `NULL` instead of `0`, so bulk scans no longer hit `UNIQUE constraint failed: media.tmdb_id`
- **Interactive TMDb setup before scan** — when TMDb is not configured, `movie scan` now prompts for a TMDb API key and TMDb access token before scanning starts; leaving both blank continues without metadata
- **TMDb bearer token support** — scan can now authenticate with either `tmdb_api_key` or `tmdb_token`

## v1.7.1

### Changed
- **Refactored `movie_scan.go`** — split from ~500 lines into 4 focused files:
  - `movie_scan.go` (~120 lines) — command definition, orchestrator, helpers
  - `movie_scan_collect.go` (~110 lines) — video file discovery and path utilities
  - `movie_scan_process.go` (~170 lines) — per-file processing and TMDb enrichment
  - `movie_scan_table.go`, `movie_scan_json.go`, `movie_scan_summary.go` — unchanged

## v1.7.0

### Added
- **`movie scan --format table`** — display scan results as a formatted table with columns for #, filename, clean title, year, type, rating, and status; works with `--dry-run` too

## v1.6.0

### Added
- **`movie scan --dry-run`** — preview what would be scanned (files found, cleaned titles, types) without writing to DB or creating `.movie-output/`

## v1.5.0

### Added
- **`movie scan --depth N` (`-d`)** — limit recursive scan to N subdirectory levels (0 = unlimited); e.g. `movie scan -r -d 2`

## v1.4.0

### Fixed
- **`movie update` works from anywhere** — no longer requires CWD to be inside the git repo; finds the repo next to the binary, clones fresh if needed

## v1.3.0

### Added
- **`movie scan --recursive` (`-r`)** — scan all subdirectories recursively instead of just top-level entries; skips `.movie-output` and hidden directories automatically

### Changed
- **Refactored scan internals** — extracted `collectVideoFiles`, `processVideoFile`, and `enrichFromTMDb` helpers for cleaner architecture and reuse

## v1.2.0
### Changed
- **`movie scan` defaults to current directory** — running `movie scan` without arguments now scans the CWD instead of a config-stored `scan_dir` path
- **Scan output to `.movie-output/`** — all scan results (per-item JSON, summary.json with categories/descriptions/metadata) are now written to `.movie-output/` inside the scanned folder

### Added
- **`summary.json`** — comprehensive scan report with total counts, genre-based categories, and full TMDb metadata per item

## v1.1.0

### Fixed
- **`run.ps1` version stamping** — now reads the version from `version/version.go` and injects commit/build date into the correct `version` package variables
- **`run.ps1` version summary** — now reports the binary that was just built/deployed instead of accidentally showing an older `movie` found earlier in `PATH`
- **Deployed changelog visibility** — `run.ps1` now copies `CHANGELOG.md` beside the deployed binary and verifies `movie changelog --latest`

## v0.2.4

### Fixed
- **`GetConfig` false warnings** — `movie_info.go` and `movie_scan.go` now explicitly ignore `sql: no rows in result set` from `GetConfig`, preventing false-positive error messages when config keys are unset
- **Indentation fix** — corrected misleading indentation in `movie_scan.go` error block

### Changed
- **JSON export completeness** — `movie_export.go` now includes all 6 previously missing metadata fields: `Runtime`, `Language`, `Budget`, `Revenue`, `TrailerURL`, `Tagline`

## v0.2.3

### Fixed
- **`db/media.go` silent scan error** — `TopGenres` now returns a wrapped error on `rows.Scan` failure instead of silently using `continue`
- **`movie_info.go` poster error swallowed** — `DownloadPoster` failures now logged to stderr
- **`movie_scan.go` poster error swallowed** — `DownloadPoster` failures now logged to stderr
- **`movie_scan.go` subdirectory read error** — `os.ReadDir` failures in subdirectory scanning now logged instead of silently skipped
- **`movie_undo.go` permission error masked** — `os.Stat` now distinguishes permission errors from file-not-found and logs them separately

## v0.2.2

### Fixed
- **`movie_search.go` unchecked `GetConfig`** — API key lookup now checks for errors before proceeding
- **`movie_suggest.go` unchecked `GetConfig`** — API key lookup now checks for errors and handles `sql: no rows` correctly
- **`movie_resolve.go` unbounded query** — `resolveByTitle` now uses `LIMIT 1` to prevent scanning full table
- **`db/media.go` missing `rows.Err()` check** — `TopGenres` now checks `rows.Err()` after iteration loop

### Changed
- **`movie_search.go` duplicate detail fetch removed** — eliminated redundant `GetMovieDetails`/`GetTVDetails` calls that were already handled by shared `fetchMovieDetails`/`fetchTVDetails` helpers

## v0.2.1

### Fixed
- **`movie_move.go` unchecked error** — `database.GetConfig("movies_dir")` error now handled instead of silently ignored
- **`movie_move.go` unchecked error** — `database.GetConfig("tv_dir")` error now handled instead of silently ignored
- **`movie_move_helpers.go` cross-drive cleanup** — copy+delete fallback now removes the source file after successful copy
- **`movie_rename.go` unchecked `InsertMoveHistory`** — rename history logging error now reported to stderr
- **`movie_play.go` unchecked `exec.Command` error** — player launch error now reported to stderr
- **`movie_stats.go` unchecked `CountMedia`** — movie/TV count errors now handled instead of silently returning zero
- **`movie_watch.go` unchecked `GetConfig`** — API key lookup now checks for errors before proceeding
- **`tmdb/client.go` unchecked `json.NewDecoder` error** — HTTP response body decoding errors now properly returned
- **`updater/updater.go` unchecked exec errors** — `git pull` and `go build` errors now returned instead of silently ignored

## v1.0.0

### Added
- **Batch move** (`movie move --all`) — move all video files at once with auto-routing to movies/TV directories, preview table, and `[y/N]` confirmation
- **JSON metadata export** — `movie scan` now writes per-file JSON metadata to `./data/json/movie/` and `json/tv/`
- **Genre-based discovery** — `movie suggest` uses `DiscoverByGenre` for TMDb genre-based recommendations (3-phase: genre discovery → recommendations → trending fallback)
- **`GenreNameToID()` helper** — reverse genre map in tmdb package for name-to-ID lookups
- **CI pipeline** (`.github/workflows/ci.yml`) — lint (`go vet` + `golangci-lint`), vulnerability scanning (`govulncheck`), parallel test matrix, cross-compiled builds (6 targets), SHA deduplication
- **Release pipeline** (`.github/workflows/release.yml`) — triggers on `release/**` branches and `v*` tags, cross-compiled binaries, SHA256 checksums, version-pinned install scripts, changelog extraction
- **Cross-platform install scripts** — `install.sh` (Linux/macOS) and `install.ps1` (Windows) with checksum verification and PATH setup
- **`.golangci.yml`** — sensible linter defaults (errcheck, govet, staticcheck, gocritic, misspell, errorlint, etc.)
- **Undo confirmation prompt** — `movie undo` shows from/to paths and asks `[y/N]` before reverting
- **Tag command** (`movie tag`) — add, remove, and list tags on media entries
- **Comprehensive CLI help** — root command shows version + categorized help with examples; `movie --version` flag; `movie version` shows Go/OS/arch

### Changed
- **`movie ls`** now only shows scan-indexed items (filters by non-empty `original_file_path`)
- **`movie suggest`** upgraded from recommendations-only to 3-phase strategy (DiscoverByGenre → Recommendations → Trending)
- **Repository migrated** from `movie-cli-v8` through to `movie-cli-v8` across all imports, workflows, and docs

### Fixed
- Timestamp bug — `saveHistoryLog` now uses `time.Now().Format(time.RFC3339)` instead of hardcoded "now"
- Deduplicated TMDb fetch logic — shared `fetchMovieDetails()`/`fetchTVDetails()` helpers
- Cross-drive move fallback — copy+delete when `os.Rename` fails with `EXDEV`

## v0.1.0

### Added
- Core CLI with Cobra: `hello`, `version`, `self-update` commands
- Movie management: `scan`, `ls`, `search`, `info`, `suggest`, `move`, `rename`, `undo`, `play`, `stats`, `config`
- SQLite database with WAL mode, 5 tables, 7 indexes
- TMDb API client (search, details, credits, recommendations, trending, posters)
- Filename cleaner (junk removal, year extraction, TV detection)
- PowerShell build & deploy pipeline (`run.ps1`)
- Full project specification in `spec/`
