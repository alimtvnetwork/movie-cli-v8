# CI/CD Issue: Release Script Version-Pinning Variable Name Mismatch

- Job: Release / Enforce version-pinning contract on install scripts
- Run ID: 35487967428
- Type: FAIL
- Detected: 2026-09-20T04:00:54Z
- Status: resolved

## 1. Why It Happened

In `.github/workflows/release.yml`, the step `Enforce version-pinning contract on install scripts` enforces that the generated `dist/install.sh` has the version-pinning declaration matching:
```bash
if ! grep -q "VERSION_PINNED=\"$VERSION\"" "$script"; then
  echo "::error file=$script::VERSION_PINNED != $VERSION"
  FAIL=1
fi
```
However, the step `Generate version-specific install.sh` wrote:
```bash
PINNED_VERSION="VERSION_PLACEHOLDER"
```
instead of `VERSION_PINNED="VERSION_PLACEHOLDER"`.
When substituted with `$VERSION`, the file contained `PINNED_VERSION="v2.326.0"`, causing `grep -q "VERSION_PINNED=\"$VERSION\""` to fail and exit 1:
```text
##[error]VERSION_PINNED != v2.326.0
##[error]Version-pinning contract violated — refusing to publish release.
```

## 2. How It Slipped Through

Local testing ran the installer smoke tester in `dryrun` mode which validated that `-DryRun` / `--dry-run` returned `dryrun.status=ok`. However, the CI release workflow contains a dedicated contract assertion step (`Enforce version-pinning contract on install scripts`) that specifically greps for the exact token `VERSION_PINNED="<version>"`.

## 3. Root Cause

Variable name disparity between the generator (`PINNED_VERSION`) and the release validator (`VERSION_PINNED`), violating the naming contract established in the workflow validation logic.

## 4. Code Fix Applied

1. Updated `.github/workflows/release.yml` in `Generate version-specific install.sh` to define `VERSION_PINNED="VERSION_PLACEHOLDER"` and alias `PINNED_VERSION="$VERSION_PINNED"`.
2. Updated local validation in `03-ai-scripts/06-cicd-local-runner.py` / `.github/scripts/smoke-installer.py` to assert the `VERSION_PINNED` contract token directly.
