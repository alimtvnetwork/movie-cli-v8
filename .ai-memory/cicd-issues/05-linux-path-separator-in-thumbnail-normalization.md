# CI/CD Issue: Linux Path Separator in Thumbnail Path Normalization

- Job: CI / Test Summary / Summarize test results
- Run ID: 35518272640
- Type: FAIL
- Detected: 2026-09-20T15:01:33Z
- Status: resolved

## 1. Why It Happened (Symptom)

In CI workflow run `#35518272640` on `ubuntu-latest`, the unit test suite failed:
```text
❌ unit: 1 failed, 68 passed
--- FAIL: TestNormalizeExistingThumb (0.00s)
    movie_rest_thumb_test.go:30: normalizeExistingThumb("C:\\Users\\AppData\\Local\\thumbnails\\red-eye.jpg") = "thumbnails/C:\\Users\\AppData\\Local\\thumbnails\\red-eye.jpg", want "thumbnails/red-eye.jpg"
##[error]Some test suites failed
##[error]Process completed with exit code 1.
```

## 2. How It Slipped Through

On Windows development environments, `filepath.Separator` is `\` and `filepath.Base` strips Windows backslash paths as expected. However, on Linux runners, `filepath.Separator` is `/` and backslash is treated as a regular filename character, causing `filepath.Base("C:\\Users\\...\\red-eye.jpg")` to return the full unstripped Windows path.

## 3. Root Cause

`normalizeExistingThumb` relied on platform-dependent `filepath.Base(path)` without first normalizing Windows backslashes (`\`) to forward slashes (`/`), producing inconsistent path segmentation on non-Windows operating systems.

## 4. Code Fix Applied

1. Updated `normalizeExistingThumb` in `cmd/movie_rest_thumb.go` to convert all backslashes to forward slashes via `strings.ReplaceAll(path, "\\", "/")`.
2. Extracted the basename using cross-platform `strings.LastIndex(clean, "/")` rather than OS-specific `filepath.Base`.
3. Added detection for `thumbnails/` substring and single-slash TMDb paths (`strings.Count(clean, "/") == 1`) to distinguish remote TMDb slugs from absolute local filesystem paths.
4. Expanded `TestNormalizeExistingThumb` in `cmd/movie_rest_thumb_test.go` with Linux paths, Windows paths, mixed-slash paths, and bare filenames.
