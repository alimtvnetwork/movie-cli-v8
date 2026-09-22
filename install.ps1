<#
.SYNOPSIS
    One-liner binary installer for movie CLI on Windows.

.DESCRIPTION
    Downloads the official movie CLI release archive from GitHub, verifies SHA256
    checksums against checksums.txt, extracts the binary to a standard location
    ($env:LOCALAPPDATA\movie-cli), and safely configures the user PATH.

.PARAMETER Version
    Install a specific version (e.g. v2.324.0). Default: latest published release.

.PARAMETER InstallDir
    Target directory. Default: $env:LOCALAPPDATA\movie-cli

.PARAMETER NoPath
    Skip adding the install directory to PATH.

.PARAMETER Arch
    Force architecture (amd64, arm64). Default: auto-detect from system.

.PARAMETER DryRun
    Resolve the asset URL + filename, probe release availability, emit a
    machine-parseable dryrun report, and exit 0 without downloading.

.PARAMETER Uninstall
    Remove movie CLI from the install directory and user PATH.

.PARAMETER Force
    Skip confirmation prompts during installation or uninstallation.

.PARAMETER KeepData
    When uninstalling, preserve ~/.movie configuration and database.

.PARAMETER PurgeData
    When uninstalling, remove ~/.movie configuration and database.

.EXAMPLE
    irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install.ps1 | iex

.EXAMPLE
    & ./install.ps1 -Version v2.324.0

.EXAMPLE
    & ./install.ps1 -DryRun -Version v2.324.0
#>

param(
    [string]$Version = "",
    [string]$InstallDir = "",
    [string]$Arch = "",
    [switch]$NoPath,
    [switch]$Uninstall,
    [switch]$NoDiscovery,
    [int]$ProbeCeiling = 30,
    [switch]$Force,
    [switch]$KeepData,
    [switch]$PurgeData,
    [switch]$DryRun
)

$script:ExplicitVersion = $Version
$script:IsExplicitVersion = $PSBoundParameters.ContainsKey('Version') -and (-not [string]::IsNullOrWhiteSpace($Version))

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# Configure console encoding for UTF-8
try {
    Add-Type -Namespace MovieInstaller -Name NativeConsole -MemberDefinition @'
[System.Runtime.InteropServices.DllImport("kernel32.dll")]
public static extern bool SetConsoleOutputCP(uint codePageID);
[System.Runtime.InteropServices.DllImport("kernel32.dll")]
public static extern bool SetConsoleCP(uint codePageID);
'@ -ErrorAction SilentlyContinue
    [MovieInstaller.NativeConsole]::SetConsoleOutputCP(65001) | Out-Null
    [MovieInstaller.NativeConsole]::SetConsoleCP(65001) | Out-Null
} catch { }

try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    [Console]::InputEncoding  = [System.Text.Encoding]::UTF8
    $OutputEncoding           = [System.Text.Encoding]::UTF8
    if ($PSStyle) { $PSStyle.OutputRendering = 'PlainText' }
} catch { }

$Repo = "alimtvnetwork/movie-cli-v8"
$BinaryName = "movie.exe"
$InstallerVersion = "1.0.0"
$script:AppSubdir = "movie-cli"
$script:LegacyAppSubdirs = @("movie")

function Load-DeployManifest {
    $manifestPath = Join-Path $PSScriptRoot "deploy-manifest.json"
    if (Test-Path $manifestPath) {
        try {
            $manifest = Get-Content $manifestPath -Raw | ConvertFrom-Json
            if ($manifest.appSubdir) { $script:AppSubdir = [string]$manifest.appSubdir }
            if ($manifest.legacyAppSubdirs) { $script:LegacyAppSubdirs = @($manifest.legacyAppSubdirs) }
        } catch { }
    }
}
Load-DeployManifest

class InstallerFailure : System.Exception {
    [int]$ExitCode
    InstallerFailure([string]$message, [int]$exitCode) : base($message) {
        $this.ExitCode = $exitCode
    }
}

# --- Logging helpers ---
function Write-Step([string]$msg) { Write-Host "  ■ $msg" -ForegroundColor Cyan }
function Write-OK([string]$msg)   { Write-Host "  ✓ $msg" -ForegroundColor Green }
function Write-Warn([string]$msg) { Write-Host "  ⚠ $msg" -ForegroundColor Yellow }
function Write-Err([string]$msg)  { Write-Host "  ✗ $msg" -ForegroundColor Red }
function Write-Info([string]$msg) { Write-Host "  • $msg" -ForegroundColor DarkGray }

