# Subtask 02: Multiple TMDb API Key/Token Pool & Automatic Rotation

> **Parent Plan:** `.ai-memory/plans/pending/12-terminal-ui-tmdb-rotation-and-search-enhancements.md`  
> **Status:** Complete  
> **Target Files:**
> - `tmdb/client.go`
> - `tmdb/http.go`
> - `cmd/movie_tmdb.go`

---

## 1. Problem Statement

1. TMDb enforces strict rate limits (HTTP 429) and user tokens can get rate limited or exhausted during large library scans (e.g. 200+ movies).
2. The current client only supports a single `ApiKey` and `AccessToken`.
3. If an API key or token encounters HTTP 429 or auth expiration, the entire scan stalls or skips metadata for remaining movies.
4. Users need support for multiple API keys/tokens configured via config table or environment variables, with automatic round-robin / failover rotation.

---

## 2. Proposed Changes

### A. Token Pool in TMDb Client (`tmdb/client.go`, `tmdb/http.go`)
- Define `Credential` struct in `tmdb`:
  ```go
  type Credential struct {
      ApiKey string
      Token  string
  }
  ```
- Extend `Client` to store a slice of `credentials []Credential`, `activeCredIndex int`, and `credMu sync.Mutex`.
- Add methods:
  - `AddCredential(apiKey, token string)`
  - `SetCredentials(creds []Credential)`
  - `RotateCredential() bool` (switches to next credential in pool; returns true if rotated to another credential).
- In `tmdb/http.go`, when `doGet` encounters HTTP 429 (rate limited) or HTTP 401/403 (auth rejected):
  - If `c.RotateCredential()` returns true, log a brief notice and immediately retry the request with the new key/token.
  - If no further credentials exist, perform normal backoff or error return.

### B. Configuration & Environment Discovery (`cmd/movie_tmdb.go`)
- Support loading multiple keys/tokens from:
  - Config keys: `tmdb_api_keys`, `tmdb_tokens` (comma- or whitespace-separated list).
  - Environment variables: `TMDB_API_KEYS`, `TMDB_TOKENS` (comma-separated list).
  - Indexed environment variables: `TMDB_API_KEY_1`, `TMDB_API_KEY_2`, `TMDB_TOKEN_1`, `TMDB_TOKEN_2`, etc.
  - Legacy `TMDB_API_KEY`, `TMDB_TOKEN`, and DB config `TmdbApiKey`, `TmdbToken`.
- Merge and deduplicate all found credentials into a slice `[]tmdb.Credential`.
- Pass full credentials list to `tmdb.Client`.

---

## 3. Verification Criteria
- Client successfully initializes with 1 or multiple credentials.
- When `RotateCredential()` is invoked, subsequent HTTP requests use the next key/token in the pool.
- HTTP 429 / 401 triggers seamless token rotation without halting the scan.
