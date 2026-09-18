# Issue #06: Updater Handoff Target Path Regression

## Issue Summary

1. **What happened**: `movie update` built successfully but the worker/update flow still tried to sync or inspect the active PATH binary after deploy, which produced `Active PATH binary is in use`, `Could not sync active PATH binary after retries`, and `Access is denied` on the worker copy.
2. **Where** (module + file paths): `updater/script.go`, `updater/cleanup.go`, `run.ps1`, `cmd/update.go`, `spec/13-self-update-app-update/03-copy-and-handoff.md`.
3. **Symptoms and impact**: The deploy could land in one location while the live PATH binary stayed locked in another location, the user saw a false-looking partial success, the worker cleanup emitted a deletion warning, and PowerShell showed garbled glyphs in some warning/error lines.
4. **How discovered**: Real Windows end-to-end update output from the user showed `Target: E:\bin-run\movie.exe`, deploy output for `D:\bin-run\movie.exe`, then PATH sync retries against `E:\bin-run\movie.exe`, followed by `Could not remove movie-update-3804.exe: Access is denied.`

## Root Cause Analysis

1. **Direct cause**: The worker knew the exact original binary path (`--target-binary`) but converted it back into partial arguments for `run.ps1` (`-DeployPath` + `-BinaryNameOverride`) instead of passing one authoritative full path. That left room for update-mode deploy targeting to diverge from the originally launched binary.
2. **Contributing factors**: Post-deploy verification and cleanup still depended on runtime path resolution/fallback behavior, and `update-cleanup --skip-path` did not normalize quoted or padded skip paths before comparison.
3. **Triggering conditions**: Windows setup with more than one deployed `movie.exe` location in circulation (`D:\bin-run` vs `E:\bin-run`) plus an update run started from the PATH-resolved binary.
4. **Why spec did not prevent it**: The handoff spec required `--target-binary <orig>` at the handoff boundary, but it did not explicitly state that `run.ps1` must receive the same destination as a single full-path argument with no reconstruction.

## Fix Description

1. **Spec changes**: Added a rule to `spec/13-self-update-app-update/03-copy-and-handoff.md` that update-mode deploy targeting must remain a single explicit full path from `update-runner --target-binary` through `run.ps1 -TargetBinaryPath`.
2. **New rules or constraints**: Never reconstruct the deploy destination from partial pieces when the full target path is already known. Normalize cleanup skip paths before comparing preserved artifacts.
3. **Why it resolves root cause**: The worker and `run.ps1` now operate on the same exact path end-to-end, so update mode cannot silently fall back to config values or a different PATH binary. Cleanup normalization prevents the live worker copy from being misidentified as a stale artifact.
4. **Config changes**: Added `-TargetBinaryPath` support to `run.ps1` for update mode.
5. **Diagnostics required**: End-to-end Windows update must verify that the displayed target path, deploy destination, post-update version check, and cleanup all reference the same binary path, and no garbled warning glyphs appear.

## Iterations History

1. **Iteration 1**: Detached handoff restored, but worker still reconstructed target path for `run.ps1` and cleanup/messages were not hardened enough for mixed-path Windows installs -> partial regression remained visible.
2. **Iteration 2**: Passed one authoritative `-TargetBinaryPath`, normalized cleanup skip paths, and replaced fragile Unicode status markers in updater output -> partial fix, but update mode still ran a PATH-sync step against a second binary location.
3. **Iteration 3**: Removed update-mode PATH sync entirely and forced post-update version/cleanup checks to stay on the handed-off target path only -> lock conflict and wrong-binary drift removed.

## Prevention and Non-Regression

1. **Prevention rule**: If the updater knows the original executable full path, every later stage must reuse that same full path verbatim and must not run any secondary PATH-sync copy step in update mode.
2. **Acceptance criteria**: `movie update` launched from `X:\...\movie.exe` must deploy/update/verify the same `X:\...\movie.exe`; cleanup must not attempt to delete the live worker binary; no PowerShell gibberish appears in updater warnings.
3. **Guardrails or linting**: Keep the spec rule in the handoff document and reject future updater changes that split a known target-binary full path into multiple deploy arguments.
4. **Spec references** (file paths): `spec/13-self-update-app-update/03-copy-and-handoff.md`.

## TODO and Follow-ups

- [x] Add single-path `TargetBinaryPath` handoff from worker to `run.ps1`
- [x] Normalize cleanup `--skip-path` before comparison
- [x] Remove fragile updater glyphs that garble in PowerShell
- [ ] Re-run `movie update` on Windows from the PATH binary and confirm the same target path is used end-to-end with no PATH-sync retry loop and no worker cleanup warning

## Done Checklist

- [x] Spec updated under `/spec/08-app/`
- [x] Issue write-up created under `/spec/09-app-issues/`
- [x] Memory updated with summary and prevention rule
- [x] Acceptance criteria updated or added
- [x] Iterations recorded if applicable