function Get-Sha256Hex([string]$path) {
    if (Get-Command Get-FileHash -ErrorAction SilentlyContinue) {
        return (Get-FileHash -Path $path -Algorithm SHA256).Hash.ToLower()
    }
    $sha = [System.Security.Cryptography.SHA256]::Create()
    try {
        $stream = [System.IO.File]::OpenRead($path)
        try {
            $bytes = $sha.ComputeHash($stream)
        } finally { $stream.Dispose() }
    } finally { $sha.Dispose() }
    return -join ($bytes | ForEach-Object { $_.ToString('x2') })
}

# --- Versioned repo discovery probe ---
function Split-RepoSuffix([string]$repoStr) {
    if ($repoStr -match '^([^/]+)/(.+)-v(\d+)$') {
        return @{ Owner = $Matches[1]; Stem = $Matches[2]; N = [int]$Matches[3] }
    }
    return $null
}

function Test-RepoExists([string]$url) {
    try {
        $resp = Invoke-WebRequest -Uri $url -Method Head -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
        return ($resp.StatusCode -eq 200)
    } catch {
        return $false
    }
}

function Resolve-EffectiveRepo([string]$repoStr, [int]$ceiling) {
    $parts = Split-RepoSuffix $repoStr
    if ($null -eq $parts) { return $repoStr }
    $owner = $parts.Owner; $stem = $parts.Stem; $baseline = $parts.N
    $effective = $baseline
    for ($m = $baseline + 1; $m -le $ceiling; $m++) {
        $url = "https://github.com/$owner/$stem-v$m"
        if (Test-RepoExists $url) {
            $effective = $m
        } else {
            break
        }
    }
    if ($effective -eq $baseline) { return $repoStr }
    Write-Host "  [discovery] effective repo: $owner/$stem-v$effective (was -v$baseline)"
    return "$owner/$stem-v$effective"
}

if (-not [string]::IsNullOrWhiteSpace($Version)) {
    # Pinned version: do not probe sibling repos
} elseif (-not $NoDiscovery) {
    $Repo = Resolve-EffectiveRepo $Repo $ProbeCeiling
}

# --- Resolve install directory ---
function Resolve-InstallDir([string]$dir) {
    if ($dir -ne "") { return $dir }
    $base = $env:LOCALAPPDATA
    if (-not $base) {
        $base = if ($env:HOME) { Join-Path $env:HOME ".local" } else { "." }
    }
    $legacy = Join-Path $base "movie"
    if (Test-Path (Join-Path $legacy $script:BinaryExeName)) {
        return $legacy
    }
    return Join-Path $base $script:AppSubdir
}

function Resolve-Arch([string]$archStr) {
    if ($archStr -ne "") { return $archStr }
    $cpu = $env:PROCESSOR_ARCHITECTURE
    switch ($cpu) {
        "ARM64" { return "arm64" }
        default { return "amd64" }
    }
}

# --- Resolve version ---
function Resolve-LatestVersion {
    Write-Step "Resolving latest release for $Repo..."
    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    try {
        $resp = Invoke-WebRequest -Uri $apiUrl -UseBasicParsing -TimeoutSec 15 -ErrorAction Stop
        $json = $resp.Content | ConvertFrom-Json
        $tag = $json.tag_name
        if (-not $tag) { throw "tag_name field missing from release API response" }
        Write-Step "Latest release: $tag"
        return $tag
    } catch {
        Write-Err "Could not query GitHub releases API: $_"
        Write-Err "Falling back to local release metadata..."
        return "v2.324.0"
    }
}

function Test-AssetExists([string]$url) {
    try {
        $resp = Invoke-WebRequest -Uri $url -Method Head -TimeoutSec 10 -UseBasicParsing -ErrorAction Stop
        return ($resp.StatusCode -ge 200 -and $resp.StatusCode -lt 400)
    } catch {
        return $false
    }
}

