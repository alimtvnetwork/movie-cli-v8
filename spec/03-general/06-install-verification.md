# 06 — Install Verification

> Post-install sanity checks for the **movie** CLI. Run after `install.sh` /
> `install.ps1` to confirm prerequisites, binary placement, and execution.

## Scripts

| Platform | Script | Invoke |
|----------|--------|--------|
| Linux / macOS | `verify.sh` | `bash verify.sh` |
| Windows / pwsh | `verify.ps1` | `pwsh verify.ps1` |

## What it checks

1. **Prerequisites**
   - `git` present in `PATH`
   - `go` present and `>= 1.22` (warning only — not required if installing from a release binary)
2. **Binary**
   - `movie` resolvable on `PATH` (or in `--dir` / `-Dir` when supplied)
   - File is executable
3. **Execution**
   - `movie version` exits 0 and prints a `vX.Y.Z` string
   - `movie --help` (or `movie help`) exits 0

## Flags

| Flag | Purpose | Default |
|------|---------|---------|
| `--binary` / `-Binary` | Binary name to probe | `movie` |
| `--dir` / `-Dir` | Specific directory to probe instead of `PATH` | (use `PATH`) |
| `-h`, `--help` | Show usage | — |

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | All required checks passed (warnings allowed) |
| `1` | One or more checks failed |
| `2` | Bad usage (bash only) |

## Output format

Three sections — `Prerequisites`, `Binary`, `Execution` — followed by a
summary line: `<N> passed  <N> warnings  <N> failed`.

Markers: `[PASS]` green, `[WARN]` yellow, `[FAIL]` red.

## Recommended usage

Append to the install one-liner:

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install.sh | bash \
  && curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/verify.sh | bash
```

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install.ps1 | iex
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/verify.ps1 | iex
```

## Constraints

- Verification scripts are **read-only**: no installs, no edits, no PATH writes.
- Subshell limits still apply (see `mem://constraints/installer-subshell.md`):
  if `movie` was just installed in the same pipe and `PATH` was only updated
  for future shells, `verify` may report `not found on PATH`. Print the
  copy-paste hint and re-run.
- No timestamp output, no git-update suggestions
  (per `spec/00-spec-authoring-guide/10-strictly-prohibited.md`).
