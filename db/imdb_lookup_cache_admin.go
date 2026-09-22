// imdb_lookup_cache_admin.go — admin-facing helpers for the ImdbLookupCache.
//
// Used by the `movie cache imdb` command to inspect and invalidate cache rows
// without opening the SQLite file directly.
package db

import "database/sql"

// ImdbCacheEntry is one row from the ImdbLookupCache table, in display form.
type ImdbCacheEntry struct {
	LookupKey  string
	CleanTitle string
	ImdbID     string
	MediaType  string
	LookedUpAt string
	Year       int
	TmdbID     int
	IsHit      bool
}

// ListImdbLookups returns every cached entry ordered by most recent first.
// Pass limit <= 0 for "all rows".
func (d *DB) ListImdbLookups(limit int) ([]ImdbCacheEntry, error) {
	rows, err := queryImdbCache(d, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanImdbCacheRows(rows)
}

func queryImdbCache(d *DB, limit int) (*sql.Rows, error) {
	conn := d.cacheConn()
	base := `SELECT LookupKey, CleanTitle, Year, ImdbId, IsHit, LookedUpAt, TmdbId, MediaType
	         FROM ImdbLookupCache
	         ORDER BY LookedUpAt DESC`

	if limit > 0 {
		return conn.Query(base+" LIMIT ?", limit)
	}

	return conn.Query(base)
}

func scanImdbCacheRows(rows *sql.Rows) ([]ImdbCacheEntry, error) {
	var out []ImdbCacheEntry

	for rows.Next() {
		var e ImdbCacheEntry

		err := rows.Scan(&e.LookupKey, &e.CleanTitle, &e.Year, &e.ImdbID, &e.IsHit, &e.LookedUpAt, &e.TmdbID, &e.MediaType)

		if err != nil {
			return nil, err
		}

		out = append(out, e)
	}

	return out, rows.Err()
}

// ListImdbLookupsUnresolved returns every cached HIT row whose TmdbId is 0
// (legacy v2 entries or partial hits where /find was never called or returned
// nothing). These are the rows that `movie cache imdb backfill` needs to
// re-resolve via TMDb /find. Misses are skipped because they have no IMDb id
// to look up.
func (d *DB) ListImdbLookupsUnresolved() ([]ImdbCacheEntry, error) {
	conn := d.cacheConn()

	rows, err := conn.Query(`SELECT LookupKey, CleanTitle, Year, ImdbId, IsHit, LookedUpAt, TmdbId, MediaType
		FROM ImdbLookupCache
		WHERE IsHit = 1 AND TmdbId = 0 AND ImdbId != ''
		ORDER BY LookedUpAt ASC`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanImdbCacheRows(rows)
}

// CountImdbLookups returns (totalRows, hitRows). missRows = total - hits.
func (d *DB) CountImdbLookups() (int, int, error) {
	conn := d.cacheConn()
	var total, hits int

	row := conn.QueryRow(`SELECT COUNT(*), COALESCE(SUM(CASE WHEN IsHit THEN 1 ELSE 0 END), 0) FROM ImdbLookupCache`)

	if err := row.Scan(&total, &hits); err != nil {
		return 0, 0, err
	}

	return total, hits, nil
}

// ClearImdbLookups deletes every row from ImdbLookupCache. Returns row count removed.
func (d *DB) ClearImdbLookups() (int64, error) {
	conn := d.cacheConn()

	res, err := conn.Exec(`DELETE FROM ImdbLookupCache`)

	if err != nil {
		return 0, err
	}

	if d.CacheDB != nil {
		_, _ = d.DB.Exec(`DELETE FROM ImdbLookupCache`)
	}

	return res.RowsAffected()
}

// ClearImdbLookupMisses deletes only rows where IsHit = 0 (negative cache).
// Useful when you want to retry titles that previously failed without losing
// the long-lived hit cache.
func (d *DB) ClearImdbLookupMisses() (int64, error) {
	conn := d.cacheConn()

	res, err := conn.Exec(`DELETE FROM ImdbLookupCache WHERE IsHit = 0`)

	if err != nil {
		return 0, err
	}

	if d.CacheDB != nil {
		_, _ = d.DB.Exec(`DELETE FROM ImdbLookupCache WHERE IsHit = 0`)
	}

	return res.RowsAffected()
}

// ForgetImdbLookup deletes the single cache row matching (cleanTitle, year).
// Returns the number of rows removed (0 if no row existed). Used by
// `movie cache imdb forget` to invalidate one stale resolution without
// nuking the entire cache.
func (d *DB) ForgetImdbLookup(cleanTitle string, year int) (int64, error) {
	conn := d.cacheConn()
	key := imdbLookupKey(cleanTitle, year)

	res, err := conn.Exec(`DELETE FROM ImdbLookupCache WHERE LookupKey = ?`, key)

	if err != nil {
		return 0, err
	}

	if d.CacheDB != nil {
		_, _ = d.DB.Exec(`DELETE FROM ImdbLookupCache WHERE LookupKey = ?`, key)
	}

	return res.RowsAffected()
}
