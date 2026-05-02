# 07 — Legacy Module Path Audit

> Read-only auditor that scans the repo for any reference to legacy module
> paths (`movie-cli-v8` … `movie-cli-v8`) left behind after the migration to
> `movie-cli-v8`.

## Why

The project has been renamed multiple times: v1 → v2 → v3 → v4 → v5 → v6 → **v7**.
Every rename risks leaving stale `github.com/alimtvnetwork/movie-cli-v<N>`
strings in Go imports, install scripts, README links, CI workflows, or
release URLs. Stale references silently break `go get`, `go build`,
`curl | bash` installs, and update flows.

This script gives you a single-command answer: *is the migration complete,
and where are the remaining mentions allowed to be?*

## Run it

```bash
bash scripts/audit-legacy-paths.sh
```

Optional flags:

| Flag       | Effect |
|------------|--------|
| `--strict` | Exit `1` if any **active** reference is found (CI-friendly). |
| `--json`   | Emit a machine-readable report to stdout (no jq required). |
| `--help`   | Print usage. |

## What it checks

Greps the entire repo (excluding `.git`, `node_modules`, `.release`, `.gitmap`)
for the regex `movie-cli-v[123456]\b`, then **classifies** each match:

| Category       | Where | Verdict |
|----------------|-------|---------|
| **HISTORICAL** | `CHANGELOG.md`, `spec/**`, `.lovable/**`, this script | ✅ Allowed — these document past renames and rename audits. |
| **ACTIVE**     | Everything else: `README.md`, `QUICKSTART.md`, `install*.{sh,ps1}`, `get*.{sh,ps1}`, `verify*.{sh,ps1}`, `bootstrap*.{sh,ps1}`, `scripts/`, `*.go`, `go.mod`, `.github/workflows/**`, `Makefile`, `powershell.json`, etc. | ❌ Must fix — replace with `github.com/alimtvnetwork/movie-cli-v8`. |

## Sample output

```
════════════════════════════════════════════════════════════════
  Legacy module-path audit — pattern: movie-cli-v[123456]\b
════════════════════════════════════════════════════════════════

── HISTORICAL (allowed) ─────────────────  6 match(es)
  ./CHANGELOG.md:76:- **Repo-root `install.ps1`** — `RepoUrl` was pointing …
  ./CHANGELOG.md:403:- **Module path renamed** — `…movie-cli-v8` → `…v6` …
  ./spec/12-ci-cd-pipeline/05-ci-cd-issues/06-release-missing-asset-404.md:90:…

── ACTIVE (must fix) ─────────────────────  0 match(es)
  ✅ none

── Summary ──────────────────────────────────────────────────────
  Historical : 6  (informational, OK to keep)
  Active     : 0  (clean)
```

## Relationship to CI guards

The audit script overlaps with — but is **broader than** — the two CI guards:

| Surface | CI guard | Audit script |
|---------|----------|--------------|
| Go source / `go.mod` / configs | `Old module path regression guard` (bans `v[23456]`) | ✅ also covers v1, with classification |
| `README.md`, `QUICKSTART.md`, install scripts | `User-facing docs v6 reference guard` (v6 only) | ✅ covers v1-v6 across the whole repo |
| Historical mentions in CHANGELOG/spec | excluded | reported as **HISTORICAL**, not a failure |

Use the audit script locally before raising a PR, or wire it into CI with
`--strict` for an extra belt-and-braces check.

## Adding a new historical surface

If you legitimately need to mention an old path in a new file (e.g. another
spec doc), add the path prefix to `HISTORICAL_FILTER` inside
`scripts/audit-legacy-paths.sh`. Keep the allowlist tight — every new entry
weakens the safety net.
