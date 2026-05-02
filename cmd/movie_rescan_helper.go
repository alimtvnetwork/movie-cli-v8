// SHARED HELPER: movie_rescan_helper.go — shared rescan logic used by both scan and rescan commands.
// movie_rescan_helper.go — shared rescan logic used by both scan and rescan commands.
//
// SHARED: title parsing + TMDb resolve + DB upsert pipeline.
// Callers: movie scan, movie rescan, movie rescan-failed.
// Do NOT re-implement the TMDb-resolve-and-upsert dance elsewhere — call
// these helpers so all paths share the same fallback chain and metadata
// merge rules.
package cmd

import (
	"regexp"
	"strconv"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

// MaxRescanAge is the staleness threshold for cached metadata.
// Rows older than this are re-fetched from TMDb even when complete.
// Spec: spec/08-app/10-remove-move-rescan/rescan-reconciliation/03-staleness-rule.md
const MaxRescanAge = 365 * 24 * time.Hour

// mediaNeedsRescan returns true when the entry is missing core fields OR
// the cached metadata is older than MaxRescanAge (1 year).
// Genre is populated from the M:N Genre/MediaGenre tables via the compat field.
func mediaNeedsRescan(m *db.Media) bool {
	if m.Genre == "" || m.TmdbRating == 0 || m.Description == "" {
		return true
	}
	return isMediaStale(m.UpdatedAt)
}

// isMediaStale parses the SQLite UpdatedAt timestamp and compares to the
// 1-year threshold. Unparseable / empty timestamps are treated as stale
// so legacy rows get refreshed once.
func isMediaStale(updatedAt string) bool {
	if updatedAt == "" {
		return true
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, updatedAt); err == nil {
			return time.Since(t) > MaxRescanAge
		}
	}
	return true
}

// rescanMediaEntry re-fetches TMDb metadata for a single media entry.
// Returns true if the entry was updated successfully.
func rescanMediaEntry(database *db.DB, client *tmdb.Client, m *db.Media) bool {
	searchTitle := m.CleanTitle
	if m.Year > 0 {
		yearStr := strconv.Itoa(m.Year)
		re := regexp.MustCompile(`\s+` + regexp.QuoteMeta(yearStr) + `$`)
		searchTitle = re.ReplaceAllString(searchTitle, "")
	}
	searchQuery := searchTitle
	if m.Year > 0 {
		searchQuery += " " + strconv.Itoa(m.Year)
	}

	tmdbResults, tmdbErr := client.SearchWithFallback(searchTitle, m.Year)
	if tmdbErr != nil {
		errlog.Warn("rescan TMDb search failed for '%s': %v", searchQuery, tmdbErr)
		return false
	}
	if len(tmdbResults) == 0 {
		return false
	}

	best := tmdbResults[0]
	m.TmdbID = best.ID
	m.TmdbRating = best.VoteAvg
	m.Popularity = best.Popularity
	m.Description = best.Overview
	m.Genre = tmdb.GenreNames(best.GenreIDs)

	m.Type = resolveMediaType(best.MediaType)
	fetchDetailsByType(client, best.ID, m)

	if !updateRescanEntry(database, m) {
		return false
	}

	linkRescanRelations(database, m)
	return true
}

// updateRescanEntry persists updated media to DB, trying TmdbID first.
func updateRescanEntry(database *db.DB, m *db.Media) bool {
	if m.TmdbID <= 0 {
		return updateByID(database, m)
	}
	if err := database.UpdateMediaByTmdbID(m); err == nil {
		return true
	}
	return updateByID(database, m)
}

func updateByID(database *db.DB, m *db.Media) bool {
	if err := database.UpdateMediaByID(m); err != nil {
		errlog.Error("rescan DB update failed for '%s': %v", m.Title, err)
		return false
	}
	return true
}

func linkRescanRelations(database *db.DB, m *db.Media) {
	if m.ID <= 0 {
		return
	}
	if m.Genre != "" {
		_ = database.ReplaceMediaGenres(m.ID, m.Genre)
	}
	if m.Director != "" {
		database.ReplaceMediaDirectors(m.ID, m.Director)
	}
}
