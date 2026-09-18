# Zero-Storage GitHub Actions Mandate

> **Scope:** CI/CD Workflows (`.github/workflows/*.yml`)  
> **Source:** `AGENTS.md` Section 7 & `02-spec/17-consolidated-guidelines/34-compiled-simple-coding-guidelines.md`

---

## 1. Total Ban on `actions/upload-artifact` in Routine CI

- Routine GitHub Actions workflows (`ci.yml`, test suites, linting, matrix runs) MUST NOT upload build artifacts, test results, coverage files, drift reports, or diagnostic logs.
- Free-tier accounts have a strict 0.5 GB quota across all account repositories; uploading artifacts quickly exhausts this quota and locks all repository actions.

## 2. Zero-Storage Diagnostic Reporting

- All test reports, drift summaries, and lint outputs MUST be written directly to `$GITHUB_STEP_SUMMARY` or streamed to console standard output.
- Diagnostic failures MUST use GitHub Actions annotations (`::error::` / `::warning::`).
- Pull request summaries MUST use sticky PR comments.

## 3. Release Assets Exemption

- Binaries and archives attached directly to GitHub Releases via `gh release create` or `gh release upload` are exempt from this ban because GitHub Release assets do not consume the Actions artifact storage quota.