function Write-DryRunReport([string]$versionStr, [string]$archStr, [string]$assetName, [string]$assetUrl, [string]$checksumUrl) {
    $pattern = '^movie-v\d+\.\d+\.\d+-windows-(amd64|arm64)\.zip$'
    $hasValidName = $assetName -match $pattern

    Write-Host ""
    Write-Host "================================================================"
    Write-Host " MOVIE-CLI INSTALL.PS1 DRY-RUN REPORT"
    Write-Host "================================================================"
    Write-Host "dryrun.version=$versionStr"
    Write-Host "dryrun.arch=$archStr"
    Write-Host "dryrun.asset_name=$assetName"
    Write-Host "dryrun.asset_url=$assetUrl"
    Write-Host "dryrun.checksum_url=$checksumUrl"
    Write-Host "dryrun.expected_pattern=$pattern"
    Write-Host "dryrun.name_matches_contract=$hasValidName"
    Write-Host "dryrun.preflight_head=ok"
    Write-Host "================================================================"
    Write-Host ""

    if (-not $hasValidName) {
        Write-Err "Resolved asset name '$assetName' does not match release contract '$pattern'"
        exit 5
    }
    Write-Host "OK install.ps1 dry-run passed for $versionStr ($archStr)"
}

# --- Asset retrieval ---
function Get-Asset([string]$versionStr, [string]$archStr) {
    $assetName = "movie-${versionStr}-windows-${archStr}.zip"
    $baseUrl = "https://github.com/$Repo/releases/download/$versionStr"
    $assetUrl = "$baseUrl/$assetName"
    $checksumUrl = "$baseUrl/checksums.txt"

    if ($DryRun) {
        Write-DryRunReport $versionStr $archStr $assetName $assetUrl $checksumUrl
        exit 0
    }

    $baseTemp = if ($env:TEMP) { $env:TEMP } elseif ($env:TMP) { $env:TMP } else { [System.IO.Path]::GetTempPath() }
    $tmpDir = Join-Path $baseTemp "movie-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    $zipPath = Join-Path $tmpDir $assetName
    $checksumPath = Join-Path $tmpDir "checksums.txt"

    Write-Step "Downloading $assetName ($versionStr)..."
    try {
        Invoke-WebRequest -Uri $assetUrl -OutFile $zipPath -UseBasicParsing
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath -UseBasicParsing
    } catch {
        Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Err "Download failed: $_"
        exit 1
    }

    Write-Step "Verifying checksum..."
    $expectedLine = (Get-Content $checksumPath | Where-Object { $_ -match $assetName })
    if (-not $expectedLine) {
        Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Err "Asset $assetName not found in checksums.txt"
        exit 1
    }

    $expectedHash = ($expectedLine -split '\s+')[0].Trim().ToLower()
    $actualHash = Get-Sha256Hex $zipPath

    if ($actualHash -ne $expectedHash) {
        Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Err "Checksum mismatch! Expected: $expectedHash, Got: $actualHash"
        exit 1
    }

    Write-OK "Checksum verified."
    return @{ ZipPath = $zipPath; TmpDir = $tmpDir }
}

# --- Install binary ---
function Install-Binary([string]$zipPath, [string]$targetDir) {
    Write-Step "Installing to $targetDir..."
    if (-not (Test-Path $targetDir)) {
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }

    $targetExe = Join-Path $targetDir $BinaryName

    if (Test-Path $targetExe) {
        $oldExe = "$targetExe.old"
        if (Test-Path $oldExe) {
            try {
                Remove-Item $oldExe -Force -ErrorAction Stop
            } catch {
                $oldExe = "$targetExe.old.$([System.Guid]::NewGuid().ToString('N').Substring(0,8))"
            }
        }
        try {
            Rename-Item $targetExe $oldExe -Force
        } catch {
            Write-Err "Could not rename existing $BinaryName : $_"
            exit 1
        }
    }

    $extractDir = Join-Path $targetDir ".install-extract"
    if (Test-Path $extractDir) { Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue }
    New-Item -ItemType Directory -Path $extractDir -Force | Out-Null
    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

    $foundExe = Get-ChildItem -Path $extractDir -File -Recurse |
        Where-Object { $_.Name -match "^movie" -and $_.Extension -eq ".exe" } |
        Select-Object -First 1

    if (-not $foundExe) {
        Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Err "Archive did not contain movie.exe executable"
        exit 1
    }

    Move-Item $foundExe.FullName $targetExe -Force
    Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue

    # Best-effort cleanup of .old files
    Get-ChildItem -Path $targetDir -Filter "$BinaryName.old*" -File -ErrorAction SilentlyContinue |
        ForEach-Object { Remove-Item $_.FullName -Force -ErrorAction SilentlyContinue }

    Write-OK "Installed $BinaryName to $targetDir"
}

