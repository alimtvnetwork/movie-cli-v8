// movie_cd_resolve.go — Target directory resolution logic for movie cd.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// CdTargetResult describes the successfully resolved directory.
type CdTargetResult struct {
	TargetDirectory string
	MatchName       string
	MatchType       string
	MovieTitle      string
	ItemCount       int
}

// CdSuggestion represents an entry in an interactive suggestion list.
type CdSuggestion struct {
	Name   string
	Type   string
	Path   string
	Detail string
	Number int
}

func resolveCdTarget(database *db.DB, query string, alias string) (*CdTargetResult, []CdSuggestion, error) {
	if alias != "" {
		return resolveByAlias(database, alias)
	}

	trimmed := strings.TrimSpace(query)

	if trimmed == "" {
		return nil, buildAllSuggestions(database), nil
	}

	if trimmed == "." {
		cwd, _ := os.Getwd()

		return &CdTargetResult{
			TargetDirectory: cwd,
			MatchName:       filepath.Base(cwd),
			MatchType:       "current-directory",
		}, nil, nil
	}

	if fi, statErr := os.Stat(trimmed); statErr == nil {
		if fi.IsDir() {
			absPath, _ := filepath.Abs(trimmed)

			return &CdTargetResult{
				TargetDirectory: absPath,
				MatchName:       filepath.Base(absPath),
				MatchType:       "directory",
			}, nil, nil
		}
	}

	if num, err := strconv.Atoi(trimmed); err == nil && num > 0 {
		return resolveByNumber(database, num)
	}

	if res, _, err := resolveByAlias(database, trimmed); err == nil && res != nil {
		return res, nil, nil
	}

	if res, suggs, err := resolveByFolderSubstring(database, trimmed); err == nil {
		if res != nil {
			return res, nil, nil
		}

		if len(suggs) > 0 {
			return nil, suggs, nil
		}
	}

	return resolveByMovieTitle(database, trimmed)
}

func resolveByAlias(database *db.DB, alias string) (*CdTargetResult, []CdSuggestion, error) {
	cleanAlias := strings.ToLower(strings.TrimSpace(alias))

	if path, err := database.GetFolderAlias(cleanAlias); err == nil && path != "" {
		return &CdTargetResult{
			TargetDirectory: path,
			MatchName:       cleanAlias,
			MatchType:       "alias",
		}, nil, nil
	}

	stats, _ := database.ListScanFoldersWithStats()

	for _, s := range stats {
		if strings.EqualFold(s.Alias, cleanAlias) {
			return &CdTargetResult{
				TargetDirectory: s.FolderPath,
				MatchName:       s.Alias,
				MatchType:       "folder-alias",
				ItemCount:       s.ItemCount,
			}, nil, nil
		}
	}

	return nil, nil, fmt.Errorf("no folder alias matching '%s'", alias)
}

func resolveByNumber(database *db.DB, num int) (*CdTargetResult, []CdSuggestion, error) {
	stats, err := database.ListScanFoldersWithStats()

	if err != nil || len(stats) == 0 {
		return nil, nil, fmt.Errorf("no scanned folders registered")
	}

	if num <= len(stats) {
		selected := stats[num-1]

		return &CdTargetResult{
			TargetDirectory: selected.FolderPath,
			MatchName:       selected.Alias,
			MatchType:       "folder",
			ItemCount:       selected.ItemCount,
		}, nil, nil
	}

	return nil, nil, fmt.Errorf("folder number %d out of range (1-%d)", num, len(stats))
}

