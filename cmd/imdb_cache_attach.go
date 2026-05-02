// SHARED HELPER: imdb_cache_attach.go — shared helper to optionally attach the IMDb cache
// imdb_cache_attach.go — shared helper to optionally attach the IMDb cache
// to a TMDb client based on a --no-cache style flag.
//
// Used by `movie rescan` and `movie rescan-failed` so a user can force a
// fresh DuckDuckGo + TMDb /find lookup without having to clear the cache
// (and lose every other cached hit) first.
package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

// attachImdbCacheUnless wires the persistent IMDb cache into client unless
// noCache is true. When bypassed it prints a one-line notice naming the
// command so the user can see the cache was intentionally skipped.
func attachImdbCacheUnless(client *tmdb.Client, database *db.DB, noCache bool, commandName string) {
	if noCache {
		fmt.Printf("⚠️  --no-cache: bypassing IMDb cache for this %s run (forcing fresh DuckDuckGo + /find).\n", commandName)
		client.SetImdbCache(nil)
		return
	}
	client.SetImdbCache(newImdbCacheAdapter(database))
}
