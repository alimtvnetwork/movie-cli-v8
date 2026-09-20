<#
.SYNOPSIS
    One-liner uninstaller for movie CLI on Windows.

.DESCRIPTION
    Removes the movie CLI binary, installation directory, User PATH entry,
    and optionally cleans up the ~/.movie user configuration and database.

.PARAMETER InstallDir
    Target directory to clean. Default: $env:LOCALAPPDATA\movie-cli

.PARAMETER KeepData
    Preserve user configuration and database at ~/.movie.

.PARAMETER Yes
    Skip confirmation prompts.

.EXAMPLE
    irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/uninstall-quick.ps1 | iex

.EXAMPLE
    ./uninstall-quick.ps1 -Yes
#>

param(
    [string]$InstallDir = "",
    [switch]$KeepData,
    [switch]$Yes
)

$ErrorActionPreference = "Continue"
$ProgressPreference    = "SilentlyContinue"

function Write-Step($msg) { Write-Host "  $msg" -ForegroundColor Cyan }
function Write-Ok($msg)   { Write-Host "    $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "    $msg" -ForegroundColor Yellow }
function Write-Err($msg)  { Write-Host "    $msg" -ForegroundColor Red }

function Resolve-TargetDir {
    if ($InstallDir) { return $InstallDir }
    $baseAppDir = $env:LOCALAPPDATA
    if (-not $baseAppDir) {
        $baseAppDir = if ($env:HOME) { Join-Path $env:HOME ".local" } else { "." }
    }
    $defaultCandidate = Join-Path $baseAppDir "movie-cli"
    if (Test-Path $defaultCandidate) { return $defaultCandidate }
    $legacyCandidate = Join-Path $baseAppDir "movie"
    if (Test-Path $legacyCandidate) { return $legacyCandidate }
    return $defaultCandidate
}

function Remove-PathEntry([string]$dir) {
    if (-not $dir) { return }
    try {
        $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if (-not $userPath) { return }
        $parts = $userPath -split ";" | Where-Object { $_.Trim() -and ($_.Trim() -ine $dir) }
        $newPath = $parts -join ";"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Ok "Cleaned $dir from User PATH."
    } catch {
        # Ignore on non-Windows
    }
}

function Try-SelfUninstall {
    $cmd = Get-Command movie -ErrorAction SilentlyContinue | Select-Object -First 1
    if (-not $cmd) {
        return $false
    }
    $activeBinary = [string]$cmd.Source
    if ([string]::IsNullOrWhiteSpace($activeBinary)) {
        return $false
    }
    Write-Step "Found active binary: $activeBinary"
    Write-Step "Attempting self-uninstall via: $activeBinary uninstall -y"
    try {
        & $activeBinary uninstall -y
        if ($LASTEXITCODE -eq 0) {
            Write-Ok "Self-uninstall completed cleanly."
            return $true
        }
        return $false
    } catch {
        return $false
    }
}

Write-Host ""
Write-Host "  movie CLI Quick Uninstaller" -ForegroundColor White
Write-Host "  ===========================" -ForegroundColor DarkGray
Write-Host ""

if (-not $InstallDir) {
    if (Try-SelfUninstall) {
        Write-Host ""
        Write-Ok "Uninstall complete."
        exit 0
    }
}

$target = Resolve-TargetDir
Write-Step "Inspecting installation at: $target"

if (Test-Path $target) {
    $exePath = Join-Path $target "movie.exe"
    if (Test-Path $exePath) {
        try {
            Remove-Item $exePath -Force -ErrorAction Stop
            Write-Ok "Removed $exePath"
        } catch {
            Write-Warn "Could not remove $exePath (ensure movie.exe is not currently running)"
        }
    }

    # Clean old backups
    Get-ChildItem -Path $target -Filter "*.old" -File -ErrorAction SilentlyContinue |
        ForEach-Object { Remove-Item $_.FullName -Force -ErrorAction SilentlyContinue }

    $remaining = Get-ChildItem -Path $target -Force -ErrorAction SilentlyContinue
    if (-not $remaining -or $remaining.Count -eq 0) {
        try {
            Remove-Item $target -Force -Recurse -ErrorAction Stop
            Write-Ok "Removed folder $target"
        } catch {
            Write-Warn "Could not remove $target : $_"
        }
    }
} else {
    Write-Step "No installation folder found at $target"
}

Remove-PathEntry $target

$userHome = if ($env:USERPROFILE) { $env:USERPROFILE } elseif ($env:HOME) { $env:HOME } else { "." }
$userData = Join-Path $userHome ".movie"
if (Test-Path $userData) {
    if ($KeepData) {
        Write-Step "Preserving user data: $userData"
    } else {
        $shouldDelete = $Yes
        if (-not $shouldDelete) {
            Write-Host ""
            Write-Host "  Found user configuration & database at $userData" -ForegroundColor Yellow
            Write-Host "  Delete ~/.movie user data? [y/N]: " -ForegroundColor Yellow -NoNewline
            $ans = Read-Host
            if ($ans -match '^(y|yes)$') { $shouldDelete = $true }
        }

        if ($shouldDelete) {
            try {
                Remove-Item $userData -Recurse -Force -ErrorAction Stop
                Write-Ok "Removed user data: $userData"
            } catch {
                Write-Warn "Could not remove $userData : $_"
            }
        } else {
            Write-Step "Preserved user data: $userData"
        }
    }
}

Write-Host ""
Write-Ok "Uninstall complete."
Write-Host ""
