// media_query.go — query methods extracted from media.go.
package db

import (
	"database/sql"
	"path/filepath"
)

// ListMedia returns paginated media records with genres populated.
func (d *DB) ListMedia(offset, limit int) ([]Media, error) {
	rows, err := d.Query(`SELECT `+mediaColumns+`
		FROM Media WHERE OriginalFilePath != ''
		ORDER BY CleanTitle ASC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}
	d.populateGenres(items)
	return items, nil
}

// ListMediaFiltered returns paginated media records using a filter mode.
// Modes:
//   - "scanned" (default): only items with a known file path (OriginalFilePath != ”)
//   - "all":     every media row, including metadata-only entries with no file
//   - "missing": only items with NO file path (metadata-only / never scanned or removed)
func (d *DB) ListMediaFiltered(offset, limit int, mode string) ([]Media, error) {
	where := mediaFilterWhere(mode)
	rows, err := d.Query(`SELECT `+mediaColumns+`
		FROM Media `+where+`
		ORDER BY CleanTitle ASC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}
	d.populateGenres(items)
	return items, nil
}

// CountMediaFiltered returns the total count for a given filter mode.
// See ListMediaFiltered for accepted modes.
func (d *DB) CountMediaFiltered(mode string) (int, error) {
	var count int
	where := mediaFilterWhere(mode)
	err := d.QueryRow(`SELECT COUNT(*) FROM Media ` + where).Scan(&count)
	return count, err
}

// mediaFilterWhere maps a filter mode to its SQL WHERE clause.
// SHARED: used by ListMediaFiltered and CountMediaFiltered to keep the
// scanned/all/missing semantics in one place.
func mediaFilterWhere(mode string) string {
	switch mode {
	case "all":
		return "WHERE 1=1"
	case "missing":
		return "WHERE COALESCE(OriginalFilePath, '') = ''"
	default: // "scanned"
		return "WHERE OriginalFilePath != ''"
	}
}

// SearchMedia searches by title (fuzzy via LIKE).
func (d *DB) SearchMedia(query string) ([]Media, error) {
	rows, err := d.Query(`SELECT `+mediaColumns+`
		FROM Media WHERE CleanTitle LIKE ? OR Title LIKE ?
		ORDER BY Popularity DESC LIMIT 20`, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}
	d.populateGenres(items)
	return items, nil
}

// GetMediaByID returns a single media record with genres populated.
func (d *DB) GetMediaByID(id int64) (*Media, error) {
	row := d.QueryRow(`SELECT `+mediaColumns+` FROM Media WHERE MediaId = ?`, id)
	m, err := scanMediaRow(row)
	if err != nil {
		return nil, err
	}
	d.populateGenre(m)
	return m, nil
}

// GetMediaByTmdbID returns a media record by its TMDb ID with genres populated.
func (d *DB) GetMediaByTmdbID(tmdbID int) (*Media, error) {
	row := d.QueryRow(`SELECT `+mediaColumns+` FROM Media WHERE TmdbId = ?`, tmdbID)
	m, err := scanMediaRow(row)
	if err != nil {
		return nil, err
	}
	d.populateGenre(m)
	return m, nil
}

// CountMedia returns total count of scan-indexed items.
func (d *DB) CountMedia(mediaType string) (int, error) {
	var count int
	if mediaType == "" {
		return count, d.QueryRow("SELECT COUNT(*) FROM Media WHERE OriginalFilePath != ''").Scan(&count)
	}
	return count, d.QueryRow("SELECT COUNT(*) FROM Media WHERE Type = ? AND OriginalFilePath != ''", mediaType).Scan(&count)
}

// ListAllMedia returns all media records that have a file path with genres populated.
func (d *DB) ListAllMedia() ([]Media, error) {
	rows, err := d.Query(`SELECT ` + mediaColumns + `
		FROM Media WHERE OriginalFilePath != ''
		ORDER BY CleanTitle ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}
	d.populateGenres(items)
	return items, nil
}

// GetMediaWithMissingTmdbID returns entries that never matched a TMDb id
// (TmdbId is NULL or 0). Used by `movie rescan-failed`.
func (d *DB) GetMediaWithMissingTmdbID() ([]Media, error) {
	rows, err := d.Query(`SELECT ` + mediaColumns + `
		FROM Media WHERE OriginalFilePath != ''
		AND COALESCE(TmdbId, 0) = 0
		ORDER BY CleanTitle ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}
	d.populateGenres(items)
	return items, nil
}

