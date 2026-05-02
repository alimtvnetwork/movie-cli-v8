package updater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

// executeUpdateWindows writes a temp PowerShell script and runs it.
func executeUpdateWindows(repoPath, targetBinary string) error {
	scriptPath, err := writeUpdateScript(repoPath, targetBinary)
	if err != nil {
		return apperror.Wrap("cannot write update script", err)
	}
	defer os.Remove(scriptPath)

	return runPowerShellScript(scriptPath)
}

// executeUpdateUnix runs the update via pwsh.
func executeUpdateUnix(repoPath, targetBinary string) error {
	if !hasPwshWithRunPS1(repoPath) {
		runPS1 := filepath.Join(repoPath, "run.ps1")
		return apperror.New("pwsh is required to run %s", runPS1)
	}
	scriptPath, err := writeUpdateScript(repoPath, targetBinary)
	if err != nil {
		return apperror.Wrap("cannot write update script", err)
	}
	defer os.Remove(scriptPath)
	return runPowerShellScript(scriptPath)
}

func hasPwshWithRunPS1(repoPath string) bool {
	if _, err := exec.LookPath("pwsh"); err != nil {
		return false
	}
	runPS1 := filepath.Join(repoPath, "run.ps1")
	_, statErr := os.Stat(runPS1)
	return statErr == nil
}

