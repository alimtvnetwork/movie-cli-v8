<#
.SYNOPSIS
    Quick interactive installer for movie CLI on Windows.

.DESCRIPTION
    Prompts for an install folder (default: $env:LOCALAPPDATA\movie-cli),
    then delegates to the canonical install.ps1 with the target path.

    Versioned repo discovery: if the source repo URL ends with -v<N>, this
    script probes for higher-numbered sibling repos (-v<N+1>, -v<N+2>, ...)
    and delegates to the latest available one.

    Run via one-liner:
      irm https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install-quick.ps1 | iex

    Or locally:
      ./install-quick.ps1
      ./install-quick.ps1 -InstallDir "$env:LOCALAPPDATA\movie-cli"
      ./install-quick.ps1 -NoDiscovery
#>

param(
    [string]$InstallDir = "",
    [string]$Version = "",
    [switch]$Interactive,
    [switch]$NoDiscovery,
    [string]$LogFile = "",
    [int]$DiscoveryWindow = 20
)

$ErrorActionPreference = "Stop"
$ProgressPreference    = "SilentlyContinue"

$Repo         = "alimtvnetwork/movie-cli-v8"
$InstallerUrl = "https://raw.githubusercontent.com/$Repo/main/install.ps1"
$baseAppDir   = $env:LOCALAPPDATA
if (-not $baseAppDir) {
    $baseAppDir = if ($env:HOME) { Join-Path $env:HOME ".local" } else { "." }
}
$DefaultDir   = Join-Path $baseAppDir "movie-cli"

if ([string]::IsNullOrWhiteSpace($LogFile)) {
    $stamp    = (Get-Date).ToString("yyyyMMdd-HHmmss")
    $baseTemp = if ($env:TEMP) { $env:TEMP } elseif ($env:TMP) { $env:TMP } else { [System.IO.Path]::GetTempPath() }
    $LogFile  = Join-Path $baseTemp "movie-install-quick-$stamp.log"
}

$script:InstallErrors = New-Object System.Collections.Generic.List[string]

function Write-Log([string]$message, [string]$level = "INFO") {
    $line = "[{0}] [{1}] {2}" -f (Get-Date -Format "yyyy-MM-dd HH:mm:ss"), $level, $message
    try { Add-Content -Path $LogFile -Value $line -Encoding UTF8 -ErrorAction SilentlyContinue } catch { }
}

function Invoke-Safe {
    param(
        [Parameter(Mandatory)][string]$Step,
        [Parameter(Mandatory)][scriptblock]$Action,
        [switch]$Fatal
    )
    Write-Log "BEGIN: $Step"
    try {
        $result = & $Action
        Write-Log "OK:    $Step"
        return $result
    } catch {
        $msg = "FAIL:  $Step :: $($_.Exception.Message)"
        Write-Log $msg "ERROR"
        $script:InstallErrors.Add("$Step -> $($_.Exception.Message)")
        Write-Host "  [ERROR] $Step : $($_.Exception.Message)" -ForegroundColor Red
        if ($Fatal) { throw }
        return $null
    }
}

function Split-RepoSuffix([string]$targetRepo) {
    if ($targetRepo -match '^([^/]+)/(.+)-v(\d+)$') {
        return @{
            Owner = $Matches[1]
            Stem  = $Matches[2]
            N     = [int]$Matches[3]
        }
    }
    return $null
}

