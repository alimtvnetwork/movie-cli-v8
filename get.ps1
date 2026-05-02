<#
.SYNOPSIS
  Smart installer for movie-cli — tries GitHub Release first, then falls back to
  source-build from main, with a clear message either way.

.DESCRIPTION
  Use this when you don't care HOW movie gets installed — only that it does.

  Resolution order:
    1. GitHub Release (releases/latest/download/install.ps1)
       → installs a pre-built binary, no Go toolchain needed.
    2. Source-build fallback (raw.githubusercontent.com/.../main/install.ps1)
       → clones the repo and builds locally; requires Git + Go 1.22+.

  EVERYTHING is wrapped in try/catch and a transcript log is written to
  $env:TEMP\movie-get.log so a crash NEVER takes down the user's terminal
  silently. On any failure the user gets actionable next steps + a log path.

.NOTES
  Invoked via:
    irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/get.ps1 | iex
#>

[CmdletBinding()]
param(
    [string]$DeployPath = ""
)

# IMPORTANT: do NOT set $ErrorActionPreference='Stop' at script scope.
# When this script is run via `irm | iex`, any uncaught terminating error
# bubbles up to the outer pipeline and surfaces as a confusing
# "Cannot convert 'System.Byte[]' to the type 'System.String'" error from
# iex itself. We handle errors locally with try/catch instead.
$ErrorActionPreference = 'Continue'

# ── Config ────────────────────────────────────────────────────
$Owner       = 'alimtvnetwork'
$Repo        = 'movie-cli-v8'
$Branch      = 'main'
$ReleaseUrl  = "https://github.com/$Owner/$Repo/releases/latest/download/install.ps1"
$SourceUrl   = "https://raw.githubusercontent.com/$Owner/$Repo/$Branch/install.ps1"
$ReleasesUI  = "https://github.com/$Owner/$Repo/releases"
$ProbeTimeout = 10
$LogFile     = Join-Path $env:TEMP 'movie-get.log'

# ── Logging ───────────────────────────────────────────────────
function Write-Log {
    param([string]$Message)
    try {
        $stamp = (Get-Date).ToString('yyyy-MM-ddTHH:mm:ss')
        Add-Content -Path $LogFile -Value "[$stamp] $Message" -ErrorAction SilentlyContinue
    } catch { }
}

function Write-Step  { param([string]$M) Write-Host "  -> " -ForegroundColor Cyan -NoNewline; Write-Host $M -ForegroundColor Gray;  Write-Log "STEP $M" }
function Write-Ok    { param([string]$M) Write-Host "  OK " -ForegroundColor Green -NoNewline; Write-Host $M -ForegroundColor Green; Write-Log "OK   $M" }
function Write-Warn  { param([string]$M) Write-Host "  !! " -ForegroundColor Yellow -NoNewline; Write-Host $M -ForegroundColor Yellow; Write-Log "WARN $M" }
function Write-Note  { param([string]$M) Write-Host "     " -NoNewline; Write-Host $M -ForegroundColor DarkGray; Write-Log "     $M" }
function Write-Err   { param([string]$M) Write-Host "  XX " -ForegroundColor Red -NoNewline; Write-Host $M -ForegroundColor Red; Write-Log "ERR  $M" }

function Show-Failure {
    param([string]$Stage, $ErrorRecord)
    Write-Host ""
    Write-Err "Installer failed during: $Stage"
    if ($ErrorRecord) {
        $msg = $ErrorRecord.Exception.Message
        Write-Note "Error: $msg"
        Write-Log  "EXCEPTION ($Stage): $msg"
        if ($ErrorRecord.ScriptStackTrace) {
            Write-Log "STACK: $($ErrorRecord.ScriptStackTrace)"
        }
    }
    Write-Note "Log file: $LogFile"
    Write-Note "Next steps:"
    Write-Note "  1. Open $ReleasesUI to verify a release exists"
    Write-Note "  2. Or clone + build manually:"
    Write-Note "       git clone https://github.com/$Owner/$Repo.git"
    Write-Note "       cd $Repo; pwsh ./install.ps1"
    Write-Host ""
}

# Decode IWR .Content safely. GitHub raw sometimes returns a byte[] (when the
# content-type isn't text/*), and Invoke-Expression can't accept byte[]. Always
# coerce to a UTF-8 string before piping into iex / ScriptBlock::Create.
function ConvertTo-ScriptText {
    param($Response)
    try {
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
    } catch {
        Write-Log "ConvertTo-ScriptText failed: $($_.Exception.Message)"
        return ""
    }
}

