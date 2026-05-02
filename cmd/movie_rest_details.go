// movie_rest_details.go — modal payload endpoint: /api/media/{id}/details.
package cmd

import (
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
		"media":   card,
		"tags":    tags,
		"similar": fetchSimilarOrEmpty(database, m),
	}
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