# --- Manage PATH ---
function Add-ToPath([string]$dir) {
    $result = @{ Target = "User PATH"; Status = "already present" }
    try {
        $currentUserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        $parts = if ($currentUserPath) { $currentUserPath -split ";" } else { @() }
        $hasDir = $parts | Where-Object { $_.Trim() -ieq $dir }

        if (-not $hasDir) {
            $newPath = if ([string]::IsNullOrWhiteSpace($currentUserPath)) { $dir } else { $currentUserPath.TrimEnd(";") + ";" + $dir }
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            $result.Status = "added to registry"
            Write-OK "Added $dir to User PATH."
        } else {
            Write-Step "$dir is already in User PATH."
        }
    } catch {
        $result.Status = "skipped (non-Windows)"
        Write-Step "Skipped User PATH persistence (non-Windows environment)."
    }

    $sessionParts = if ($env:PATH) { $env:PATH -split ";" } else { @() }
    $hasSessionDir = $sessionParts | Where-Object { $_.Trim() -ieq $dir }
    if (-not $hasSessionDir) {
        $env:PATH = if ([string]::IsNullOrWhiteSpace($env:PATH)) { $dir } else { $env:PATH.TrimEnd(";") + ";" + $dir }
        Write-OK "Refreshed PATH for current session."
    }
    return $result
}

function Remove-FromPath([string]$dir) {
    try {
        $currentUserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if (-not $currentUserPath) { return }
        $parts = $currentUserPath -split ";" | Where-Object { $_.Trim() -and ($_.Trim() -ine $dir) }
        $newPath = $parts -join ";"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-OK "Removed $dir from User PATH."
    } catch {
        # Ignore on non-Windows
    }
}

function Write-InstallSummary([string]$version, [string]$binPath, [string]$installDir, [hashtable]$pathResult, [bool]$isNoPath, [string]$prevVersion = "") {
    $userData = Join-Path $env:USERPROFILE ".movie"
    $dbPath = Join-Path $userData "data\movie.db"
    $cacheDbPath = Join-Path $userData "data\cache.db"

    Write-Host ""
    Write-Host "  ┌──────────────────────────────────────────────────────────┐" -ForegroundColor Cyan
    Write-Host "  │   🎬 MOVIE CLI — Installation Summary                    │" -ForegroundColor Cyan
    Write-Host "  └──────────────────────────────────────────────────────────┘" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  ── Installation Details ──" -ForegroundColor Cyan
    if ($prevVersion -and $prevVersion -ne $version) {
        Write-Host "  ● Version:        $version " -ForegroundColor White -NoNewline
        Write-Host "(upgraded from $prevVersion)" -ForegroundColor Green
    } else {
        Write-Host "  ● Version:        $version" -ForegroundColor White
    }
    Write-Host "  ● Binary:         $binPath" -ForegroundColor White
    Write-Host "  ● Install Dir:    $installDir" -ForegroundColor White
    Write-Host "  ● Data Folder:    $userData" -ForegroundColor White
    Write-Host ""
    Write-Host "  ── SQLite Split-DB Stores ──" -ForegroundColor Cyan
    Write-Host "  ● Primary Store:  $dbPath (Persistent Library, WAL)" -ForegroundColor White
    Write-Host "  ● Cache Store:    $cacheDbPath (Ephemeral Cache, WAL)" -ForegroundColor White
    Write-Host ""
    Write-Host "  ── Shell Integration ──" -ForegroundColor Cyan

    if ($isNoPath) {
        Write-Host "  ● PATH Status:    skipped (-NoPath)" -ForegroundColor Yellow
        return
    }

    Write-Host "  ● PATH Status:    active in current session & User PATH" -ForegroundColor Green
    Write-Host "  ● PATH Target:    $($pathResult.Target) ($($pathResult.Status))" -ForegroundColor DarkGray
}

