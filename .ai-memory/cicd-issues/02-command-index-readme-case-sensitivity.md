# CI/CD Issue: Case Sensitivity on Linux Runners for Command Index Check

- **Job:** CI (`#35486041660`)
- **Step:** Lint / Command-index drift guard
- **Type:** FAIL
- **Detected:** 2026-09-20 03:14:06 UTC
- **Status:** resolved

## Error
```text
FileNotFoundError: [Errno 2] No such file or directory: '/home/runner/work/movie-cli-v8/movie-cli-v8/README.md'
##[error]Command index is stale. Run: python3 scripts/gen-command-index.py
##[error]Process completed with exit code 1.
```

## Root Cause
1. `scripts/gen-command-index.py` had hardcoded `README = REPO_ROOT / "README.md"`.
2. On Windows hosts (case-insensitive NTFS), resolving `"README.md"` silently succeeded against `readme.md`.
3. On Linux GitHub Actions runners (case-sensitive ext4), `README.md` threw `FileNotFoundError` because the tracked file in git is `readme.md`.
4. Similarly, `scripts/check-quickstart.sh` and `scripts/sync-install-from-readme.sh` targeted `README.md` directly.

## Fix Applied
1. Updated `scripts/gen-command-index.py` to dynamically fallback:
   `README = (REPO_ROOT / "readme.md") if (REPO_ROOT / "readme.md").exists() else (REPO_ROOT / "README.md")`
2. Updated `scripts/check-quickstart.sh` and `scripts/sync-install-from-readme.sh` to check for `readme.md` before `README.md`.
3. Added `readme.md` to `TARGETS` in `.github/workflows/ci.yml` and `SURFACES` in `.github/workflows/release.yml`.
4. Formatted `version/info.go` to satisfy `gofmt` in `golangci-lint`.
5. Validated locally via `python 03-ai-scripts/06-cicd-local-runner.py --all-paths --run-tests` (18/18 gates pass) and verified on GitHub Actions CI run `#35486260710` (100% green).
