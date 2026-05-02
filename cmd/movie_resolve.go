// SHARED HELPER: movie_resolve.go — shared media resolver.
// movie_resolve.go — shared media resolver.
//
// SHARED: resolveMediaByQuery(db, query) — accepts a numeric ID or a fuzzy
// title match (exact → prefix → first result) and returns the matching
// *db.Media (DB-first, TMDb fallback for unknown titles).
// Callers: movie info, movie play, movie ls (detail view), movie move,
// movie rename, movie undo, movie suggest, movie cache, movie cache-forget.
// Do NOT re-implement ID-vs-title disambiguation in command files — always
// route through this resolver so behavior stays consistent.
//
// All commands that accept an <id-or-title> argument should use this
// helper to keep resolution logic consistent.  Do NOT duplicate the
// ID-parse → exact-match → prefix-match → fallback chain elsewhere.
package cmd

import (
	"strconv"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// resolveMediaByQuery resolves a media item by numeric ID or fuzzy title query.
func resolveMediaByQuery(database *db.DB, query string) (*db.Media, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, apperror.New("empty media identifier")
	}

	if id, parseErr := strconv.ParseInt(query, 10, 64); parseErr == nil {
		m, getErr := database.GetMediaByID(id)
		if getErr != nil {
			return nil, apperror.New("media not found for ID %d", id)
		}
		return m, nil
	}

	results, searchErr := database.SearchMedia(query)
	if searchErr != nil {
		return nil, searchErr
	}
	if len(results) == 0 {
		return nil, apperror.New("media not found for %q", query)
	}

	for i := range results {
		if strings.EqualFold(results[i].CleanTitle, query) || strings.EqualFold(results[i].Title, query) {
			picked := results[i]
			return &picked, nil
		}
	}

	queryLower := strings.ToLower(query)
	for i := range results {
		if strings.HasPrefix(strings.ToLower(results[i].CleanTitle), queryLower) ||
			strings.HasPrefix(strings.ToLower(results[i].Title), queryLower) {
			picked := results[i]
			return &picked, nil
		}
	}

	picked := results[0]
	return &picked, nil
}
