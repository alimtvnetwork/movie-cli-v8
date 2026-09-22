// movie_rest_details.go — modal payload endpoint: /api/media/{id}/details.
package cmd

import (
	"database/sql"
	"net/http"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func handleMediaDetails(w http.ResponseWriter, r *http.Request, database *db.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	id := parseMediaSubpath(r.URL.Path, "details")

	if id <= 0 {
		http.Error(w, "invalid path", http.StatusBadRequest)

		return
	}

	m, err := database.GetMediaByID(id)

	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)

		return
	}

	writeJSON(w, buildDetailsPayload(database, m))
}

func buildDetailsPayload(database *db.DB, m *db.Media) map[string]interface{} {
	tags, _ := database.GetTagsByMediaID(int(m.ID))
	card := mediaToCard(database, m)

	return map[string]interface{}{
		"media":    card,
		"tags":     tags,
		"similar":  fetchSimilarOrEmpty(database, m),
		"versions": findAlternateVersions(database, m),
	}
}

func findAlternateVersions(database *db.DB, m *db.Media) []map[string]interface{} {
	var rows *sql.Rows
	var err error

	if m.TmdbID > 0 {
		rows, err = database.Query(`
			SELECT MediaId, Title, Year, OriginalFilePath, CurrentFilePath, OriginalFileName, FileSizeMb
			FROM Media
			WHERE IsDeleted = 0 AND TmdbId = ? AND MediaId != ?`, m.TmdbID, m.ID)
	} else if m.CleanTitle != "" {
		rows, err = database.Query(`
			SELECT MediaId, Title, Year, OriginalFilePath, CurrentFilePath, OriginalFileName, FileSizeMb
			FROM Media
			WHERE IsDeleted = 0 AND CleanTitle = ? AND MediaId != ?`, m.CleanTitle, m.ID)
	}

	if err != nil || rows == nil {
		return []map[string]interface{}{}
	}

	defer rows.Close()

	var versions []map[string]interface{}

	for rows.Next() {
		var id int64
		var title, origPath, curPath, fileName string
		var year int
		var fileSizeMb float64

		if scanErr := rows.Scan(&id, &title, &year, &origPath, &curPath, &fileName, &fileSizeMb); scanErr == nil {
			path := curPath

			if path == "" {
				path = origPath
			}

			versions = append(versions, map[string]interface{}{
				"id":           id,
				"title":        title,
				"year":         year,
				"file_path":    path,
				"file_name":    fileName,
				"file_size_mb": fileSizeMb,
			})
		}
	}

	if versions == nil {
		return []map[string]interface{}{}
	}

	return versions
}

func fetchSimilarOrEmpty(database *db.DB, m *db.Media) interface{} {
	if m.TmdbID == 0 {
		return []interface{}{}
	}

	if results := fetchSimilarFromTMDb(database, m); results != nil {
		return results
	}

	return []interface{}{}
}
