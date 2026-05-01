// SHARED HELPER: imdb_cache_adapter.go — adapts *db.DB to the tmdb.ImdbCache interface.
// imdb_cache_adapter.go — adapts *db.DB to the tmdb.ImdbCache interface.
//
// Lives in cmd/ so the db package stays independent of tmdb and vice-versa.
// Wire it up at every TMDb client construction site that has a *db.DB in scope:
//
//	client := tmdb.NewClientWithToken(...)
//	client.SetImdbCache(newImdbCacheAdapter(database))
package cmd

import (
	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
)

// imdbCacheAdapter wraps *db.DB so it satisfies tmdb.ImdbCache.
type imdbCacheAdapter struct {
	database *db.DB
}

func newImdbCacheAdapter(database *db.DB) *imdbCacheAdapter {
	if database == nil {
		return nil
	}
	return &imdbCacheAdapter{database: database}
}

// Look returns the cached lookup, swallowing DB errors so a broken cache
// degrades gracefully into a fresh web call rather than failing the search.
func (a *imdbCacheAdapter) Look(cleanTitle string, year int) (string, int, string, bool, bool) {
	if a == nil || a.database == nil {
		return "", 0, "", false, false
	}
	res, err := a.database.GetImdbLookup(cleanTitle, year)
	if err != nil {
		errlog.Warn("imdb cache lookup failed for '%s' (%d): %v", cleanTitle, year, err)
		return "", 0, "", false, false
	}
	return res.ImdbID, res.TmdbID, res.MediaType, res.IsHit, res.Found
}

// Store records a hit or miss; errors are logged and swallowed.
func (a *imdbCacheAdapter) Store(cleanTitle string, year int, imdbID string, tmdbID int, mediaType string) error {
	if a == nil || a.database == nil {
		return nil
	}
	if err := a.database.SetImdbLookup(cleanTitle, year, imdbID, tmdbID, mediaType); err != nil {
		errlog.Warn("imdb cache store failed for '%s' (%d): %v", cleanTitle, year, err)
		return err
	}
	return nil
}