# Run a downloaded script in an isolated scriptblock so any error/exit inside
# it cannot terminate THIS script and propagate to the outer `irm | iex`.
function Invoke-RemoteScript {
    param([string]$ScriptText, [hashtable]$BoundArgs = @{})
    if ([string]::IsNullOrWhiteSpace($ScriptText)) {
        throw "Downloaded script body was empty"
    }
    try {
        $sb = [ScriptBlock]::Create($ScriptText)
    } catch {
        throw "Failed to parse downloaded script: $($_.Exception.Message)"
    }
    if ($BoundArgs.Count -gt 0) {
        & $sb @BoundArgs
    } else {
        & $sb
    }
    # Normalise exit code so callers can `return $code` safely.
    if ($null -eq $LASTEXITCODE) { return 0 }
    return [int]$LASTEXITCODE
}

# ── Main, fully wrapped ───────────────────────────────────────
function Invoke-Main {
    Write-Host ""
    Write-Host " +======================================+" -ForegroundColor DarkCyan
    Write-Host " | movie smart installer                |" -ForegroundColor Cyan
    Write-Host " +======================================+" -ForegroundColor DarkCyan
    Write-Host ""
    Write-Log "=== get.ps1 start (DeployPath='$DeployPath') ==="

    # ── 1. Probe the GitHub Release asset ──────────────────────
    Write-Step "Checking for a published GitHub Release..."

    $releaseAvailable = $false
    try {
        $resp = Invoke-WebRequest -Uri $ReleaseUrl `
                                  -Method Head `
                                  -UseBasicParsing `
                                  -TimeoutSec $ProbeTimeout `
                                  -MaximumRedirection 5 `
                                  -ErrorAction Stop
        if ($resp.StatusCode -eq 200) { $releaseAvailable = $true }
    } catch {
        Write-Log "Release probe failed: $($_.Exception.Message)"
        $releaseAvailable = $false
    }

    if ($releaseAvailable) {
        Write-Ok "Release found — installing pre-built binary"
        Write-Note "Source: $ReleaseUrl"
        Write-Host ""
        try {
            $installScript = Invoke-WebRequest -Uri $ReleaseUrl `
                                               -UseBasicParsing `
                                               -TimeoutSec 30 `
                                               -ErrorAction Stop
        } catch {
            Show-Failure -Stage "Downloading release install.ps1" -ErrorRecord $_
            return 1
        }
        try {
            $text = ConvertTo-ScriptText $installScript
            $boundArgs = @{}
            if ($DeployPath) { $boundArgs['DeployPath'] = $DeployPath }
            $code = Invoke-RemoteScript -ScriptText $text -BoundArgs $boundArgs
            Write-Log "Release install completed with code $code"
            return $code
        } catch {
            Show-Failure -Stage "Running release install.ps1" -ErrorRecord $_
            return 1
        }
    }

    # ── 2. Fall back to source-build ──────────────────────────
    Write-Warn "No published GitHub Release found for $Owner/$Repo."
    Write-Note "Falling back to source-build from branch $Branch."
    Write-Note "(This needs Git + Go 1.22+ on PATH. Build takes ~30s.)"
    Write-Host ""
    Write-Note "Tip for maintainers: publish a release at $ReleasesUI"
    Write-Host ""
    Write-Step "Downloading source installer: $SourceUrl"

    try {
        $sourceScript = Invoke-WebRequest -Uri $SourceUrl `
                                          -UseBasicParsing `
                                          -TimeoutSec 30 `
                                          -ErrorAction Stop
    } catch {
        Show-Failure -Stage "Downloading source install.ps1" -ErrorRecord $_
        return 1
    }

    try {
        $text = ConvertTo-ScriptText $sourceScript
        $boundArgs = @{}
        if ($DeployPath) { $boundArgs['DeployPath'] = $DeployPath }
        $code = Invoke-RemoteScript -ScriptText $text -BoundArgs $boundArgs
        Write-Log "Source install completed with code $code"
        return $code
    } catch {
        Show-Failure -Stage "Running source install.ps1" -ErrorRecord $_
        return 1
    }
}

# ── Outermost guard — NOTHING escapes from here ───────────────
$exitCode = 1
try {
    $exitCode = Invoke-Main
    if ($null -eq $exitCode) { $exitCode = 0 }
} catch {
    Show-Failure -Stage "Top-level (unhandled)" -ErrorRecord $_
    $exitCode = 1
}

Write-Log "=== get.ps1 end (exit=$exitCode) ==="
# Set $LASTEXITCODE so callers can inspect it, but DO NOT call `exit` —
# `exit` inside an `irm | iex` pipeline can terminate the host shell on
# some PowerShell versions. Just let the script return naturally.
$global:LASTEXITCODE = [int]$exitCode
