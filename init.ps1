<#
.SYNOPSIS
  One-shot repo init and preflight validation for movie CLI.

.DESCRIPTION
  Order:
    1) Environment & Go toolchain preflight check
    2) Deploy manifest verification (deploy-manifest.json)
    3) Installer dry-run contract check (install.ps1 -DryRun)

  Pass -DryRun to preview both steps without modifying environment.

.EXAMPLE
  .\init.ps1
  .\init.ps1 -DryRun
#>

[CmdletBinding()]
param(
    [switch]$DryRun,
    [Alias('h')][switch]$Help
)

$ErrorActionPreference = 'Continue'
$Script:HereDir = Split-Path -Parent $MyInvocation.MyCommand.Path

function Show-InitHelp {
    @"
init.ps1 - One-shot repo init and preflight validation for movie-cli.

Usage:
  .\init.ps1            # run toolchain check and installer dry-run validation
  .\init.ps1 -DryRun    # preview steps without environment impact
"@ | Write-Host
}

if ($Help) { Show-InitHelp; exit 0 }

Write-Host ""
Write-Host "==> [movie-cli init] Starting preflight validation..." -ForegroundColor Cyan

# Step 1: Deploy manifest
$manifestPath = Join-Path $Script:HereDir "deploy-manifest.json"
if (Test-Path $manifestPath) {
    Write-Host "  [OK] deploy-manifest.json found." -ForegroundColor Green
} else {
    Write-Host "  [WARN] deploy-manifest.json missing." -ForegroundColor Yellow
}

# Step 2: Go toolchain
$go = Get-Command go -ErrorAction SilentlyContinue
if ($go) {
    $goVer = go version
    Write-Host "  [OK] $goVer" -ForegroundColor Green
} else {
    Write-Host "  [WARN] 'go' command not found on PATH." -ForegroundColor Yellow
}

# Step 3: Installer dry-run
$installerPath = Join-Path $Script:HereDir "install.ps1"
if (Test-Path $installerPath) {
    Write-Host "  [TEST] Running install.ps1 -DryRun..." -ForegroundColor Cyan
    & $installerPath -DryRun -NoDiscovery
    $installerRc = $LASTEXITCODE
    if ($installerRc -eq 0) {
        Write-Host "  [OK] install.ps1 dry-run contract passed." -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] install.ps1 dry-run failed with exit code $installerRc" -ForegroundColor Red
        exit $installerRc
    }
}

Write-Host ""
Write-Host "==> [movie-cli init] Preflight validation completed successfully." -ForegroundColor Green
exit 0
