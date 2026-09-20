# Completed Plan: TMDb API Key Verification, Re-Prompt Loop & Minor Release

- **ID:** 17-invalid-api-key-reprompt-and-release
- **Status:** COMPLETED
- **Completed Date:** 2026-09-20
- **Release Version:** v2.328.0
- **Total Self-Loop Steps:** 3 steps

## Overview & Scope
Fixed the issue where an invalid or rejected TMDb API key logged repeated error messages (`❌ TMDb API key is invalid. Run: movie config set tmdb_api_key YOUR_KEY`) for every file in the library (e.g. 208 times) without prompting the user for a valid key.

## Deliverables & Accomplishments

1. **`tmdb.Client.VerifyAuth() error` (`tmdb/client.go`, `tmdb/http.go`):**
   - Implemented proactive credential verification querying TMDb `/configuration`.
   - Returns `nil` on success (HTTP 200).
   - Returns `ErrAuthMissing` if credentials are not configured.
   - Returns `ErrAuthInvalid` if TMDb rejects the key (HTTP 401).
   - Returns network errors / timeouts as-is without falsely classifying them as invalid credentials.
   - Added unit test suite in `tmdb/client_test.go` covering 200 OK, 401 Unauthorized, and missing auth.

2. **Interactive Re-Prompt Loop (`cmd/movie_tmdb.go`):**
   - Refactored `resolveScanTmdbCredentials`:
     - Checks existing DB/env credentials with `VerifyAuth()` before starting scan.
     - Detects invalid keys upfront and notifies user.
     - In interactive terminals (`isatty.IsTerminal`), launches a loop that prompts for new credentials, verifies them immediately, and **prompts again** if the newly entered key is also rejected.
     - Once valid, saves to database and proceeds.
     - If left blank or non-interactive, safely continues without metadata fetching.
   - Implemented `ensureValidTmdbClient` for all TMDb-backed commands (`movie search`, `movie suggest`, `movie discover`, `movie info`).

3. **In-Flight Auth Error Handling & Suppression (`cmd/movie_scan_process.go`):**
   - Implemented `handleAuthFailure()` on `*ScanContext` with mutex protection.
   - If an auth failure occurs mid-scan, safely pauses to prompt the user for a valid key and retries the failed search.
   - In non-interactive mode or if the user leaves it blank, disables further TMDb lookups (`ctx.HasTMDb = false`), completely eliminating repetitive error logs across the remaining files.

4. **Integration with Interactive Commands:**
   - Updated `cmd/movie_search.go`, `cmd/movie_suggest.go`, `cmd/movie_discover.go`, and `cmd/movie_info.go` to use `ensureValidTmdbClient`.

## Verification Outcomes
- Unit tests: `TestVerifyAuth_MissingAuth`, `TestVerifyAuth_ValidAuth`, `TestVerifyAuth_InvalidAuth` passed.
- Command unit tests: `TestReadTmdbCredentials_FromConfig`, `TestEnsureValidTmdbClient_NoAuth` passed.
- Local Quality Gates: 20/20 gates passed 100% green via `python 03-ai-scripts/06-cicd-local-runner.py --all-paths --run-tests`.
- Golangci-lint: 0 issues across all packages.
