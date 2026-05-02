// SHARED HELPER: movie_global_cache.go — cross-folder global JSON cache.
// movie_global_cache.go — cross-folder global JSON cache.
//
// The per-folder sidecars at <scanDir>/.movie-output/json/{movie,tv}/<slug>.json
// are great for portability with one library, but they live and die with
// the folder. The GLOBAL cache mirrors every successful sidecar write into
// a per-user location keyed by TmdbID so:
//
//   - reinstalling the OS / wiping .movie-output keeps your metadata
//   - moving a file from folder A to folder B keeps its enrichment
//   - future commands can answer "do we already know this TmdbID?" with
//     a single file stat instead of a TMDb call
//
// Layout:  $HOME/.movie/cache/json/{movie|tv}/<tmdbid>.json
//
// The cache is best-effort — every helper silently no-ops on error so a
// broken cache never blocks a scan. Disable with env MOVIE_NO_GLOBAL_CACHE=1.
package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// EnvDisableGlobalCache turns the mirror + lookup into no-ops when set to "1".
const EnvDisableGlobalCache = "MOVIE_NO_GLOBAL_CACHE"

// globalCacheDir returns the per-user cache root, or "" when unavailable.
func globalCacheDir() string {
	if os.Getenv(EnvDisableGlobalCache) == "1" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".movie", "cache", "json")
}

// globalCachePath returns the absolute path for a TmdbID/Type, or "".
func globalCachePath(mediaType string, tmdbID int) string {
	root := globalCacheDir()
	if root == "" || tmdbID <= 0 || mediaType == "" {
		return ""
	}
	return filepath.Join(root, db.JsonSubDir(mediaType), strconv.Itoa(tmdbID)+".json")
}

// MirrorToGlobalCache writes a copy of the per-folder sidecar into the
// shared cache. Safe to call after every successful sidecar write.
func MirrorToGlobalCache(item scanMediaJSON) {
	dst := globalCachePath(item.Type, item.TmdbID)
	if dst == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return
	}
	raw, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(dst, raw, 0644)
}

// LookupGlobalCache returns the cached sidecar for a TmdbID/Type, or
// (zero, false) when missing/disabled/corrupt.
func LookupGlobalCache(mediaType string, tmdbID int) (scanMediaJSON, bool) {
	path := globalCachePath(mediaType, tmdbID)
	if path == "" {
		return scanMediaJSON{}, false
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return scanMediaJSON{}, false
	}
	var item scanMediaJSON
	if jsonErr := json.Unmarshal(raw, &item); jsonErr != nil {
		return scanMediaJSON{}, false
	}
	if item.Title == "" || item.Type == "" {
		return scanMediaJSON{}, false
	}
	return item, true
}