// writeUpdateScript generates a temp PowerShell script for the update.
func writeUpdateScript(repoPath, targetBinary string) (string, error) {
	script := buildUpdateScriptContent(repoPath, targetBinary, currentBinaryPath())

	tmpFile, err := os.CreateTemp(os.TempDir(), "movie-update-script-*.ps1")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// UTF-8 BOM for PowerShell compatibility
	bom := []byte{0xEF, 0xBB, 0xBF}
	if _, err := tmpFile.Write(bom); err != nil {
		return "", err
	}
	if _, err := tmpFile.WriteString(script); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func currentBinaryPath() string {
	binaryPath, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, evalErr := filepath.EvalSymlinks(binaryPath); evalErr == nil {
		return resolved
	}
	return binaryPath
}

// buildUpdateScriptContent generates the PowerShell script content.
func buildUpdateScriptContent(repoPath, targetBinary, workerBinary string) string {
	repoPath = powerShellString(repoPath)
	targetBinary = powerShellString(targetBinary)
	workerBinary = powerShellString(workerBinary)

	return fmt.Sprintf(`$ErrorActionPreference = "Stop"
$repoPath     = "%s"
$targetBinary = "%s"
$workerBinary = "%s"

try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $OutputEncoding = [System.Text.Encoding]::UTF8
} catch {
}

# Indent every Write-Host line with this prefix so update output sits a
# little further in than run.ps1's own output and is easy to scan.
$P = "    "

function To-ConsoleSafe {
    param([string]$Text)
    if ($null -eq $Text) { return "" }
    $Text = $Text.Replace([string][char]0x2014, " - ")
    $Text = $Text.Replace([string][char]0x2013, "-")
    $Text = $Text.Replace([string][char]0x2192, "->")
    $Text = $Text.Replace([string][char]0x2018, "'")
    $Text = $Text.Replace([string][char]0x2019, "'")
    $Text = $Text.Replace([string][char]0x201C, '"')
    $Text = $Text.Replace([string][char]0x201D, '"')
    return $Text
}

function Say     { param($msg, $color = "Gray")  Write-Host ($P + (To-ConsoleSafe $msg)) -ForegroundColor $color }
function SayOk   { param($msg) Write-Host ($P + "[OK] " + (To-ConsoleSafe $msg)) -ForegroundColor Green }
function SayWarn { param($msg) Write-Host ($P + "[WARN] " + (To-ConsoleSafe $msg)) -ForegroundColor Yellow }
function SayErr  { param($msg) Write-Host ($P + "[ERR] " + (To-ConsoleSafe $msg)) -ForegroundColor Red }

function Resolve-VersionBinary {
    if ($targetBinary -and (Test-Path $targetBinary)) {
        return $targetBinary
    }
    return $null
}

function Schedule-WorkerSelfDelete {
    if (-not $workerBinary) { return }
    if (-not (Test-Path $workerBinary)) { return }
    # Spawn a hidden cmd.exe that waits ~2 s, then deletes the worker copy.
    # ping is the most portable "sleep" on a bare Windows shell.
    $cmdLine = 'ping 127.0.0.1 -n 3 > nul & del /f /q "' + $workerBinary + '"'
    Start-Process -FilePath "cmd.exe" -ArgumentList "/c", $cmdLine -WindowStyle Hidden | Out-Null
}

# Capture current version
$versionBinary = Resolve-VersionBinary
$oldVersion = "unknown"
if ($versionBinary -and (Test-Path $versionBinary)) {
    $oldVersion = (& $versionBinary version 2>&1) -join " "
}
Say "Version before: $oldVersion"

# Wait for the parent process to fully exit and release its file lock on
# $targetBinary before we ask run.ps1 to overwrite it.
Start-Sleep -Seconds 1.2

# Build and deploy from repo root via run.ps1
$runScript = Join-Path $repoPath "run.ps1"
if (-not (Test-Path $runScript)) {
    SayErr "run.ps1 not found at $runScript"
    exit 1
}

Say "Running update via $runScript" "Cyan"
$runExit = 0
if ($targetBinary) {
    Say "Deploy target: $targetBinary"
    & $runScript -Update -TargetBinaryPath $targetBinary
    $runExit = $LASTEXITCODE
} else {
    & $runScript -Update
    $runExit = $LASTEXITCODE
}

if ($runExit -ne 0) {
    SayErr "run.ps1 exited with code $runExit"
    Schedule-WorkerSelfDelete
    exit $runExit
}

# Compare versions from the original target binary
$versionBinary = Resolve-VersionBinary
$newVersion = "unknown"
if ($versionBinary -and (Test-Path $versionBinary)) {
    $newVersion = (& $versionBinary version 2>&1) -join " "
}

Write-Host ""
if ($oldVersion -eq $newVersion) {
    SayWarn "Version unchanged after update - was version/info.go bumped?"
} else {
    SayOk "Updated: $oldVersion -> $newVersion"
}

# Show changelog from the updated target binary
if ($versionBinary -and (Test-Path $versionBinary)) {
    Write-Host ""
    Say "Latest changelog:" "Cyan"
    $clOutput = & $versionBinary changelog --latest 2>&1
    foreach ($cl in $clOutput) { Write-Host ($P + "  " + (To-ConsoleSafe "$cl")) }
}

# Belt-and-braces sweeper for any older worker copies. The current worker
# is preserved via --skip-path; the detached self-deleter below handles it.
if ($versionBinary -and (Test-Path $versionBinary)) {
    $cleanupArgs = @("update-cleanup")
    if ($workerBinary) { $cleanupArgs += @("--skip-path", $workerBinary) }
    $prevPref = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    $null = & $versionBinary @cleanupArgs *> $null
    $ErrorActionPreference = $prevPref
}

# Top-and-tail banner
Write-Host ""
Write-Host ($P + "+======================================+") -ForegroundColor Cyan
Write-Host ($P + "|  [OK] Update complete               |") -ForegroundColor Cyan
Write-Host ($P + "+======================================+") -ForegroundColor Cyan
Write-Host ""


Schedule-WorkerSelfDelete

# Give the user a beat to see the result before the new console window closes.
Start-Sleep -Seconds 2
`, repoPath, targetBinary, workerBinary)
}

func powerShellString(value string) string {
	value = strings.ReplaceAll(value, "`", "``")
	value = strings.ReplaceAll(value, "\"", "`\"")
	return value
}

// runPowerShellScript executes a PowerShell script with output piped to terminal.
func runPowerShellScript(scriptPath string) error {
	psBin := "powershell"
	if runtime.GOOS != "windows" {
		psBin = "pwsh"
	}

	cmd := exec.Command(psBin, "-ExecutionPolicy", "Bypass", "-NoProfile", "-NoLogo", "-File", scriptPath)
	return runAttached(cmd)
}
