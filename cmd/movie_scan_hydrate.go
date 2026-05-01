// movie_scan_hydrate.go — Phase 5 hydration: JSON sidecars → DB Media rows.
//
// No TMDb calls. Used when the DB is empty for a scanned folder but JSON
// sidecars exist (typical: user copied .movie-output across machines).
package cmd

import (
	"database/sql"
	"encoding/json"
	"os"

	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
)

func loadJsonItems(paths []string) []scanMediaJSON {
	out := make([]scanMediaJSON, 0, len(paths))
	for _, p := range paths {
		item, ok := readSidecarFile(p)
		if !ok {
			continue
		}
		out = append(out, item)
	}
	return out
}

func readSidecarFile(path string) (scanMediaJSON, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		errlog.Warn("recon: read sidecar %s: %v", path, err)
		return scanMediaJSON{}, false
	}
	var item scanMediaJSON
	if jsonErr := json.Unmarshal(raw, &item); jsonErr != nil {
		errlog.Warn("recon: parse sidecar %s: %v", path, jsonErr)
		return scanMediaJSON{}, false
	}
	if item.Title == "" || item.Type == "" {
		errlog.Warn("recon: sidecar %s missing required fields", path)
		return scanMediaJSON{}, false
	}
	return item, true
}

func hydrateAllFromJson(database *db.DB, items []scanMediaJSON) int {
	count := 0
	for i := range items {
		if hydrateOne(database, &items[i]) {
			count++
		}
	}
	return count
}

func hydrateOne(database *db.DB, item *scanMediaJSON) bool {
	media := jsonItemToMedia(item)
	id, err := database.InsertMedia(media)
	if err != nil {
		errlog.Warn("recon: hydrate %s: %v", item.Title, err)
		return false
	}
	_, _ = database.InsertReconciliation(db.ReconInput{
		MediaID:    sql.NullInt64{Int64: id, Valid: true},
		ActionType: db.ReconActionHydratedFromJson,
		Details:    item.Title,
	})
	return true
}

func jsonItemToMedia(item *scanMediaJSON) *db.Media {
	return &db.Media{
		Title:            item.Title,
		CleanTitle:       item.CleanTitle,
		Type:             item.Type,
		ImdbID:           item.ImdbID,
		Description:      item.Description,
		Director:         item.Director,
		ThumbnailPath:    item.ThumbnailPath,
		OriginalFileName: item.OriginalFileName,
		OriginalFilePath: item.OriginalFilePath,
		CurrentFilePath:  item.CurrentFilePath,
		FileExtension:    item.FileExtension,
		ImdbRating:       item.ImdbRating,
		TmdbRating:       item.TmdbRating,
		Popularity:       item.Popularity,
		FileSize:         item.FileSize,
		Year:             item.Year,
		TmdbID:           item.TmdbID,
	}
}
