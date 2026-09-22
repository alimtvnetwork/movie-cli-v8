# Subtask 03: Multi-Engine Search Fallbacks (Relaxed TMDb, DuckDuckGo & Google Search)

> **Parent Plan:** `.ai-memory/plans/pending/12-terminal-ui-tmdb-rotation-and-search-enhancements.md`  
> **Status:** Complete  
> **Target Files:**
> - `tmdb/fallback.go`
> - `tmdb/client.go`
> - `cmd/movie_scan_process.go`

---

## 1. Problem Statement

1. During library scans, several well-known movies encounter search misses:
   - `IF (2024)`: TMDb multi-search query was formatted as `"IF 2024"`, causing TMDb to search for the literal string "IF 2024" rather than matching title "IF" with release year 2024.
   - Short/single-word titles (e.g. `Mud (2012)`, `IF (2024)`): `tryProgressiveTrim` requires `len(words) >= 3` and completely skips 1- or 2-word titles.
   - Release year discrepancies: festival release year vs theatrical release year often differ by ±1 year (e.g. 2025 vs 2026).
   - DuckDuckGo fallback uses outdated bot user-agent string and can get blocked or rate limited.
   - No Google Search fallback exists when DuckDuckGo fails or returns no IMDb ID.

---

## 2. Proposed Changes

### A. Relaxed TMDb Query Matching (`tmdb/fallback.go`)
- Tier 1: Search TMDb with clean title (WITHOUT appending year to text query string). If year is provided:
  - Match candidate where `abs(candidateYear - year) <= 1`.
- Tier 2: Search TMDb movie/tv endpoints directly with `year` and `primary_release_year` parameter.
- Tier 3: Search without year restriction if no candidates match with year.
- Tier 4: Progressive trimming for multi-word titles (`len(words) >= 2`).

### B. Robust DuckDuckGo Scraper (`tmdb/fallback.go`)
- Modern Desktop Browser User-Agent header:
  `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36`
- Scrape `https://html.duckduckgo.com/html/?q=` and fallback to `https://lite.duckduckgo.com/lite/`.
- Regex extraction for `tt\d{7,10}`.

### C. Google Search Fallback Scraper (`tmdb/fallback.go`)
- When DuckDuckGo returns empty, execute secondary search against Google Search:
  `https://www.google.com/search?q={url-encoded-query}+imdb`
- Include desktop headers (User-Agent, Accept-Language).
- Parse HTML response body for IMDb ID pattern `tt\d{7,10}`.

### D. Persistent Caching & TMDb Resolution (`tmdb/fallback.go`)
- If IMDb ID is resolved from either DuckDuckGo or Google:
  - Query TMDb `/find/{imdb_id}?external_source=imdb_id`.
  - Store resolved TMDb ID and media type in `ImdbLookupCache` table.

---

## 3. Verification Criteria
- Searches for "IF" (2024), "Mud" (2012), "Lucky Day" (2019) successfully resolve without returning empty.
- Google search scraper acts as robust fallback when DuckDuckGo yields no match.
- All resolved IMDb IDs are cached for zero-overhead repeat scans.
