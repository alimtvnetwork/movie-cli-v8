// director.go — Director normalization via M:N table.
package db

import (
	"strings"
)

// LinkMediaDirectors normalizes a comma-separated director string into
// the Director + MediaDirector M:N tables.
func (d *DB) LinkMediaDirectors(mediaID int64, directorCSV string) error {
	names := splitDirectors(directorCSV)
	for _, name := range names {
		dirID, err := d.ensureDirector(name)
		if err != nil {
			return err
		}
		if dirID <= 0 {
			continue
		}
		_, err = d.Exec(
			"INSERT OR IGNORE INTO MediaDirector (MediaId, DirectorId) VALUES (?, ?)",
			mediaID, dirID)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReplaceMediaDirectors replaces all director links for a media entry.
func (d *DB) ReplaceMediaDirectors(mediaID int64, directorCSV string) {
	_, _ = d.Exec("DELETE FROM MediaDirector WHERE MediaId = ?", mediaID)
	_ = d.LinkMediaDirectors(mediaID, directorCSV)
}

// DirectorsByMediaID returns comma-separated director names for a media entry.
func (d *DB) DirectorsByMediaID(mediaID int64) string {
	rows, err := d.Query(`
		SELECT d.Name FROM Director d
		JOIN MediaDirector md ON md.DirectorId = d.DirectorId
		WHERE md.MediaId = ?
		ORDER BY d.Name`, mediaID)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			names = append(names, name)
		}
	}
	return strings.Join(names, ", ")
}

// ensureDirector inserts or finds a director by name, returns DirectorId.
//
// IMPORTANT: Never trust res.LastInsertId() after INSERT OR IGNORE — when the
// UNIQUE(Name) conflict fires, mattn/go-sqlite3 returns the last successful
// rowid on the connection (potentially from a different table), which would
// produce a stale DirectorId and trigger MediaDirector FK error 787.
// Always SELECT the canonical PK back. See spec/09-app-issues/09-director-fk-stale-lastinsertid.md.
func (d *DB) ensureDirector(name string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, nil
	}
	if _, err := d.Exec("INSERT OR IGNORE INTO Director (Name) VALUES (?)", name); err != nil {
		return 0, err
	}
	var dirID int64
	err := d.QueryRow("SELECT DirectorId FROM Director WHERE Name = ?", name).Scan(&dirID)
	return dirID, err
}

// splitDirectors splits a comma-separated director string into trimmed names.
func splitDirectors(csv string) []string {
	parts := strings.Split(csv, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
