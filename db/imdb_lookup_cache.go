// imdb_lookup_cache.go — read/write helpers for the ImdbLookupCache table.
//
// The cache stores the result of every DuckDuckGo→IMDb lookup performed by
// tmdb.Client.tryImdbViaWeb so that repeated runs (movie scan, movie
// rescan, movie rescan-failed) do not re-hit the web for the same
// (clean title, year) pair.
//
// Since v3 the cache also stores the TmdbId + MediaType resolved from
// TMDb /find?external_source=imdb_id. When both are present a "warm" hit
// can skip the /find call too and return a synthetic SearchResult directly.
//
// Hits and misses are both cached. Misses use a shorter TTL so that titles
// that were previously unmatched eventually get retried as TMDb / IMDb data
// improves over time.
package db

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

// ImdbCacheTTLHit is how long a positive ImdbId lookup stays valid.
// IMDb ids are stable so this can be very long.
const ImdbCacheTTLHit = 180 * 24 * time.Hour // 180 days

// ImdbCacheTTLMiss is how long a "no match" result stays valid before we
// allow another web lookup. Shorter so we eventually retry.
const ImdbCacheTTLMiss = 7 * 24 * time.Hour // 7 days

// ImdbLookupResult is the outcome of a cache lookup.
type ImdbLookupResult struct {
	ImdbID    string // empty when IsHit is false
	MediaType string // "movie" or "tv"; empty when not yet resolved via /find
	TmdbID    int    // 0 when not yet resolved via /find
	IsHit     bool   // true when the cached entry recorded a real IMDb id
	Found     bool   // true when an unexpired cache entry exists at all
}

// imdbLookupKey returns the canonical cache key for (cleanTitle, year).
func imdbLookupKey(cleanTitle string, year int) string {
	normalized := strings.ToLower(strings.TrimSpace(cleanTitle))
	return normalized + "|" + strconv.Itoa(year)
}

// GetImdbLookup returns any unexpired cached result for the title/year.
// Found=false means the caller should perform a fresh web lookup.
func (d *DB) GetImdbLookup(cleanTitle string, year int) (ImdbLookupResult, error) {
	conn := d.cacheConn()
	key := imdbLookupKey(cleanTitle, year)

	row := conn.QueryRow(
		`SELECT ImdbId, IsHit, LookedUpAt, TmdbId, MediaType FROM ImdbLookupCache WHERE LookupKey = ?`,
		key,
	)

	var imdbID, lookedUpAt, mediaType string
	var tmdbID int
	var isHit bool

	err := row.Scan(&imdbID, &isHit, &lookedUpAt, &tmdbID, &mediaType)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if d.CacheDB != nil {
				fallbackRow := d.DB.QueryRow(
					`SELECT ImdbId, IsHit, LookedUpAt, TmdbId, MediaType FROM ImdbLookupCache WHERE LookupKey = ?`,
					key,
				)

				if fallbackErr := fallbackRow.Scan(&imdbID, &isHit, &lookedUpAt, &tmdbID, &mediaType); fallbackErr == nil {
					if !isCacheEntryExpired(lookedUpAt, isHit) {
						_ = d.SetImdbLookup(cleanTitle, year, imdbID, tmdbID, mediaType)

						return ImdbLookupResult{
							ImdbID:    imdbID,
							MediaType: mediaType,
							TmdbID:    tmdbID,
							IsHit:     isHit,
							Found:     true,
						}, nil
					}
				}
			}

			return ImdbLookupResult{}, nil
		}

		return ImdbLookupResult{}, err
	}

	if isCacheEntryExpired(lookedUpAt, isHit) {
		return ImdbLookupResult{}, nil
	}

	return ImdbLookupResult{
		ImdbID:    imdbID,
		MediaType: mediaType,
		TmdbID:    tmdbID,
		IsHit:     isHit,
		Found:     true,
	}, nil
}

// SetImdbLookup upserts a cache entry. Pass empty imdbID to record a miss.
// Pass tmdbID=0 / mediaType="" when the IMDb id has been resolved but the
// TMDb /find lookup hasn't run yet (or returned nothing).
func (d *DB) SetImdbLookup(cleanTitle string, year int, imdbID string, tmdbID int, mediaType string) error {
	conn := d.cacheConn()
	key := imdbLookupKey(cleanTitle, year)
	isHit := imdbID != ""
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := conn.Exec(`
		INSERT INTO ImdbLookupCache (LookupKey, CleanTitle, Year, ImdbId, IsHit, LookedUpAt, TmdbId, MediaType)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(LookupKey) DO UPDATE SET
			CleanTitle = excluded.CleanTitle,
			Year       = excluded.Year,
			ImdbId     = excluded.ImdbId,
			IsHit      = excluded.IsHit,
			LookedUpAt = excluded.LookedUpAt,
			TmdbId     = excluded.TmdbId,
			MediaType  = excluded.MediaType
	`, key, cleanTitle, year, imdbID, isHit, now, tmdbID, mediaType)

	return err
}

func isCacheEntryExpired(lookedUpAt string, isHit bool) bool {
	stamp, err := time.Parse(time.RFC3339, lookedUpAt)
	if err != nil {
		return true
	}
	ttl := ImdbCacheTTLMiss
	if isHit {
		ttl = ImdbCacheTTLHit
	}
	return time.Since(stamp) > ttl
}
