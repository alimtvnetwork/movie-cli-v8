# Contributing to Movie CLI

Thank you for your interest in contributing! This guide covers everything you need to get started.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Guidelines](#code-guidelines)
- [Project Naming Rules](#project-naming-rules)
- [Acronym MixedCaps Rules](#acronym-mixedcaps-rules)
- [Undo / Redo Worked Examples](#undo--redo-worked-examples)
- [Undo / Redo Edge Cases](#undo--redo-edge-cases)
- [Pre-push Checklist](#pre-push-checklist)
- [Commit Messages](#commit-messages)
- [Pull Requests](#pull-requests)
- [Issue Reporting](#issue-reporting)
- [Architecture Overview](#architecture-overview)
- [Testing](#testing)
- [Release Process](#release-process)

---

## Getting Started

### Prerequisites

| Tool | Minimum Version | Install |
|------|----------------|---------|
| **Go** | 1.22+ | [go.dev/dl](https://go.dev/dl/) |
| **Git** | 2.x | [git-scm.com](https://git-scm.com/) |
| **PowerShell** | 7+ (cross-platform) | `brew install --cask powershell` |

### Setup

```bash
# Fork the repo on GitHub, then:
git clone https://github.com/<your-username>/movie-cli-v8.git
cd movie-cli-v8
go mod tidy
make build
```

Verify:

```bash
./movie version
```

See the [Install Guide](spec/03-general/01-install-guide.md) for detailed setup instructions.

---

## Development Workflow

1. **Create a branch** from `main`:
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes** — keep each commit focused on a single concern.

3. **Run checks locally** before pushing:
   ```bash
   go vet ./...
   go test ./... -v -count=1
   ```

4. **Push and open a Pull Request** against `main`.

### Branch Naming

| Prefix | Purpose | Example |
|--------|---------|---------|
| `feature/` | New functionality | `feature/export-csv` |
| `fix/` | Bug fixes | `fix/scan-empty-dir` |
| `refactor/` | Code restructuring | `refactor/split-move-cmd` |
| `docs/` | Documentation only | `docs/update-readme` |
| `release/` | Release branches (maintainers) | `release/v1.4.0` |

---

## Code Guidelines

### General Rules

- **One file per command**, max ~200 lines
- Shared helpers go in `movie_info.go` and `movie_resolve.go`
- Use descriptive variable names — no abbreviations
- All exported functions must have doc comments

### Go Style

- Format with `gofmt` (enforced by CI)
- Pass `go vet ./...` and `golangci-lint` with zero warnings
- Error messages: lowercase, no trailing punctuation
  ```go
  // ✅ Good
  return fmt.Errorf("file not found: %s", path)

  // ❌ Bad
  return fmt.Errorf("File not found: %s.", path)
  ```

### Error Handling

- Always handle errors explicitly — never use `_` for error returns
- Wrap errors with context using `fmt.Errorf("operation: %w", err)`
- Use early returns to reduce nesting

### File & Folder Naming

- Spec files: `01-name-of-file.md` (numbered prefix)
- Go files: `snake_case.go`
- Root spec files: lowercase (`spec.md`, `ai-handoff.md`)

### Dependencies

- Prefer stdlib over third-party packages
- Any new dependency requires justification in the PR description
- Current deps: Cobra (CLI), modernc.org/sqlite (no CGo)

---

## Project Naming Rules

The user-facing binary, the Go module's display name, all CLI examples,
docs, env vars, paths, user-agents, and dotfolder conventions MUST use the
canonical name **`movie`**.

The previous codename (referred to throughout this guide as `<LEGACY>` so
that this document does not itself contain the banned token) is permanently
retired. CI fails the build on any case-insensitive match outside the
allowed historical zones (CHANGELOG and the checker's own source).

### Substitution Mapping

When migrating any file (including external snippets pasted in), apply
these case-preserving substitutions in order:

| Found            | Replace with | Use case                          |
|------------------|--------------|-----------------------------------|
| `<LEGACY>_<X>`   | `MOVIE_<X>`  | env var prefixes (`MOVIE_DB`, `MOVIE_HOME`) |
| `<LEGACY>`       | `MOVIE`      | uppercase usage in shell/Go consts |
| `<Legacy>`       | `Movie`      | TitleCase prose, struct field names |
| `<legacy>`       | `movie`      | binary name, paths, user-agents, CLI examples |

### Examples

| ❌ Banned                                | ✅ Correct                              |
|-----------------------------------------|-----------------------------------------|
| `` `<legacy>` v2.178.0 ``               | `` `movie` v2.178.0 ``                  |
| `go build -o /tmp/<legacy> .`           | `go build -o /tmp/movie .`              |
| `export <LEGACY>_DB=...`                | `export MOVIE_DB=...`                   |
| `<legacy>-cli/2.x` (User-Agent)         | `movie-cli/2.x`                         |
| `~/.<legacy>/config.json`               | `~/.movie/config.json` (or per spec, `.movie-output/`) |
| `<legacy> scan testdata/library`        | `movie scan testdata/library`           |

The fuzzy auto-fixer also normalizes whitespace/formatting variants such as
`m a h i n`, `m-a-h-i-n`, `m_a_h_i_n`, `m.a.h.i.n`, and zero-width-joined
forms — but you should never write them in the first place.

### Tooling

| Mode      | Command                                                          | What it does |
|-----------|------------------------------------------------------------------|--------------|
| Check     | `bash scripts/check-binary-name.sh`                              | Lists every offending occurrence with `file:line` and exits non-zero. Used by CI. |
| Dry-run   | `bash scripts/check-binary-name.sh --dry-run`                    | Previews per-file replacement counts without writing. |
| Fix       | `bash scripts/check-binary-name.sh --fix`                        | Applies the strict case-preserving substitutions in place. |
| Fuzzy fix | `bash scripts/check-binary-name.sh --fix --fuzzy`                | Also normalizes whitespace/formatting variants before the strict pass. |
| JSON      | `bash scripts/check-binary-name.sh --fix --json /tmp/sum.json`   | Writes a structured summary (`files_changed`, `total_replacements`, per-file `before`/`after`/`replaced` + `by_pattern` breakdown). |

The repo-wide CI guard (`scripts/guard-forbidden-terms.sh`) is invoked
automatically by `.github/workflows/ci.yml` → job **Lint** → step
**Forbidden term guard**.

### CI Mapping

| Local command                                  | CI job → step                                         | Log location                                        |
|-----------------------------------------------|-------------------------------------------------------|-----------------------------------------------------|
| `bash scripts/guard-forbidden-terms.sh`       | `ci.yml` → Lint → "Forbidden term guard"              | Actions → CI → Lint → "Forbidden term guard"        |
| `bash scripts/check-binary-name.sh`           | `ci.yml` → Lint → "Binary name consistency check"     | Actions → CI → Lint → "Binary name consistency check" |

---

## Acronym MixedCaps Rules

Spec: [`spec/01-coding-guidelines/03-coding-guidelines-spec/03-golang/09-acronym-naming.md`](spec/01-coding-guidelines/03-coding-guidelines-spec/03-golang/09-acronym-naming.md)

In Go **identifiers**, the following acronyms must be written in MixedCaps
form whenever they are followed by another uppercase letter (i.e. mid-word).
Comments and string literals are exempt — only identifier names are checked.

| Acronym | Use in identifiers | Examples (✅ / ❌) |
|---------|--------------------|--------------------|
| `IMDb`  | `Imdb`             | `ImdbId`, `fetchImdbRecord` / `IMDbID`, `fetchIMDbRecord` |
| `TMDb`  | `Tmdb`             | `TmdbClient` / `TMDbClient` |
| `API`   | `Api`              | `ApiKey`, `ApiBaseUrl` / `APIKey`, `APIBaseURL` |
| `HTTP`  | `Http`             | `HttpClient`, `HttpTimeout` / `HTTPClient` |
| `URL`   | `Url`              | `UrlPath`, `BaseUrl` / `URLPath`, `BaseURL` |
| `JSON`  | `Json`             | `JsonResponse`, `parseJsonBody` / `JSONResponse` |
| `SQL`   | `Sql`              | `SqlBuilder` / `SQLBuilder` |
| `HTML`  | `Html`             | `HtmlReport` / `HTMLReport` |
| `XML`   | `Xml`              | `XmlParser` / `XMLParser` |

**Trailing initialism is allowed** (the acronym is not followed by another
uppercase letter): `imdbID`, `tmdbID`, `baseURL`, `reqURL` are fine.

### Tools

- Check only:&nbsp;&nbsp; `python3 scripts/check-acronym-naming.py`
- Auto-rename:&nbsp; `python3 scripts/rename-acronyms.py --write`

The codemod skips comments and string/rune literals; review the diff before
committing.

---

## Undo / Redo Worked Examples

`movie undo` reverts the **last** move/rename/popout action recorded in
`ActionHistory`. `movie redo` re-applies the most recently undone action.
Both operate on a single action at a time — re-run them to walk further
back/forward through history.

### Example 1 — Undo a rename

```bash
# Starting state
$ movie ls --limit 1
ID  Title                              File
12  The.Matrix.1999.1080p.BluRay.mkv   /library/raw/The.Matrix.1999.1080p.BluRay.mkv

# Clean up the filename
$ movie rename 12
✅ renamed: The.Matrix.1999.1080p.BluRay.mkv → The Matrix (1999).mkv
   action_id=4711 recorded in ActionHistory

# Oops — wanted to keep the original name
$ movie undo
↩️  reverting action_id=4711 (rename)
✅ The Matrix (1999).mkv → The.Matrix.1999.1080p.BluRay.mkv
   IsReverted=true
```

**Outcome:** the file is back at its original path; the row in
`ActionHistory` is marked `IsReverted=true` but is **not** deleted, so
`movie redo` can re-apply it.

### Example 2 — Redo after an undo

```bash
$ movie redo
↪️  re-applying action_id=4711 (rename)
✅ The.Matrix.1999.1080p.BluRay.mkv → The Matrix (1999).mkv
   IsReverted=false
```

**Outcome:** the rename is reinstated and `IsReverted` flips back to
`false`. A fresh `movie undo` would now revert it again.

### Example 3 — Undo a move, then a rename (LIFO)

```bash
$ movie move 12 /library/movies      # action_id=4712
$ movie rename 12                    # action_id=4713

$ movie undo                         # reverts 4713 (the rename) first
$ movie undo                         # then reverts 4712 (the move)
```

**Outcome:** undo always targets the newest non-reverted action, so
two `movie undo` calls walk the history in LIFO order back to the
pre-move state. Two `movie redo` calls would replay them in the
original order (`4712` then `4713`).

> **Scope reminder:** undo/redo cover `move`, `rename`, and `popout`
> only. They do **not** touch TMDb metadata, tags, watchlist entries,
> or scan results — those have no `ActionHistory` row to revert.

---

## Undo / Redo Edge Cases

This section documents every "unhappy path" you may hit. Each case lists
the exact log line, exit code, and DB/filesystem state so you can
troubleshoot without reading source.

### 1. Nothing to undo / redo

When no matching history row exists in scope:

```bash
$ movie undo
📭 No operations to undo in this scope.

$ movie redo
📭 No reverted operations to redo in this scope.
```

- **Exit code:** `20` (`ExitNothingMatched`) — distinct from `0` so
  scripts can branch on "empty" vs "succeeded".
- **DB state:** unchanged. No row is read or written.
- **Filesystem:** untouched.
- **Common causes:** running from a directory with no history rows
  (default scope is cwd — pass `--global` to ignore it), every recent
  row is already reverted (try `movie undo --list`), or `--include` /
  `--exclude` globs filtered every candidate out (the summary line
  reports the skipped count).

### 2. Targeting an already-reverted (or already-applied) action

```bash
$ movie undo --id 4712
⚠️  Action 4712 has already been reverted.

$ movie redo --id 4712
⚠️  Action 4712 is not reverted — nothing to redo.
```

- **Exit code:** `20` (`ExitNothingMatched`).
- **DB state:** unchanged — `IsReverted` is not flipped twice.
- **Why:** undo only acts on `IsReverted=0` rows and redo only on
  `IsReverted=1` rows. This is what protects you from double-applying
  the same revert.

### 3. Missing source file (move undo / redo)

The recorded `ToPath` (undo) or `FromPath` (redo) no longer exists —
typically because the user moved or deleted it manually outside the CLI.

```bash
$ movie undo --move-id 42
[undo] kind=Move         move_id=42   before=/dst/x.mkv after=/src/x.mkv
❌ Undo move 42 failed: file not found at /dst/x.mkv — may have been moved manually
```

- **Exit code:** `2` (`ExitGenericError`).
- **DB state:** unchanged. `MoveHistory.IsReverted` is **not** flipped,
  so a subsequent `movie undo --move-id 42` will retry once you put
  the file back (or fix the path manually).
- **Filesystem:** no partial move — `MoveFile` is never called when the
  pre-flight `os.Stat` fails.
- **Recovery:** restore the file at the printed `before=` path, then
  re-run the same command; or use `movie undo --list` to pick a
  different target.

### 4. Destination conflict (file already at the target path)

If `before=` and `after=` both exist (e.g. you manually copied the file
back before running undo, or two history rows point at the same
destination), `MoveFile` will refuse to overwrite:

- **Exit code:** `2` (`ExitGenericError`).
- **DB state:** unchanged.
- **Filesystem:** both files remain in place — no data is lost.
- **Recovery:** delete or rename whichever copy is stale, then re-run.
  Inspect with `ls -la <before> <after>` first; the structured
  `[undo] … before= after=` log line tells you exactly which two paths
  to compare.

### 5. Missing snapshot (action undo / redo)

Action types that rebuild rows from a JSON snapshot (`Delete`,
`ScanRemove`, `RescanUpdate`, `Restore`) require `MediaSnapshot` to be
present. If the row is missing it (legacy data, manual edit):

```
❌ Undo action 99 failed: no snapshot available for action 99 — cannot restore
```

- **Exit code:** `2` (`ExitGenericError`).
- **DB state:** unchanged.
- **Recovery:** there is none — the snapshot is the only source of
  truth for what to restore. Skip the row with `movie undo --list` and
  pick the next undoable action.

### 6. Partial batch undo / redo

`movie undo --batch` / `movie redo --batch` walk every action in the
batch in LIFO / FIFO order. If some succeed and others fail:

```
⚠️  Batch a1b2c3d4: 7 reverted, 2 failed.
```

- **Exit code:**
  - `0` if at least one action succeeded (partial success is **not**
    an error so cron jobs can keep running).
  - `2` only when **every** action in the batch failed.
- **DB state:** each succeeded action has `IsReverted` flipped; failed
  ones stay as they were. The batch is therefore in a *mixed* state —
  re-running the same `--batch` command will retry only the still-
  unreverted (or still-reverted, for redo) rows.
- **Logging:** every failure is printed via `errlog.Warn` with the
  offending `action_id`, so the full failure list is grep-able from
  the run output.

### 7. Scope rejection at the confirmation prompt

When you run undo/redo from a non-`$HOME` cwd without `--global`, the
CLI shows a preview of how many rows are in scope and asks to confirm:

```
Undo will act on 0 moves, 3 actions under /current/dir. Continue? [y/N]:
```

- Answering `n` (or hitting Enter) exits with code `10`
  (`ExitScopeRejected`). Nothing is changed.
- Pass `--yes` (or `--assume-yes`) for scripted runs to skip the prompt.

### Quick reference — exit codes

| Code | Constant              | Meaning                                          |
| ---- | --------------------- | ------------------------------------------------ |
| 0    | `ExitOK`              | At least one row reverted/replayed successfully  |
| 2    | `ExitGenericError`    | Filesystem error, missing snapshot, full failure |
| 20   | `ExitNothingMatched`  | Empty scope, already-reverted, no candidates     |
| 10   | `ExitScopeRejected`   | User declined the cwd-scope confirmation prompt  |
| 11   | `ExitRowDeclined`     | User answered `n` to the per-row confirm prompt  |

---

## Pre-push Checklist

Run these locally before pushing — they are the same checks CI enforces, so
catching issues here saves a round-trip.

Each step lists the exact CI **workflow → job → step** that re-runs it, plus
where to find the log/artifact when something fails. Workflow files live in
[`.github/workflows/`](.github/workflows/); GitHub Actions logs for a run are
at `https://github.com/<owner>/<repo>/actions/runs/<run-id>` and step logs are
expandable under the named step.

```bash
# 1. Format & vet
#    CI: ci.yml → job "Lint" → step "Go vet"
#    Log: Actions → CI → Lint → "Go vet"
gofmt -l .                # must print nothing
go vet ./...

# 2. Lint (golangci-lint)
#    CI: ci.yml → job "Lint" → step "golangci-lint" (golangci/golangci-lint-action@v6)
#    Log: Actions → CI → Lint → "golangci-lint"
golangci-lint run --timeout=5m

# 3. Identifier-only acronym MixedCaps guard
#    CI: ci.yml → job "Lint" → step "Acronym MixedCaps guard"
#    Log: Actions → CI → Lint → "Acronym MixedCaps guard"
python3 scripts/check-acronym-naming.py

# 4. Legacy module-path auditor (must report 0 ACTIVE)
#    CI: ci.yml → job "Lint" → step "Legacy module-path auditor (strict)"
#    Log:      Actions → CI → Lint → "Legacy module-path auditor (strict)"
#    Artifact: Actions → CI run → Artifacts → "legacy-audit-report"
#              (uploaded by step "Upload legacy audit report", retained 14 days)
bash scripts/audit-legacy-paths.sh --strict

# 5. Project naming guards (canonical name is `movie`)
#    CI: ci.yml → job "Lint" → step "Forbidden term guard"
#                            → step "Binary name consistency check"
#    Auto-fix: bash scripts/check-binary-name.sh --fix [--fuzzy]
bash scripts/guard-forbidden-terms.sh
bash scripts/check-binary-name.sh

# 6. Build & test the whole tree
#    Build CI: ci.yml → job "Build (<os>/<arch>)" → step "Build binary"
#              Artifact: "movie-<os>-<arch>" per matrix entry
#    Test  CI: ci.yml → job "Test (unit)" and "Test (integration)" → step "Run tests"
#              Artifact: "test-results-unit" / "test-results-integration"
#              Summary:  job "Test Summary" → step "Summarize test results"
go build ./...
go test  ./...

# 7. Bump version (any code change requires at least a minor bump)
#    CI: enforced at release time by release.yml (pre-release audit job)
$EDITOR version/info.go
```

One-shot equivalent for steps 1–6:

```bash
bash scripts/pre-release.sh
```

Other related CI jobs (not part of the per-push checklist, but worth knowing
where the logs live):

| Concern | Workflow → Job | Log location |
|---------|----------------|--------------|
| Vulnerability scan (`govulncheck`) | `ci.yml` → **Vulnerability Scan** → "Run govulncheck" | Actions → CI → Vulnerability Scan |
| Cross-compile matrix | `ci.yml` → **Build (\<os\>/\<arch\>)** | Actions → CI → Build → Artifacts |
| Release pre-flight audit | `release.yml` → pre-release audit job | Actions → Release → audit step |
| End-to-end smoke | `e2e.yml` | Actions → E2E |
| Standalone vuln workflow | `vulncheck.yml` | Actions → Vulncheck |

If the acronym guard fails, run the codemod, review the diff, and re-run
the checklist:

```bash
python3 scripts/rename-acronyms.py --write
python3 scripts/check-acronym-naming.py
```


---

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description>

[optional body]
```

### Types

| Type | When to use |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or updating tests |
| `chore` | Build, CI, tooling changes |

### Examples

```
feat(scan): add recursive directory scanning
fix(move): handle cross-device rename with fallback copy
docs(readme): add build-all target to make table
chore(ci): pin golangci-lint to v1.64.8
```

---

## Pull Requests

### Before Submitting

- [ ] Code compiles: `make build`
- [ ] Tests pass: `go test ./... -v -count=1`
- [ ] Linter clean: `go vet ./...`
- [ ] No new files exceed 200 lines
- [ ] Commit messages follow conventions
- [ ] Any new dependency is justified

### PR Description Template

```markdown
## What

Brief description of the change.

## Why

Context and motivation.

## How

Implementation approach and key decisions.

## Testing

How this was tested (commands run, edge cases checked).
```

### Review Process

1. All PRs require at least one approval
2. CI must pass (lint, vulncheck, tests, cross-compile)
3. Maintainers may request changes — address each comment
4. Squash-merge is preferred for clean history

---

## Issue Reporting

### Bug Reports

Include:
- **Movie CLI version**: output of `movie version`
- **OS and architecture**: e.g., Windows 11 amd64, macOS 15 arm64
- **Steps to reproduce**: minimal, specific commands
- **Expected vs actual behavior**
- **Relevant logs or error output**

### Feature Requests

Include:
- **Use case**: what problem does this solve?
- **Proposed solution**: how should it work?
- **Alternatives considered**: what else could work?

---

## Architecture Overview

```
movie-cli-v8/
├── main.go              # Entry point
├── cmd/                  # One file per CLI command
│   ├── scan.go
│   ├── ls.go
│   ├── move.go
│   └── ...
├── movie_info.go         # Shared TMDb/metadata helpers
├── movie_resolve.go      # Shared resolution helpers
├── version/              # Version info (injected at build)
├── data/                 # Runtime data (SQLite DB, posters)
├── spec/                 # Project specifications
├── assets/               # Icons and static assets
└── .github/workflows/    # CI and release pipelines
```

### Key Design Decisions

- **Pure-Go SQLite** (`modernc.org/sqlite`) — no CGo, single binary
- **WAL mode** — concurrent reads during scans
- **TMDb API** — user provides their own API key
- **Self-contained binary** — no external runtime dependencies

---

## Testing

### Running Tests

```bash
# All tests
go test ./... -v -count=1

# With coverage
go test ./... -v -coverprofile=coverage.out -covermode=atomic

# Specific package
go test ./cmd/... -v

# Via Makefile
make tidy
```

### Writing Tests

- Place tests in `_test.go` files alongside the code
- Use table-driven tests for multiple cases
- Mock external APIs (TMDb) — never call real APIs in tests
- Test error paths, not just happy paths

---

## Release Process

Releases are automated via GitHub Actions. Only maintainers create releases.

### Creating a Release

```bash
# Option A: release branch
git checkout -b release/v1.4.0
git push origin release/v1.4.0

# Option B: tag
git tag v1.4.0
git push origin v1.4.0
```

### What Happens

1. CI validates (lint, vuln scan, tests)
2. Cross-compiles 6 binaries (Windows/Linux/macOS × amd64/arm64)
3. Embeds icon into Windows binaries
4. Generates platform-specific install scripts
5. Creates GitHub Release with checksums and changelog

### Version Bumping

- Any code change bumps at least the **minor** version
- Bug-only fixes bump the **patch** version
- Breaking changes bump the **major** version

See [Release Pipeline Spec](spec/pipeline/01-release-pipeline.md) for details.

---

## Maintainer

### [Md. Alim Ul Karim](https://www.google.com/search?q=alim+ul+karim)

**[Creator & Lead Architect](https://alimkarim.com)** | [Chief Software Engineer](https://www.google.com/search?q=alim+ul+karim), [Riseup Asia LLC](https://riseup-asia.com)

|  |  |
|---|---|
| **Website** | [alimkarim.com](https://alimkarim.com/) · [my.alimkarim.com](https://my.alimkarim.com/) |
| **LinkedIn** | [linkedin.com/in/alimkarim](https://linkedin.com/in/alimkarim) |
| **Stack Overflow** | [stackoverflow.com/users/513511/md-alim-ul-karim](https://stackoverflow.com/users/513511/md-alim-ul-karim) |
| **Google** | [Alim Ul Karim](https://www.google.com/search?q=Alim+Ul+Karim) |
| **Role** | Chief Software Engineer, [Riseup Asia LLC](https://riseup-asia.com) — the top software company in Wyoming |

---

## Questions?

Open an issue or start a discussion. We're happy to help!
