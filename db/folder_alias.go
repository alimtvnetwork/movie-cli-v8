// folder_alias.go — Folder aliases and scanned folder statistics.
package db

import (
	"path/filepath"
	"strings"
)

// FolderAliasRecord represents a shortcut alias for a scanned directory.
type FolderAliasRecord struct {
	CreatedAt  string `json:"created_at"`
	AliasName  string `json:"alias_name"`
	FolderPath string `json:"folder_path"`
	ID         int64  `json:"id"`
}

// ScanFolderStat represents a scanned root folder with item counts and alias.
type ScanFolderStat struct {
	FolderPath  string `json:"folder_path"`
	Alias       string `json:"alias"`
	ItemCount   int    `json:"item_count"`
	MoviesCount int    `json:"movies_count"`
	TvCount     int    `json:"tv_count"`
}

// DeriveCleanAlias produces a short alphanumeric alias from a folder path.
func DeriveCleanAlias(folderPath string) string {
	clean := filepath.Clean(folderPath)
	base := filepath.Base(clean)

	var sb strings.Builder

	for _, ch := range strings.ToLower(base) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			sb.WriteRune(ch)
		}
	}

	alias := sb.String()

	if alias == "" {
		alias = "folder"
	}

	return alias
}

// SetFolderAlias creates or updates an alias for a folder path.
func (d *DB) SetFolderAlias(name, folderPath string) error {
	cleanName := strings.ToLower(strings.TrimSpace(name))
	cleanPath := filepath.Clean(strings.TrimSpace(folderPath))

	_, err := d.Exec(`
		INSERT INTO FolderAlias (AliasName, FolderPath)
		VALUES (?, ?)
		ON CONFLICT(AliasName) DO UPDATE SET FolderPath = excluded.FolderPath`, cleanName, cleanPath)

	return err
}

// GetFolderAlias returns the folder path associated with an alias.
func (d *DB) GetFolderAlias(name string) (string, error) {
	cleanName := strings.ToLower(strings.TrimSpace(name))

	var path string
	err := d.QueryRow(`SELECT FolderPath FROM FolderAlias WHERE AliasName = ?`, cleanName).Scan(&path)

	if err != nil {
		return "", err
	}

	return path, nil
}

// ListFolderAliases returns all registered folder aliases.
func (d *DB) ListFolderAliases() ([]FolderAliasRecord, error) {
	rows, err := d.Query(`SELECT AliasId, AliasName, FolderPath, CreatedAt FROM FolderAlias ORDER BY AliasName ASC`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var records []FolderAliasRecord

	for rows.Next() {
		var r FolderAliasRecord

		if scanErr := rows.Scan(&r.ID, &r.AliasName, &r.FolderPath, &r.CreatedAt); scanErr == nil {
			records = append(records, r)
		}
	}

	return records, nil
}

// DeleteFolderAlias deletes a folder alias by name.
func (d *DB) DeleteFolderAlias(name string) error {
	cleanName := strings.ToLower(strings.TrimSpace(name))

	_, err := d.Exec(`DELETE FROM FolderAlias WHERE AliasName = ?`, cleanName)

	return err
}

// ListScanFoldersWithStats returns all registered scan folders with item counts and resolved aliases.
func (d *DB) ListScanFoldersWithStats() ([]ScanFolderStat, error) {
	folders, err := d.ListDistinctScanFolders()

	if err != nil {
		return nil, err
	}

	aliases, _ := d.ListFolderAliases()
	aliasMap := make(map[string]string)

	for _, a := range aliases {
		aliasMap[filepath.Clean(a.FolderPath)] = a.AliasName
	}

	stats := make([]ScanFolderStat, 0, len(folders))

	for _, f := range folders {
		cleanFolder := filepath.Clean(f)
		alias := aliasMap[cleanFolder]

		if alias == "" {
			alias = DeriveCleanAlias(cleanFolder)
		}

		prefix := cleanFolder + string(filepath.Separator)

		var total, movies, tv int
		row := d.QueryRow(`
			SELECT
				COUNT(*),
				COUNT(CASE WHEN Type = 'movie' THEN 1 END),
				COUNT(CASE WHEN Type = 'tv' THEN 1 END)
			FROM Media
			WHERE IsDeleted = 0
			  AND (
			      CurrentFilePath = ? OR CurrentFilePath LIKE ?
			      OR (CurrentFilePath = '' AND (OriginalFilePath = ? OR OriginalFilePath LIKE ?))
			  )`, cleanFolder, prefix+"%", cleanFolder, prefix+"%")

		_ = row.Scan(&total, &movies, &tv)

		stats = append(stats, ScanFolderStat{
			FolderPath:  cleanFolder,
			Alias:       alias,
			ItemCount:   total,
			MoviesCount: movies,
			TvCount:     tv,
		})
	}

	return stats, nil
}