func resolveByFolderSubstring(database *db.DB, query string) (*CdTargetResult, []CdSuggestion, error) {
	stats, err := database.ListScanFoldersWithStats()

	if err != nil {
		return nil, nil, err
	}

	queryLower := strings.ToLower(query)
	var matches []db.ScanFolderStat

	for _, s := range stats {
		if strings.Contains(strings.ToLower(s.FolderPath), queryLower) || strings.Contains(strings.ToLower(s.Alias), queryLower) {
			matches = append(matches, s)
		}
	}

	if len(matches) == 1 {
		return &CdTargetResult{
			TargetDirectory: matches[0].FolderPath,
			MatchName:       matches[0].Alias,
			MatchType:       "folder",
			ItemCount:       matches[0].ItemCount,
		}, nil, nil
	}

	if len(matches) > 1 {
		suggestions := make([]CdSuggestion, 0, len(matches))

		for i, m := range matches {
			suggestions = append(suggestions, CdSuggestion{
				Number: i + 1,
				Name:   m.Alias,
				Type:   "Folder",
				Path:   m.FolderPath,
				Detail: fmt.Sprintf("%d items", m.ItemCount),
			})
		}

		return nil, suggestions, nil
	}

	return nil, nil, nil
}

func resolveByMovieTitle(database *db.DB, query string) (*CdTargetResult, []CdSuggestion, error) {
	pattern := "%" + query + "%"
	rows, err := database.Query(`
		SELECT MediaId, Title, CleanTitle, Year, CurrentFilePath, OriginalFilePath
		FROM Media
		WHERE IsDeleted = 0
		  AND (Title LIKE ? OR CleanTitle LIKE ? OR OriginalFileName LIKE ?)
		ORDER BY TmdbRating DESC LIMIT 15`, pattern, pattern, pattern)

	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	var matches []movieMatch

	for rows.Next() {
		var m movieMatch

		if scanErr := rows.Scan(&m.ID, &m.Title, &m.CleanTitle, &m.Year, &m.CurPath, &m.OrigPath); scanErr == nil {
			matches = append(matches, m)
		}
	}

	return selectBestMovieMatch(matches, query)
}

type movieMatch struct {
	Title      string
	CleanTitle string
	CurPath    string
	OrigPath   string
	ID         int64
	Year       int
}

func selectBestMovieMatch(matches []movieMatch, query string) (*CdTargetResult, []CdSuggestion, error) {
	if len(matches) == 0 {
		return nil, nil, fmt.Errorf("no movie or folder matches '%s'", query)
	}

	for _, m := range matches {
		if strings.EqualFold(m.Title, query) || strings.EqualFold(m.CleanTitle, query) {
			dir := extractMovieDir(m)

			return &CdTargetResult{
				TargetDirectory: dir,
				MatchName:       m.Title,
				MatchType:       "movie",
				MovieTitle:      m.Title,
			}, nil, nil
		}
	}

	if len(matches) == 1 {
		m := matches[0]
		dir := extractMovieDir(m)

		return &CdTargetResult{
			TargetDirectory: dir,
			MatchName:       m.Title,
			MatchType:       "movie",
			MovieTitle:      m.Title,
		}, nil, nil
	}

	suggestions := make([]CdSuggestion, 0, len(matches))

	for i, m := range matches {
		suggestions = append(suggestions, CdSuggestion{
			Number: i + 1,
			Name:   m.Title,
			Type:   "Movie",
			Path:   extractMovieDir(m),
			Detail: fmt.Sprintf("(%d)", m.Year),
		})
	}

	return nil, suggestions, nil
}

func extractMovieDir(m movieMatch) string {
	path := m.CurPath

	if path == "" {
		path = m.OrigPath
	}

	return filepath.Dir(path)
}

func buildAllSuggestions(database *db.DB) []CdSuggestion {
	stats, _ := database.ListScanFoldersWithStats()
	suggestions := make([]CdSuggestion, 0, len(stats))

	for i, s := range stats {
		suggestions = append(suggestions, CdSuggestion{
			Number: i + 1,
			Name:   s.Alias,
			Type:   "Root Folder",
			Path:   s.FolderPath,
			Detail: fmt.Sprintf("%d items", s.ItemCount),
		})
	}

	return suggestions
}
