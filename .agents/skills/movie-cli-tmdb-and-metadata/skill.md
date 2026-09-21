---
name: movie-cli-tmdb-and-metadata
description: "TMDb API integration, rate limiting, IMDb fallback chains, lookup caching, and media metadata hydration in movie-cli-v8."
---

# Movie CLI TMDb and Metadata Skill

## Overview

The metadata subsystem of `movie-cli-v8` interfaces with The Movie Database (TMDb) v3 REST API to fetch media details, release dates, ratings, cast, crew, genres, collections, poster art, and backdrops. It includes a fallback scraper and persistent SQLite lookup cache for hard-to-match titles.

## Core Files and Architecture (`tmdb/`)

- `tmdb/client.go`: `Client` struct, credentials handling, authentication checks, image downloading.
- `tmdb/http.go`: Generic HTTP request dispatch, error parsing, JSON decoding, status code handling.
- `tmdb/limiter.go`: Rate-limiting logic preventing API throttling and handling HTTP 429 retries.
- `tmdb/fallback.go`: DuckDuckGo-to-IMDb scraper fallback for titles missed by direct TMDb queries.
- `tmdb/omdb.go`: Optional OMDb API fallback bridge.
- `tmdb/types.go`: Struct definitions for movies, TV series, search responses, credits, and images.
- `db/imdb_lookup_cache.go`: SQLite-backed lookup cache table storing resolved titles and IMDb IDs.

## API Configuration & Authentication

### Base Endpoints
- **API Base:** `https://api.themoviedb.org/3`
- **Image CDN Base:** `https://image.tmdb.org/t/p/w500`

### Authentication Sources (In Order of Precedence)
1. In-memory `ApiKey` or `AccessToken` passed to `NewClientWithToken(key, token)`
2. Database configuration: `Config` table key `tmdb_api_key` (via `movie config set tmdb-key <key>`)
3. Environment variables: `TMDB_API_KEY` (v3 key) or `TMDB_TOKEN` (v4 Read Access Token)

### Verification Check
```go
client := tmdb.NewClient(apiKey)
if err := client.VerifyAuth(); err != nil {
    // ErrAuthMissing or ErrAuthInvalid
}
```

## Primary Client Methods

```go
// Search methods with title and year filtering
SearchMovie(title string, year int) ([]MovieResult, error)
SearchTV(title string, year int) ([]TVResult, error)

// Full media metadata hydration
GetMovieDetails(tmdbID int) (*MovieDetails, error)
GetTVDetails(tmdbID int) (*TVDetails, error)
GetTVSeason(tmdbID, seasonNum int) (*SeasonDetails, error)

// Credits & Recommendations
GetCredits(tmdbID int, mediaType string) (*Credits, error)
GetRecommendations(tmdbID int, mediaType string) ([]MovieResult, error)
GetTrending(mediaType, timeWindow string) ([]TrendingResult, error)

// Asset Download
DownloadImage(posterPath, destFilePath string) error
```

## DuckDuckGo & IMDb Search Fallback Chain

When standard title queries return zero results on TMDb (common with foreign, obscure, or heavily abbreviated releases):

1. **Cache Inspection:** Queries `ImdbLookupCache` in `movie.db` by `(CleanTitle, Year)`.
   - Full hit: returns cached `(imdbID, tmdbID, mediaType)`.
   - Known miss: skips web lookup immediately.
2. **DuckDuckGo Query:** Searches `"<CleanTitle>" <Year> site:imdb.com/title`.
3. **IMDb ID Parsing:** Extracts regex pattern `tt\d{7,8}` from search result URLs.
4. **TMDb Find by External ID:** Queries TMDb endpoint `/find/{imdbID}?external_source=imdb_id`.
5. **Cache Persistence:** Stores result in `ImdbLookupCache` to avoid repeat queries.

## Rate Limiting & Sentinel Errors

Calls are protected by `limiter.go` to avoid exceeding TMDb request ceilings:
- `ErrAuthInvalid`: 401 Unauthorized (invalid API key).
- `ErrAuthMissing`: No credentials configured in DB or environment.
- `ErrRateLimited`: 429 Too Many Requests (backoff sleep triggered).
- `ErrServerError`: 5xx upstream outage.
- `ErrNetworkError`: Connection failure or DNS resolution issue.
- `ErrTimeout`: Request exceeded context deadline (default 15s).