function Resolve-EffectiveRepo([string]$targetRepo, [int]$window) {
    $parts = Split-RepoSuffix $targetRepo
    if ($null -eq $parts) {
        return $targetRepo
    }

    $owner    = $parts.Owner
    $stem     = $parts.Stem
    $baseline = $parts.N

    $maxConcurrency = 20
    if (-not [string]::IsNullOrWhiteSpace($env:GITHUB_TOKEN)) {
        $maxConcurrency = 50
    }
    if ($window -gt $maxConcurrency) { $window = $maxConcurrency }

    $candidates = @()
    for ($m = $baseline + 1; $m -le ($baseline + $window); $m++) {
        $candidates += [pscustomobject]@{
            M   = $m
            Url = "https://github.com/$owner/$stem-v$m"
        }
    }

    $pool = [runspacefactory]::CreateRunspacePool(1, $window)
    $pool.Open()

    $jobs = @()
    foreach ($c in $candidates) {
        $ps = [powershell]::Create().AddScript({
            param($url, $m)
            try {
                $resp = Invoke-WebRequest -Uri $url -Method Head -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
                if ($resp.StatusCode -eq 200) {
                    [pscustomobject]@{ M = $m; Url = $url; Hit = $true }
                } else {
                    [pscustomobject]@{ M = $m; Url = $url; Hit = $false }
                }
            } catch {
                [pscustomobject]@{ M = $m; Url = $url; Hit = $false }
            }
        }).AddArgument($c.Url).AddArgument($c.M)
        $ps.RunspacePool = $pool
        $jobs += [pscustomobject]@{ PS = $ps; Handle = $ps.BeginInvoke(); M = $c.M; Url = $c.Url }
    }

    $effective = $baseline
    foreach ($j in $jobs) {
        try {
            $result = $j.PS.EndInvoke($j.Handle)
            $r = $result | Select-Object -First 1
            if ($r -and $r.Hit) {
                if ($r.M -gt $effective) { $effective = $r.M }
            }
        } catch {
        } finally {
            $j.PS.Dispose()
        }
    }

    $pool.Close()
    $pool.Dispose()

    if ($effective -eq $baseline) {
        return $targetRepo
    }

    Write-Host "  [discovery] effective repo: $owner/$stem-v$effective (was -v$baseline)" -ForegroundColor Cyan
    return "$owner/$stem-v$effective"
}

function Invoke-DelegatedInstaller([string]$effectiveRepo, [string]$targetDir, [string]$ver) {
    $delegatedUrl = "https://raw.githubusercontent.com/$effectiveRepo/main/install-quick.ps1"
    Write-Host "  [discovery] delegating to $delegatedUrl" -ForegroundColor Cyan

    $env:INSTALLER_DELEGATED = "1"
    try {
        $script = (Invoke-WebRequest -Uri $delegatedUrl -UseBasicParsing -TimeoutSec 15).Content
    } catch {
        Remove-Item Env:INSTALLER_DELEGATED -ErrorAction SilentlyContinue
        return $false
    }

    $block = [ScriptBlock]::Create($script)
    $passArgs = @{}
    if (-not [string]::IsNullOrWhiteSpace($targetDir)) { $passArgs.InstallDir = $targetDir }
    if (-not [string]::IsNullOrWhiteSpace($ver))       { $passArgs.Version    = $ver }

    & $block @passArgs
    return $true
}

$alreadyDelegated = ($env:INSTALLER_DELEGATED -eq "1")
if (-not $alreadyDelegated -and -not $NoDiscovery -and [string]::IsNullOrWhiteSpace($Version)) {
    $effective = Resolve-EffectiveRepo $Repo $DiscoveryWindow
    if ($effective -ne $Repo) {
        $delegated = Invoke-DelegatedInstaller $effective $InstallDir $Version
        if ($delegated) { return }
    }
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

function Save-DeployPath([string]$dir) {
    try {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
        $cfgPath = Join-Path $dir "powershell.json"
        $cfg = [ordered]@{
            deployPath  = $dir
            buildOutput = "./bin"
            binaryName  = "movie.exe"
            goSource    = "."
            copyData    = $true
        }
        ($cfg | ConvertTo-Json) | Set-Content -Path $cfgPath -Encoding UTF8
        Write-Log "Saved deployPath -> $cfgPath"
    } catch {
        Write-Log "Could not save powershell.json: $_" "WARN"
    }
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

Invoke-Safe "Save deploy path" { Save-DeployPath $InstallDir }

$localInstaller = Join-Path $PSScriptRoot "install.ps1"
$scriptContent = $null

if (Test-Path $localInstaller) {
    $scriptContent = Get-Content $localInstaller -Raw -Encoding UTF8
} else {
    Write-Log "Downloading canonical installer: $InstallerUrl"
    $scriptContent = Invoke-Safe "Download installer" { (Invoke-WebRequest -Uri $InstallerUrl -UseBasicParsing).Content } -Fatal
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

if ($script:InstallErrors.Count -gt 0) {
    Write-Host "  [SUMMARY] Completed with $($script:InstallErrors.Count) warning(s)." -ForegroundColor Yellow
}
