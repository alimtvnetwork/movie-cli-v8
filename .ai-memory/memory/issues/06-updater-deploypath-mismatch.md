---
name: Updater deploy-path vs PATH mismatch
description: In update mode, deploy target MUST be the active PATH binary, not powershell.json deployPath. Old fix: v2.121.0.
type: constraint
---
When `run.ps1 -Update` runs, the deploy target is the binary the user is actually running (active PATH `movie`), NOT `powershell.json`'s `deployPath`. If these point to different drives/dirs, the active binary stays frozen forever and every subsequent `movie update` re-spawns the same old handoff worker.

**Why:** powershell.json deployPath is a developer-build convenience. Update mode must always replace the binary the user invoked, otherwise the update is invisible.

**How to apply:**
- `Resolve-DeployTarget` falls back to `Get-Command movie` when `-Update` and no `-TargetBinaryPath`.
- Post-deploy PATH-sync retry loop is unconditionally skipped in update mode (it only ever caused "Access is denied" spam against the still-running parent).
- `updater/cleanup.go` silently skips locked `*-update-*` files and routes warnings to stdout (stderr triggers PowerShell NativeCommandError).
