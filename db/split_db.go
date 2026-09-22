// split_db.go — SQLite Split-DB architecture for movie-cli.
// Separates Primary Library DB (movie.db) from Ephemeral Cache DB (cache.db) with WAL mode tuning.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

const (
	// CacheDBFileName is the filename of the ephemeral cache database.
	CacheDBFileName = "cache.db"
)

// SplitDBTierInfo holds diagnostic metadata for a single database tier.
type SplitDBTierInfo struct {
	Name          string
	Type          string
	Location      string
	SizeFormatted string
	JournalMode   string
	Purpose       string
	SizeBytes     int64
	TableCount    int
}

// SplitDBStatus holds the complete multi-tier SQLite Split-DB status.
type SplitDBStatus struct {
	TotalSize  string
	MasterTier SplitDBTierInfo
	CacheTier  SplitDBTierInfo
	TotalFiles int
}

func (d *DB) cacheConn() *sql.DB {
	if d.CacheDB != nil {
		return d.CacheDB
	}

	return d.DB
}

func openAndConfigureCacheDB(base string) (*sql.DB, error) {
	cachePath := filepath.Join(base, CacheDBFileName)
	conn, err := sql.Open("sqlite", cachePath)

	if err != nil {
		return nil, appfault.Wrap("cannot open cache database", err)
	}

	conn.SetMaxOpenConns(1)

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA synchronous = NORMAL",
	}

	for _, p := range pragmas {
		if _, execErr := conn.Exec(p); execErr != nil {
			conn.Close()

			return nil, appfault.Wrap("cannot configure cache pragma: "+p, execErr)
		}
	}

	return conn, nil
}

func (d *DB) initCacheSchema() error {
	if d.CacheDB == nil {
		return nil
	}

	schema := `
		CREATE TABLE IF NOT EXISTS ImdbLookupCache (
			LookupKey TEXT PRIMARY KEY,
			CleanTitle TEXT NOT NULL,
			Year INTEGER NOT NULL,
			ImdbId TEXT NOT NULL DEFAULT '',
			IsHit INTEGER NOT NULL DEFAULT 0,
			LookedUpAt TEXT NOT NULL,
			TmdbId INTEGER NOT NULL DEFAULT 0,
			MediaType TEXT NOT NULL DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS idx_imdb_lookup_key ON ImdbLookupCache(LookupKey);
	`

	_, err := d.CacheDB.Exec(schema)

	if err != nil {
		return appfault.Wrap("cannot initialize cache schema", err)
	}

	return nil
}

// GetSplitDBStatus compiles metadata for master and cache SQLite databases.
func (d *DB) GetSplitDBStatus() SplitDBStatus {
	masterPath := filepath.Join(d.BasePath, dbFile)
	cachePath := filepath.Join(d.BasePath, CacheDBFileName)

	masterTier := queryTierInfo(d.DB, "movie.db", "Primary Library DB", masterPath,
		"Central SQLite database storing indexed movies, TV series, episodes, watchlists, tags, and action history.")

	cacheConn := d.cacheConn()
	cacheTier := queryTierInfo(cacheConn, "cache.db", "Lookup & Search Cache DB", cachePath,
		"Isolates DuckDuckGo IMDb lookups, TMDb search cache, and transient web queries to prevent database write locks and contention.")

	totalBytes := masterTier.SizeBytes + cacheTier.SizeBytes

	return SplitDBStatus{
		MasterTier: masterTier,
		CacheTier:  cacheTier,
		TotalSize:  formatByteSize(totalBytes),
		TotalFiles: 2,
	}
}

func queryTierInfo(conn *sql.DB, name, tierType, path, purpose string) SplitDBTierInfo {
	info := SplitDBTierInfo{
		Name:     name,
		Type:     tierType,
		Location: path,
		Purpose:  purpose,
	}

	fi, statErr := os.Stat(path)

	if statErr == nil {
		info.SizeBytes = fi.Size()
		info.SizeFormatted = formatByteSize(fi.Size())
	}

	if conn != nil {
		var jMode string
		if err := conn.QueryRow("PRAGMA journal_mode;").Scan(&jMode); err == nil {
			info.JournalMode = jMode
		}

		var count int
		if err := conn.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table';").Scan(&count); err == nil {
			info.TableCount = count
		}
	}

	return info
}

func formatByteSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	kb := float64(bytes) / 1024.0

	if kb < 1024.0 {
		return fmt.Sprintf("%.1f KB", kb)
	}

	mb := kb / 1024.0

	return fmt.Sprintf("%.1f MB", mb)
}
