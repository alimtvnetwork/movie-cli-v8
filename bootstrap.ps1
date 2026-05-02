<#
.SYNOPSIS
  Version-discovery bootstrap. Probes sibling repos (-v<N+k>) and delegates
  install to the highest existing one.

.DESCRIPTION
  Implements spec/03-general/05-install-latest-sibling-repo.md.

  Given the starting repo URL embedded below, it probes
  https://raw.githubusercontent.com/<owner>/<base>-v<N+k>/main/install.ps1
  for k = MAX_LOOKAHEAD..0 (highest first). The first HTTP 200 wins; we
  delegate to that install.ps1 via `irm | iex`. If every probe misses, we
  fall back to the starting repo's own install.ps1.

.NOTES
  Invoked by users via:
    irm https://raw.githubusercontent.com/<owner>/<base>-v<N>/main/bootstrap.ps1 | iex

  Each sibling repo ships its own copy of bootstrap.ps1 — they all behave
  identically because the algorithm probes by base name, not by version.

  No retries. No caching. No GitHub API. 5s timeout per probe.
#>

[CmdletBinding()]
param(
    [string]$RepoUrl = "https://github.com/alimtvnetwork/movie-cli-v8"
)

# Do NOT use 'Stop' at script scope — when run via `irm | iex` an uncaught
# terminating error bubbles up to the outer pipeline and surfaces as a
# confusing "Cannot convert 'System.Byte[]' to 'System.String'" message from
# iex. Handle errors locally with try/catch instead.
$ErrorActionPreference = 'Continue'

# ── Constants (spec §3) ───────────────────────────────────────
$MAX_LOOKAHEAD     = 25
$PROBE_TIMEOUT_SEC = 5
$PROBE_BRANCH      = 'main'

# Optional persistent log file
$LogFile = Join-Path $env:TEMP 'movie-bootstrap.log'

# ── Logging ───────────────────────────────────────────────────
function Write-Log {
    param(
        [string]$Message,
        [string]$Color = 'Gray',
        [switch]$NoNewline
    )
    if ($NoNewline) {
        Write-Host $Message -ForegroundColor $Color -NoNewline
    } else {
        Write-Host $Message -ForegroundColor $Color
        try { Add-Content -Path $LogFile -Value $Message -ErrorAction SilentlyContinue } catch {}
    }
}

# ── URL parsing (spec §2) ─────────────────────────────────────
function Get-ParsedRepoUrl {
    param([string]$Url)
    # Strip trailing slash and .git
    $clean = $Url.TrimEnd('/')
    if ($clean.EndsWith('.git')) { $clean = $clean.Substring(0, $clean.Length - 4) }
    # Reject /tree/<branch> etc — only bare repo URLs supported
    if ($clean -match '/tree/' -or $clean -match '/blob/') {
        Write-Log "[bootstrap] error: only bare repo URLs supported (no /tree/ or /blob/)" 'Red'
        return $null
    }
    if ($clean -match '^https://github\.com/([^/]+)/(.+?)-v(\d+)$') {
        return @{
            Owner = $Matches[1]
            Base  = $Matches[2]
            N     = [int]$Matches[3]
            Clean = $clean
        }
    }
    return $null
}

# ── Probe one candidate (spec §4 step 3) ──────────────────────
function Test-InstallScript {
    param([string]$Owner, [string]$Base, [int]$Version)
    $url = "https://raw.githubusercontent.com/$Owner/$Base-v$Version/$PROBE_BRANCH/install.ps1"
    try {
        $resp = Invoke-WebRequest -Uri $url `
                                  -Method Get `
                                  -TimeoutSec $PROBE_TIMEOUT_SEC `
                                  -UseBasicParsing `
                                  -ErrorAction Stop
        if ($resp.StatusCode -eq 200) {
            return @{ Hit = $true; Url = $url; Reason = 'HIT' }
        }
        return @{ Hit = $false; Url = $url; Reason = "status $($resp.StatusCode)" }
    } catch {
        $reason = 'miss'
        if ($_.Exception.Message -match 'timed out|timeout') { $reason = 'timeout' }
        elseif ($_.Exception.Message -match '404')           { $reason = '404' }
        elseif ($_.Exception.Message -match 'resolve|DNS')   { $reason = 'dns' }
        return @{ Hit = $false; Url = $url; Reason = $reason }
    }
}

