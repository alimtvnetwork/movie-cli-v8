# 05 — Install Latest Sibling Repo (Version-Discovery Bootstrap)

> **Purpose**: When a user runs the install one-liner against a repo URL like
> `…/movie-cli-v8`, the installer should not blindly install **v5**. It should
> first probe sibling repositories (`v6`, `v7`, …) on the same GitHub owner,
> pick the **highest existing** one, and delegate to **that** repo's installer
> script. This way an old install link auto-upgrades to the newest generation
> of the project without the user knowing the URL changed.
>
> This spec is **generic**: the rules apply to any repo whose name ends in a
> `-v<N>` suffix (e.g. `gitmap-v2`, `coding-guidelines-v35`,
> `movie-cli-v8`). Any AI implementing this MUST be able to do so from this
> document alone.

---

## 1. Glossary

| Term | Meaning |
|------|---------|
| **Base name** | The repo name with the trailing `-v<N>` stripped. e.g. `movie-cli-v8` → `movie-cli` |
| **Current version** | The `<N>` in the URL the user invoked. e.g. `movie-cli-v8` → `5` |
| **Candidate** | A repo URL of the form `https://github.com/<owner>/<base>-v<N+k>` for `k = 0..MAX_LOOKAHEAD` |
| **Winner** | The candidate with the **largest** `N+k` whose repo exists AND whose `install.ps1` is reachable |
| **Delegation** | The act of running the winner's `install.ps1` via `irm | iex`, replacing the current install attempt |

---

## 2. Inputs

The bootstrap installer takes a single canonical input: the **starting repo URL**.

```
https://github.com/<OWNER>/<BASE>-v<N>
```

Examples (all valid starting URLs):

| Starting URL | Owner | Base | N |
|--------------|-------|------|---|
| `https://github.com/alimtvnetwork/movie-cli-v8` | `alimtvnetwork` | `movie-cli` | `5` |
| `https://github.com/acme/coding-guidelines-v35` | `acme` | `coding-guidelines` | `35` |
| `https://github.com/foo/gitmap-v2` | `foo` | `gitmap` | `2` |

If the URL does **not** match the `…-v<N>` pattern, fall back to installing
that URL directly (no probing). Log: `URL has no -v<N> suffix; installing as-is.`

---

## 3. Constants

| Name | Default | Notes |
|------|---------|-------|
| `MAX_LOOKAHEAD` | `25` | Probe up to N+25. Tunable per repo family. |
| `PROBE_TIMEOUT_SEC` | `5` | Per HTTP HEAD/GET; **fail fast**, no retries. |
| `PROBE_BRANCH` | `main` | Branch where `install.ps1` is expected. |

These constants are hard-coded in the bootstrap. No config file, no flags,
no env vars — keep it dead simple.

---

## 4. Algorithm

```
1.  Parse starting URL → (owner, base, N)
    └─ if no -v<N> suffix: install starting URL directly, STOP.

2.  Build candidate list, HIGHEST FIRST:
        for k in MAX_LOOKAHEAD..0:
            candidate = "https://github.com/{owner}/{base}-v{N+k}"

3.  For each candidate (highest first):
        a. probe install.ps1 raw URL:
             https://raw.githubusercontent.com/{owner}/{base}-v{N+k}/{PROBE_BRANCH}/install.ps1
        b. issue HTTP GET (or HEAD if server supports it) with PROBE_TIMEOUT_SEC.
        c. on HTTP 200  →  this is the WINNER. break.
        d. on 404 / timeout / DNS / TLS error  →  log "miss v{N+k}", continue.
        e. on any other error (5xx, network down) →  log + continue.

4.  If a WINNER was found:
        - log "selected: {winner-url}"
        - if winner version != starting version, log
              "auto-upgraded v{N} → v{N+k}"
        - download winner's install.ps1 and execute it (irm | iex).
        - exit with the installer's exit code.

5.  If NO WINNER was found (every candidate missed):
        - log "no -v<N..N+MAX> repo found; falling back to starting URL"
        - install starting URL directly.
```

### Fail-fast policy