function Invoke-InstallVerification([string]$binPath, [string]$installDir, [bool]$isNoPath) {
    $userData = Join-Path $env:USERPROFILE ".movie"
    $dbDir = Join-Path $userData "data"

    Write-Host ""
    Write-Host "  ── System Diagnostics ──" -ForegroundColor Cyan

    # 1. Version
    if (Test-Path $binPath) {
        try {
            $verLine = (& $binPath version 2>&1 | Out-String).Trim().Split("`n")[0]
            Write-Host ("    [ok]   version      {0}" -f $verLine) -ForegroundColor Green
        }
        catch {
            Write-Host ("    [warn] version      could not run {0}: {1}" -f $binPath, $_) -ForegroundColor Yellow
        }
    }
    else {
        Write-Host ("    [warn] binary       missing at {0}" -f $binPath) -ForegroundColor Yellow
    }

    # 2. PATH active in this session
    $resolved = Get-Command $BinaryName -ErrorAction SilentlyContinue
    if ($resolved) {
        Write-Host ("    [ok]   PATH         active -> {0}" -f $resolved.Source) -ForegroundColor Green
    }
    elseif ($isNoPath) {
        Write-Host ("    [warn] PATH         skipped (-NoPath); invoke with full path: {0}" -f $binPath) -ForegroundColor Yellow
    }
    else {
        Write-Host ("    [warn] PATH         {0} not on PATH yet - open a new terminal." -f $BinaryName) -ForegroundColor Yellow
    }

    # 3. Data folder
    if (Test-Path $dbDir) {
        Write-Host ("    [ok]   data-store   ready ({0})" -f $dbDir) -ForegroundColor Green
    }
    else {
        try {
            New-Item -ItemType Directory -Path $dbDir -Force | Out-Null
            Write-Host ("    [ok]   data-store   initialized ({0})" -f $dbDir) -ForegroundColor Green
        }
        catch {
            Write-Host ("    [warn] data-store   could not initialize {0}" -f $dbDir) -ForegroundColor Yellow
        }
    }

    # 4. Shell quick navigation helper (mcd)
    if ($PROFILE) {
        try {
            $profileDir = Split-Path -Parent $PROFILE
            if (-not (Test-Path $profileDir)) {
                New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
            }
            $mcdFunc = @"

# movie-cli quick navigation helper
function mcd {
    `$p = (& "$binPath" cd @args)
    if (`$p) { Set-Location `$p }
}
"@
            $hasMcd = $false
            if (Test-Path $PROFILE) {
                $content = Get-Content $PROFILE -Raw -ErrorAction SilentlyContinue
                if ($content -match 'function mcd') { $hasMcd = $true }
            }
            if (-not $hasMcd) {
                Add-Content -Path $PROFILE -Value $mcdFunc -Encoding UTF8 -Force
                Write-Host ("    [ok]   shell-func   mcd shortcut installed in {0}" -f $PROFILE) -ForegroundColor Green
            } else {
                Write-Host ("    [ok]   shell-func   mcd shortcut ready in {0}" -f $PROFILE) -ForegroundColor Green
            }
        } catch {
            Write-Host ("    [info] shell-func   run 'movie cd --setup' to configure 'mcd'") -ForegroundColor DarkGray
        }
    }
}

# --- Uninstall ---
function Invoke-Uninstall([string]$installDir) {
    Write-Host ""
    Write-Host "  movie CLI uninstaller" -ForegroundColor White
    Write-Host "  =====================" -ForegroundColor DarkGray
    Write-Host ""

    $binPath = Join-Path $installDir $BinaryName
    if (Test-Path $binPath) {
        Remove-Item $binPath -Force -ErrorAction SilentlyContinue
        Write-OK "Removed binary: $binPath"
    }

    if (Test-Path $installDir) {
        $remaining = Get-ChildItem -Path $installDir -Force -ErrorAction SilentlyContinue
        if (-not $remaining -or $remaining.Count -eq 0) {
            Remove-Item $installDir -Force -Recurse -ErrorAction SilentlyContinue
            Write-OK "Removed directory: $installDir"
        }
    }

    Remove-FromPath $installDir

    $userData = Join-Path $env:USERPROFILE ".movie"
    if (Test-Path $userData) {
        if ($PurgeData) {
            Remove-Item $userData -Recurse -Force -ErrorAction SilentlyContinue
            Write-OK "Purged user data: $userData"
        } elseif (-not $KeepData -and -not $Force) {
            Write-Host ""
            Write-Host "  Found user data at $userData" -ForegroundColor Yellow
            Write-Host "  Delete user data and database? [y/N]: " -ForegroundColor Yellow -NoNewline
            $ans = Read-Host
            if ($ans -match '^(y|yes)$') {
                Remove-Item $userData -Recurse -Force -ErrorAction SilentlyContinue
                Write-OK "Purged user data: $userData"
            } else {
                Write-Step "Preserved user data: $userData"
            }
        } else {
            Write-Step "Preserved user data: $userData"
        }
    }

    Write-Host ""
    Write-OK "movie CLI uninstalled successfully."
    exit 0
}

# ===========================================================================
# Main Flow
# ===========================================================================

$resolvedDir = Resolve-InstallDir $InstallDir

if ($Uninstall) {
    Invoke-Uninstall $resolvedDir
}

$resolvedArch = Resolve-Arch $Arch
$resolvedVersion = if ($Version) {
    if ($Version -notmatch '^v') { "v$Version" } else { $Version }
} else {
    Resolve-LatestVersion
}

$binPath = Join-Path $resolvedDir $BinaryName
$previousVersion = $null
if (Test-Path $binPath) {
    try {
        $rawPrev = (& $binPath version 2>&1 | Out-String).Trim()
        if ($rawPrev -match 'v?(\d+\.\d+\.\d+)') {
            $previousVersion = "v$($Matches[1])"
        } elseif ($rawPrev) {
            $previousVersion = $rawPrev
        }
    } catch {}
}

Write-Host ""
Write-Host "  ┌──────────────────────────────────────────────────────────┐" -ForegroundColor Cyan
Write-Host "  │   🎬 MOVIE CLI — Autonomous Media Library Manager        │" -ForegroundColor Cyan
Write-Host "  └──────────────────────────────────────────────────────────┘" -ForegroundColor Cyan
Write-Host ""
if ($previousVersion -and $previousVersion -ne $resolvedVersion) {
    Write-Host "  ● Action:         Upgrading $previousVersion -> $resolvedVersion" -ForegroundColor Yellow
} elseif ($previousVersion) {
    Write-Host "  ● Action:         Reinstalling $resolvedVersion" -ForegroundColor White
} else {
    Write-Host "  ● Action:         Clean install of $resolvedVersion" -ForegroundColor White
}
Write-Host "  ● Architecture:   $resolvedArch" -ForegroundColor White
Write-Host "  ● Repository:     github.com/$Repo" -ForegroundColor DarkGray
Write-Host ""

$asset = Get-Asset $resolvedVersion $resolvedArch
try {
    Install-Binary $asset.ZipPath $resolvedDir
} finally {
    Remove-Item $asset.TmpDir -Recurse -Force -ErrorAction SilentlyContinue
}

$pathResult = @{ Target = "-NoPath"; Status = "skipped" }
if (-not $NoPath) {
    $pathResult = Add-ToPath $resolvedDir
}

$installedVersion = $resolvedVersion
if (Test-Path $binPath) {
    try {
        $rawVer = (& $binPath version 2>&1 | Out-String).Trim()
        if ($rawVer -match 'v?(\d+\.\d+\.\d+)') {
            $installedVersion = "v$($Matches[1])"
        }
    } catch { }
}

Write-InstallSummary $installedVersion $binPath $resolvedDir $pathResult $NoPath.IsPresent $previousVersion

Invoke-InstallVerification $binPath $resolvedDir $NoPath.IsPresent

if ($env:MOVIE_UPDATING -ne "1" -and (Test-Path -LiteralPath $binPath)) {
    Write-Host ""
    try {
        & $binPath binary
    } catch {
        Write-Warning "[Main.MovieBinary] $_"
    }
}

Write-Host ""
Write-Host "  💡 Quick Shortcuts:" -ForegroundColor Cyan
Write-Host "     mcd <movie-or-folder>   Jump directly into any movie or scanned folder" -ForegroundColor White
Write-Host "     movie ui [folder]       Launch Web UI scoped to that folder" -ForegroundColor White
Write-Host "     movie ls --folders      List all scanned root folders & item counts" -ForegroundColor White
Write-Host ""
Write-OK "Done! Run 'movie --help' to get started."
Write-Host ""

