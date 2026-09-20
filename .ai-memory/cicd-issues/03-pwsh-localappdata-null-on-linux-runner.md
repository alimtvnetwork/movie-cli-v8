# CI/CD Issue: PowerShell $env:LOCALAPPDATA Null on Linux Runner During Release Dry-Run

- Job: Release / Validate generated install scripts contract via dry-run
- Run ID: 35487169928
- Type: FAIL
- Detected: 2026-09-20T03:42:17Z
- Status: resolved

## 1. Why It Happened

In `.github/workflows/release.yml`, the step `Validate generated install scripts contract via dry-run` executes:
```bash
pwsh -ExecutionPolicy Bypass -File dist/install.ps1 -DryRun
```
Inside `dist/install.ps1`, the function `Resolve-InstallDir` directly invoked:
```powershell
$legacy = Join-Path $env:LOCALAPPDATA "movie"
```
On Windows, `$env:LOCALAPPDATA` expands to `C:\Users\<user>\AppData\Local`. However, on GitHub Actions Linux runners (`ubuntu-latest`), Windows-specific environment variables like `LOCALAPPDATA` are not set and evaluate to `$null`. PowerShell's `Join-Path` throws a terminating error when the `Path` parameter is `$null`:
```text
Join-Path: Cannot bind argument to parameter 'Path' because it is null.
```

## 2. How It Slipped Through

Local testing was executed on a Windows workstation where `$env:LOCALAPPDATA` is always populated. The dry-run validation was added to the Linux-hosted release workflow without cross-platform fallback guards for Windows environment variables.

## 3. Root Cause

Unchecked dereferencing of `$env:LOCALAPPDATA` and `$env:TEMP` in cross-platform PowerShell (`pwsh`) without falling back to `$env:HOME` or `/tmp`.

## 4. Code Fix Applied

1. Guarded `$env:LOCALAPPDATA` and `$env:TEMP` across all PowerShell install and uninstall scripts:
   ```powershell
   $base = $env:LOCALAPPDATA
   if (-not $base) {
       $base = if ($env:HOME) { Join-Path $env:HOME ".local" } else { "." }
   }
   $legacy = Join-Path $base "movie"
   ```
2. Applied fallback resolution across:
   - `dist/install.ps1` generation in `.github/workflows/release.yml`
   - `install.ps1`
   - `install-quick.ps1`
   - `uninstall-quick.ps1`
3. In `.github/workflows/release.yml`, ensured environment variables `LOCALAPPDATA` and `TEMP` are provided with cross-platform defaults during dry-run validation:
   ```yaml
   env:
     LOCALAPPDATA: /home/runner/.local
     TEMP: /tmp
   ```
