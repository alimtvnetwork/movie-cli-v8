<#
.SYNOPSIS
    Quick interactive installer for movie CLI on Windows.

.DESCRIPTION
    Prompts for an install folder (default: $env:LOCALAPPDATA\movie-cli),
    then delegates to the canonical install.ps1 with the target path.

    Run via one-liner:
      irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install-quick.ps1 | iex

    Or locally:
      ./install-quick.ps1
      ./install-quick.ps1 -InstallDir "$env:LOCALAPPDATA\movie-cli"
#>

param(
    [string]$InstallDir = "",
    [string]$Version = "",
    [switch]$Interactive,
    [switch]$NoDiscovery,
    [string]$LogFile = ""
)

$ErrorActionPreference = "Stop"
$ProgressPreference    = "SilentlyContinue"

$Repo         = "alimtvnetwork/movie-cli-v8"
$InstallerUrl = "https://raw.githubusercontent.com/$Repo/main/install.ps1"
$DefaultDir   = Join-Path $env:LOCALAPPDATA "movie-cli"

if ([string]::IsNullOrWhiteSpace($LogFile)) {
    $stamp   = (Get-Date).ToString("yyyyMMdd-HHmmss")
    $LogFile = Join-Path $env:TEMP "movie-install-quick-$stamp.log"
}

function Write-Log([string]$message, [string]$level = "INFO") {
    $line = "[{0}] [{1}] {2}" -f (Get-Date -Format "yyyy-MM-dd HH:mm:ss"), $level, $message
    try { Add-Content -Path $LogFile -Value $line -Encoding UTF8 -ErrorAction SilentlyContinue } catch { }
}

function Read-InstallDir([string]$default) {
    Write-Host ""
    Write-Host "  movie CLI quick installer" -ForegroundColor Cyan
    Write-Host "  ------------------------" -ForegroundColor DarkGray
    Write-Host "  Choose install directory. Press Enter to accept the default." -ForegroundColor Gray
    Write-Host "  Default: $default" -ForegroundColor DarkGray

    $answer = Read-Host "  Install path"
    if ([string]::IsNullOrWhiteSpace($answer)) { return $default }
    return $answer.Trim('"').Trim()
}

if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    if ($Interactive) {
        $InstallDir = Read-InstallDir $DefaultDir
    } else {
        $InstallDir = $DefaultDir
    }
}

Write-Host ""
Write-Host "  Installing movie CLI to: $InstallDir" -ForegroundColor Green
Write-Host "  Log file: $LogFile" -ForegroundColor DarkGray
Write-Host ""

# Check if local install.ps1 is available first (when running from repo)
$localInstaller = Join-Path $PSScriptRoot "install.ps1"
$scriptContent = $null

if (Test-Path $localInstaller) {
    $scriptContent = Get-Content $localInstaller -Raw -Encoding UTF8
} else {
    Write-Log "Downloading canonical installer: $InstallerUrl"
    $scriptContent = (Invoke-WebRequest -Uri $InstallerUrl -UseBasicParsing).Content
}

$block = [ScriptBlock]::Create($scriptContent)
$passArgs = @{ InstallDir = $InstallDir }
if (-not [string]::IsNullOrWhiteSpace($Version)) { $passArgs.Version = $Version }
if ($NoDiscovery) { $passArgs.NoDiscovery = $true }

try {
    & $block @passArgs
    Write-Log "install-quick.ps1 completed successfully"
} catch {
    Write-Log "Error during install: $_" "ERROR"
    Write-Host "  [ERROR] Install failed: $_" -ForegroundColor Red
    exit 1
}
