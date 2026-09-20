# Completed Plan: TMDb Invalid API Key Re-Prompt Loop & Error Spam Elimination

- **ID:** 19-invalid-api-key-reprompt-loop-and-spam-suppression
- **Status:** COMPLETED
- **Completed Date:** 2026-09-20
- **Release Version:** v2.330.0
- **Total Self-Loop Steps:** 3 steps

## Overview & Scope
Refined the TMDb authentication and re-prompt system so that when an invalid API key is provided or rejected, the CLI proactively warns the user, prompts again in an interactive loop until valid credentials are provided or skipped, eliminates multi-step prompt friction, and suppresses 200+ repetitive error logs across scan workers.

## Deliverables & Accomplishments
1. **HTTP Status 401 & 403 Classification (`tmdb/http.go`):**
   - Treated both HTTP `401 Unauthorized` and `403 Forbidden` as `ErrAuthInvalid`.
   - Prevented forbidden/rejected API keys from falsely surfacing as general network connectivity errors.
2. **Dual-Case Config Resolution & Streamlined Prompt (`cmd/movie_tmdb.go`):**
   - Updated `readTmdbCredentials()` to resolve both `TmdbApiKey` (PascalCase) and `tmdb_api_key` (snake_case), as well as `TmdbToken` and `tmdb_token`.
   - Updated `saveTmdbCredentialsToDB()` to persist both casing styles to SQLite.
   - Streamlined `promptForValidTmdbCredentials()` into a single clear input: `Enter TMDb API key or Bearer token (or press Enter to skip):`.
   - Automatically detects Bearer tokens (`eyJ...`) versus v3 API keys, trims surrounding quotes and whitespace, verifies immediately against TMDb, and re-prompts on failure.
3. **Repetitive Error Spam Suppression (`cmd/movie_scan_process.go`):**
   - Updated `enrichFromTMDb()`: when authentication fails and is disabled (`!ctx.HasTMDb`), suppressed repetitive `logTMDbSearchError` invocations across the file list.
4. **Unit Tests (`cmd/movie_tmdb_test.go`, `tmdb/client_test.go`):**
   - Added `TestReadTmdbCredentials_FromSnakeCaseConfig` and `TestSaveTmdbCredentialsToDB_SavesBothCases`.
   - Added `TestVerifyAuth_ForbiddenAuth`.

## Verification Outcomes
- 20/20 local CI/CD quality gates passed (100% green).
- Golangci-lint: 0 issues.
- All Go unit tests passed (100%).