- **No retries.** A timeout or 404 is treated as "this candidate doesn't exist."
- **No HTTP cache.** Each probe is a fresh request.
- **No GitHub API.** Use only the raw.githubusercontent.com URL — no auth,
  no rate-limit risk.
- Total probe budget is `MAX_LOOKAHEAD * PROBE_TIMEOUT_SEC` worst case
  (default ≈ 125s if every probe times out — acceptable, and in practice
  most probes return 404 instantly).

---

## 5. Logging

Every bootstrap run MUST emit a numbered, timestamped trace so users (and
future AIs debugging) can see exactly what was probed and chosen.

### Required log lines

```
[bootstrap] starting URL: https://github.com/alimtvnetwork/movie-cli-v8
[bootstrap] parsed: owner=alimtvnetwork base=movie-cli current=v5
[bootstrap] probing v30 ... miss (404)
[bootstrap] probing v29 ... miss (timeout)
...
[bootstrap] probing v7  ... HIT
[bootstrap] selected: https://github.com/alimtvnetwork/movie-cli-v8
[bootstrap] auto-upgrade: v5 -> v7
[bootstrap] delegating to: https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install.ps1
```

### Log destinations

- **Always** `Write-Host` to stdout (visible in the user's terminal).
- **Optionally** append to `$env:TEMP/movie-bootstrap.log` on Windows or
  `/tmp/movie-bootstrap.log` on Unix for post-mortem.

Color hints (PowerShell):
- `miss` → DarkGray
- `HIT` → Green
- `selected` / `auto-upgrade` → Cyan
- `delegating` → Magenta

---

## 6. PowerShell Reference Implementation Sketch

> This is a **sketch**, not the final code. Any AI implementing the bootstrap
> MUST follow Section 4's algorithm exactly; this just shows the shape.

```powershell
# bootstrap.ps1 — entry point invoked via:
#   irm https://raw.githubusercontent.com/<owner>/<base>-v<N>/main/bootstrap.ps1 | iex

$ErrorActionPreference = 'Stop'
$MAX_LOOKAHEAD     = 25
$PROBE_TIMEOUT_SEC = 5
$PROBE_BRANCH      = 'main'

function Parse-RepoUrl {
    param([string]$Url)
    if ($Url -match '^https://github\.com/([^/]+)/(.+?)-v(\d+)/?$') {
        return @{ Owner = $Matches[1]; Base = $Matches[2]; N = [int]$Matches[3] }
    }
    return $null
}

function Test-InstallScript {
    param([string]$Owner, [string]$Base, [int]$Version)
    $url = "https://raw.githubusercontent.com/$Owner/$Base-v$Version/$PROBE_BRANCH/install.ps1"
    try {
        $null = Invoke-WebRequest -Uri $url -Method Get -TimeoutSec $PROBE_TIMEOUT_SEC -UseBasicParsing
        return @{ Hit = $true; Url = $url }
    } catch {
        return @{ Hit = $false; Url = $url }
    }
}

function Find-LatestSibling {
    param([hashtable]$Parsed)
    for ($k = $MAX_LOOKAHEAD; $k -ge 0; $k--) {
        $v = $Parsed.N + $k
        Write-Host "[bootstrap] probing v$v ... " -NoNewline -ForegroundColor DarkGray
        $r = Test-InstallScript -Owner $Parsed.Owner -Base $Parsed.Base -Version $v
        if ($r.Hit) {
            Write-Host "HIT" -ForegroundColor Green
            return @{ Version = $v; Url = $r.Url }
        }
        Write-Host "miss" -ForegroundColor DarkGray
    }
    return $null
}

# --- entry point -------------------------------------------------
param([Parameter(Mandatory)][string]$RepoUrl)

Write-Host "[bootstrap] starting URL: $RepoUrl" -ForegroundColor Cyan
$parsed = Parse-RepoUrl -Url $RepoUrl
if (-not $parsed) {
    Write-Host "[bootstrap] URL has no -v<N> suffix; installing as-is" -ForegroundColor Yellow
    irm "$RepoUrl/raw/main/install.ps1" | iex
    exit $LASTEXITCODE
}

Write-Host "[bootstrap] parsed: owner=$($parsed.Owner) base=$($parsed.Base) current=v$($parsed.N)" -ForegroundColor Cyan
$winner = Find-LatestSibling -Parsed $parsed

if (-not $winner) {
    Write-Host "[bootstrap] no -v<N..N+$MAX_LOOKAHEAD> repo found; falling back" -ForegroundColor Yellow
    irm "https://raw.githubusercontent.com/$($parsed.Owner)/$($parsed.Base)-v$($parsed.N)/$PROBE_BRANCH/install.ps1" | iex
    exit $LASTEXITCODE
}

Write-Host "[bootstrap] selected: https://github.com/$($parsed.Owner)/$($parsed.Base)-v$($winner.Version)" -ForegroundColor Cyan
if ($winner.Version -ne $parsed.N) {
    Write-Host "[bootstrap] auto-upgrade: v$($parsed.N) -> v$($winner.Version)" -ForegroundColor Cyan
}
Write-Host "[bootstrap] delegating to: $($winner.Url)" -ForegroundColor Magenta
irm $winner.Url | iex
exit $LASTEXITCODE
```

### Bash equivalent (for `curl | bash` users)

The same algorithm applies. Use `curl -fsI --max-time 5` for the probe;
`HTTP/2 200` means HIT, anything else means miss. Delegation:
`curl -fsSL "$winnerUrl" | bash`.

---

## 7. User-Facing One-Liners

Once `bootstrap.ps1` exists in **every** sibling repo (so the user can copy
any one of them and it still works), the public install command becomes:

### Windows / PowerShell

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/bootstrap.ps1 | iex
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/bootstrap.sh | bash
```

The user can paste the v5 URL forever — bootstrap will silently jump to
v6, v7, … as they get released.

---

## 8. Edge Cases & Decisions

| Case | Decision |
|------|----------|
| User passes URL with trailing `.git` | Strip `.git` before parsing |
| User passes URL with `/tree/<branch>` | Reject; only bare repo URLs supported |
| `install.ps1` exists but is broken | Out of scope. Bootstrap only checks **existence**, not validity. |
| Two consecutive misses then a hit (gap in versions, e.g. v5 missing, v6 exists) | Allowed. We probe HIGH→LOW so v6 wins regardless of v5. |
| `MAX_LOOKAHEAD` exceeded by real release count | Bump constant in this spec + bootstrap. Document the bump. |
| User is offline | Every probe times out → falls back to starting URL → that probably also fails → installer prints clear network error. Acceptable. |
| Private repos | Out of scope. Bootstrap uses unauthenticated raw.githubusercontent.com. |

---

## 9. Constraints (Inherited)

This spec inherits the **installer subshell constraint** (see
`mem://constraints/installer-subshell`): the bootstrap runs inside
`irm | iex`, so it CANNOT mutate the parent shell's PATH. The downstream
`install.ps1` is responsible for writing PATH and printing the
copy-pasteable refresh hint.

---

## 10. Acceptance Criteria

- GIVEN URL `…/movie-cli-v8` AND v7 is the highest existing sibling
  WHEN bootstrap runs
  THEN it logs probes for v30..v8 as misses, v7 as HIT, and delegates to v7's `install.ps1`.

- GIVEN URL `…/movie-cli-v8` AND only v5 exists
  WHEN bootstrap runs
  THEN it logs misses for v30..v6, HIT for v5, and delegates to v5's `install.ps1`.

- GIVEN URL `…/movie-cli-v8` AND no sibling exists (network failure or all 404)
  WHEN bootstrap runs
  THEN it logs the failure and falls back to the starting URL's `install.ps1`.

- GIVEN URL `…/some-repo-without-suffix`
  WHEN bootstrap runs
  THEN it skips probing and installs the starting URL directly.

- GIVEN any successful run
  THEN every probe attempt MUST appear in stdout with v-number and miss/HIT verdict.

---

## 11. Out of Scope

- Semantic version comparison (we only compare the integer `<N>`).
- Pre-release / beta channels.
- Mirror fallbacks (only github.com is probed).
- Auto-rollback if delegated installer fails — that's the installer's job.

---

*Spec authored: 2026-04-18 — see `mem://constraints/installer-subshell` for related constraint.*