// GetMediaWithMissingData returns entries with no genres, no rating, or no description.
func (d *DB) GetMediaWithMissingData() ([]Media, error) {
	rows, err := d.Query(`SELECT ` + mediaColumns + `
		FROM Media WHERE OriginalFilePath != ''
		AND (
			COALESCE(TmdbRating, 0) = 0
			OR COALESCE(Description, '') = ''
			OR MediaId NOT IN (SELECT DISTINCT MediaId FROM MediaGenre)
		)
		ORDER BY CleanTitle ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}
	d.populateGenres(items)
	return items, nil
}

// GetMediaByScanDir returns media whose OriginalFilePath starts with the given directory.
func (d *DB) GetMediaByScanDir(scanDir string) ([]Media, error) {
	prefix := scanDir
	if prefix != "" && prefix[len(prefix)-1] != '/' && prefix[len(prefix)-1] != '\\' {
		prefix += string([]byte{filepath.Separator})
	}
	rows, err := d.Query(`SELECT `+mediaColumns+`
		FROM Media WHERE OriginalFilePath LIKE ?
		ORDER BY CleanTitle ASC`, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}
	d.populateGenres(items)
	return items, nil
}

// DeleteMediaByIDs deletes multiple media records by their IDs.
func (d *DB) DeleteMediaByIDs(ids []int64) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	tx, err := d.Begin()
	if err != nil {
		return 0, err
	}
	deleted := 0
	for _, id := range ids {
		if _, err := tx.Exec("DELETE FROM Media WHERE MediaId = ?", id); err != nil {
			_ = tx.Rollback()
			return deleted, err
		}
		deleted++
	}
	return deleted, tx.Commit()
}

// FileSizeStats returns total, largest, and smallest file size in MB.
func (d *DB) FileSizeStats() (total float64, largest float64, smallest float64, err error) {
	err = d.QueryRow(`
		SELECT COALESCE(SUM(FileSizeMb), 0),
		       COALESCE(MAX(FileSizeMb), 0),
		       COALESCE(MIN(NULLIF(FileSizeMb, 0)), 0)
		FROM Media WHERE FileSizeMb > 0`).Scan(&total, &largest, &smallest)
	return
}

// MediaByType returns media filtered by type with genres populated.
func (d *DB) MediaByType(mediaType string, limit int) ([]Media, error) {
	rows, err := d.Query(`SELECT `+mediaColumns+`
		FROM Media WHERE Type = ? ORDER BY Popularity DESC LIMIT ?`, mediaType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanMediaRows(rows)
	if err != nil {
		return nil, err
	}
	d.populateGenres(items)
	return items, nil
}

// TopGenres returns genres sorted by frequency via the normalized Genre/MediaGenre tables.
func (d *DB) TopGenres(limit int) (map[string]int, error) {
	rows, err := d.Query(`
		SELECT g.Name, COUNT(*) as cnt
		FROM MediaGenre mg
		INNER JOIN Genre g ON mg.GenreId = g.GenreId
		GROUP BY g.Name
		ORDER BY cnt DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var name string
		var cnt int
		if err := rows.Scan(&name, &cnt); err != nil {
			return nil, err
		}
		counts[name] = cnt
	}
	return counts, rows.Err()
}

func scanMediaRow(row *sql.Row) (*Media, error) {
	m := &Media{}
	err := row.Scan(&m.ID, &m.Title, &m.CleanTitle, &m.Year, &m.Type,
		&m.TmdbID, &m.ImdbID, &m.Description, &m.ImdbRating, &m.TmdbRating,
		&m.Popularity, &m.LanguageId, &m.CollectionId,
		&m.Director, &m.ThumbnailPath,
		&m.OriginalFileName, &m.OriginalFilePath, &m.CurrentFilePath,
		&m.FileExtension, &m.FileSizeMb,
		&m.Runtime, &m.Budget, &m.Revenue, &m.TrailerURL, &m.Tagline,
		&m.ScanHistoryId, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// populateGenre loads the Genre compat field from the MediaGenre join table.
func (d *DB) populateGenre(m *Media) {
	if m == nil || m.ID == 0 {
		return
	}
	g, err := d.GetMediaGenres(m.ID)
	if err == nil {
		m.Genre = g
	}
}

// populateGenres loads Genre compat fields for a list of media.
func (d *DB) populateGenres(items []Media) {
	for i := range items {
		d.populateGenre(&items[i])
	}
}

// scanMediaRows scans multiple media rows from a query result.
func scanMediaRows(rows *sql.Rows) ([]Media, error) {
	var list []Media
	for rows.Next() {
		var m Media
		if err := rows.Scan(&m.ID, &m.Title, &m.CleanTitle, &m.Year, &m.Type,
			&m.TmdbID, &m.ImdbID, &m.Description, &m.ImdbRating, &m.TmdbRating,
			&m.Popularity, &m.LanguageId, &m.CollectionId,
			&m.Director, &m.ThumbnailPath,
			&m.OriginalFileName, &m.OriginalFilePath, &m.CurrentFilePath,
			&m.FileExtension, &m.FileSizeMb,
			&m.Runtime, &m.Budget, &m.Revenue, &m.TrailerURL, &m.Tagline,
			&m.ScanHistoryId, &m.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}