# ── Find latest sibling (spec §4 step 2-3) ────────────────────
function Find-LatestSibling {
    param([hashtable]$Parsed)
    for ($k = $MAX_LOOKAHEAD; $k -ge 0; $k--) {
        $v = $Parsed.N + $k
        Write-Log "[bootstrap] probing v$v ... " 'DarkGray' -NoNewline
        $r = Test-InstallScript -Owner $Parsed.Owner -Base $Parsed.Base -Version $v
        if ($r.Hit) {
            Write-Log "HIT" 'Green'
            return @{ Version = $v; Url = $r.Url }
        }
        Write-Log "miss ($($r.Reason))" 'DarkGray'
    }
    return $null
}

# ── Delegate to the winner's install.ps1 (spec §4 step 4) ─────
function ConvertTo-ScriptText {
    param($Response)
    $c = $Response.Content
    if ($null -eq $c) { return "" }
    if ($c -is [string]) { return $c }
    if ($c -is [byte[]]) {
        if ($c.Length -ge 3 -and $c[0] -eq 0xEF -and $c[1] -eq 0xBB -and $c[2] -eq 0xBF) {
            return [System.Text.Encoding]::UTF8.GetString($c, 3, $c.Length - 3)
        }
        return [System.Text.Encoding]::UTF8.GetString($c)
    }
    return [string]$c
}

function Invoke-DelegatedInstall {
    param([string]$InstallUrl)
    Write-Log "[bootstrap] delegating to: $InstallUrl" 'Magenta'
    try {
        $script = Invoke-WebRequest -Uri $InstallUrl `
                                    -UseBasicParsing `
                                    -TimeoutSec 30 `
                                    -ErrorAction Stop
    } catch {
        Write-Log "[bootstrap] download failed: $($_.Exception.Message)" 'Red'
        Write-Log "[bootstrap] log file: $LogFile" 'DarkGray'
        return 1
    }
    try {
        $text = ConvertTo-ScriptText $script
        if ([string]::IsNullOrWhiteSpace($text)) {
            Write-Log "[bootstrap] downloaded script body was empty" 'Red'
            return 1
        }
        $sb = [ScriptBlock]::Create($text)
        & $sb
        if ($null -eq $LASTEXITCODE) { return 0 }
        return [int]$LASTEXITCODE
    } catch {
        Write-Log "[bootstrap] delegated install failed: $($_.Exception.Message)" 'Red'
        Write-Log "[bootstrap] log file: $LogFile" 'DarkGray'
        return 1
    }
}

# ── Entry point (fully guarded) ───────────────────────────────
function Invoke-BootstrapMain {
    Write-Log "[bootstrap] starting URL: $RepoUrl" 'Cyan'
    $parsed = Get-ParsedRepoUrl -Url $RepoUrl

    if (-not $parsed) {
        Write-Log "[bootstrap] URL has no -v<N> suffix; installing as-is" 'Yellow'
        $fallback = $RepoUrl.TrimEnd('/').Replace('github.com', 'raw.githubusercontent.com') + "/$PROBE_BRANCH/install.ps1"
        return (Invoke-DelegatedInstall -InstallUrl $fallback)
    }

    Write-Log "[bootstrap] parsed: owner=$($parsed.Owner) base=$($parsed.Base) current=v$($parsed.N)" 'Cyan'
    $winner = Find-LatestSibling -Parsed $parsed

    if (-not $winner) {
        Write-Log "[bootstrap] no -v<N..N+$MAX_LOOKAHEAD> repo found; falling back to starting URL" 'Yellow'
        $fallback = "https://raw.githubusercontent.com/$($parsed.Owner)/$($parsed.Base)-v$($parsed.N)/$PROBE_BRANCH/install.ps1"
        return (Invoke-DelegatedInstall -InstallUrl $fallback)
    }

    Write-Log "[bootstrap] selected: https://github.com/$($parsed.Owner)/$($parsed.Base)-v$($winner.Version)" 'Cyan'
    if ($winner.Version -ne $parsed.N) {
        Write-Log "[bootstrap] auto-upgrade: v$($parsed.N) -> v$($winner.Version)" 'Cyan'
    }
    return (Invoke-DelegatedInstall -InstallUrl $winner.Url)
}

$bootstrapExit = 1
try {
    $bootstrapExit = Invoke-BootstrapMain
    if ($null -eq $bootstrapExit) { $bootstrapExit = 0 }
} catch {
    Write-Log "[bootstrap] unhandled error: $($_.Exception.Message)" 'Red'
    Write-Log "[bootstrap] log file: $LogFile" 'DarkGray'
    $bootstrapExit = 1
}

# Set $LASTEXITCODE for callers but DO NOT call `exit` — `exit` inside an
# `irm | iex` pipeline can terminate the host shell on some PowerShell
# versions. Let the script return naturally.
$global:LASTEXITCODE = [int]$bootstrapExit
