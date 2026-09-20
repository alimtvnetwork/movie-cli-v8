// movie_rest_thumb.go — thumbnail resolution, discovery, and on-demand serving for the web UI.
package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/cleaner"
	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

func resolveMediaThumbnail(database *db.DB, m *db.Media) string {
	if m == nil {
		return ""
	}

	if m.ThumbnailPath != "" {
		return normalizeExistingThumb(m.ThumbnailPath)
	}

	return discoverMissingThumb(database, m)
}

func normalizeExistingThumb(path string) string {
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	clean := strings.ReplaceAll(path, "\\", "/")

	idx := strings.LastIndex(clean, "/")
	base := clean
	if idx != -1 {
		base = clean[idx+1:]
	}

	hasThumbDir := strings.Contains(clean, "thumbnails/")
	if hasThumbDir {
		return "thumbnails/" + base
	}

	hasLeadingSlash := strings.HasPrefix(clean, "/")
	if hasLeadingSlash {
		hasSingleSlash := strings.Count(clean, "/") == 1
		if hasSingleSlash {
			return "https://image.tmdb.org/t/p/w342" + clean
		}
	}

	return "thumbnails/" + base
}

func discoverMissingThumb(database *db.DB, m *db.Media) string {
	if database == nil {
		return ""
	}

	thumbDir := filepath.Join(database.BasePath, "thumbnails")

	diskMatch := findThumbnailOnDisk(thumbDir, m)
	if diskMatch != "" {
		m.ThumbnailPath = "thumbnails/" + diskMatch
		_, _ = database.Exec("UPDATE Media SET ThumbnailPath = ? WHERE MediaId = ?", m.ThumbnailPath, m.ID)

		return m.ThumbnailPath
	}

	if m.BackdropPath != "" {
		if strings.HasPrefix(m.BackdropPath, "/") {
			return "https://image.tmdb.org/t/p/w342" + m.BackdropPath
		}
	}

	return constructTmdbThumbCandidate(database, m)
}

func findThumbnailOnDisk(thumbDir string, m *db.Media) string {
	entries, err := os.ReadDir(thumbDir)
	if err != nil {
		return ""
	}

	slugClean := cleaner.ToSlug(m.CleanTitle)
	slugTitle := cleaner.ToSlug(m.Title)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.ToLower(entry.Name())
		hasImageExt := strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpeg")

		if !hasImageExt {
			continue
		}

		if len(slugClean) > 0 {
			if strings.HasPrefix(name, slugClean) {
				return entry.Name()
			}
		}

		if len(slugTitle) > 0 {
			if strings.HasPrefix(name, slugTitle) {
				return entry.Name()
			}
		}
	}

	return ""
}

func constructTmdbThumbCandidate(database *db.DB, m *db.Media) string {
	targetTmdbID := m.TmdbID

	if targetTmdbID <= 0 {
		cachedID := lookupCachedTmdbID(database, m.CleanTitle, m.Year)
		if cachedID > 0 {
			targetTmdbID = cachedID
			m.TmdbID = cachedID
			_, _ = database.Exec("UPDATE Media SET TmdbId = ? WHERE MediaId = ?", cachedID, m.ID)
		}
	}

	if targetTmdbID <= 0 {
		return ""
	}

	slug := cleaner.ToSlug(m.CleanTitle)
	if len(slug) == 0 {
		slug = cleaner.ToSlug(m.Title)
	}

	if len(slug) == 0 {
		slug = "media"
	}

	var fileName string
	if m.Year > 0 {
		fileName = slug + "-" + strconv.Itoa(m.Year) + "-" + strconv.Itoa(targetTmdbID) + ".jpg"
	}

	if m.Year <= 0 {
		fileName = slug + "-" + strconv.Itoa(targetTmdbID) + ".jpg"
	}

	return "thumbnails/" + fileName
}

func lookupCachedTmdbID(database *db.DB, cleanTitle string, year int) int {
	if database == nil {
		return 0
	}

	var tmdbID int
	query := "SELECT TmdbId FROM ImdbLookupCache WHERE lower(CleanTitle) = lower(?) AND Year = ? AND IsHit = 1 AND TmdbId > 0 LIMIT 1"
	err := database.QueryRow(query, cleanTitle, year).Scan(&tmdbID)
	if err != nil {
		return 0
	}

	return tmdbID
}

func handleThumbnails(w http.ResponseWriter, r *http.Request, database *db.DB) {
	if database == nil {
		http.NotFound(w, r)

		return
	}

	idx := strings.Index(r.URL.Path, "/thumbnails/")
	if idx == -1 {
		http.NotFound(w, r)

		return
	}

	rawFile := r.URL.Path[idx+len("/thumbnails/"):]
	fileName := filepath.Base(rawFile)

	hasValidName := len(fileName) > 0 && fileName != "." && fileName != "/"
	if !hasValidName {
		http.NotFound(w, r)

		return
	}

	thumbDir := filepath.Join(database.BasePath, "thumbnails")
	filePath := filepath.Join(thumbDir, fileName)

	_, statErr := os.Stat(filePath)
	if statErr == nil {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFile(w, r, filePath)

		return
	}

	hasFetched := tryFetchThumbnailOnDemand(database, thumbDir, fileName)
	if hasFetched {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFile(w, r, filePath)

		return
	}

	http.NotFound(w, r)
}

func tryFetchThumbnailOnDemand(database *db.DB, thumbDir string, fileName string) bool {
	if database == nil {
		return false
	}

	tmdbID := extractTmdbIDFromFilename(fileName)
	if tmdbID <= 0 {
		return false
	}

	apiKey, _ := database.GetConfig("TmdbApiKey")
	token, _ := database.GetConfig("TmdbToken")

	hasKey := len(apiKey) > 0 || len(token) > 0
	if !hasKey {
		return false
	}

	client := tmdb.NewClientWithToken(apiKey, token)
	imgs, err := client.GetMovieImages(tmdbID)
	if err != nil {
		return false
	}

	if imgs == nil {
		return false
	}

	hasPoster := len(imgs.Posters) > 0
	if !hasPoster {
		return false
	}

	posterPath := imgs.Posters[0].FilePath
	hasPosterPath := len(posterPath) > 0
	if !hasPosterPath {
		return false
	}

	_ = os.MkdirAll(thumbDir, 0755)
	destPath := filepath.Join(thumbDir, fileName)

	dlErr := client.DownloadPoster(posterPath, destPath)
	if dlErr != nil {
		return false
	}

	_, _ = database.Exec("UPDATE Media SET ThumbnailPath = ? WHERE TmdbId = ?", "thumbnails/"+fileName, tmdbID)

	return true
}

func extractTmdbIDFromFilename(fileName string) int {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	lastDash := strings.LastIndex(base, "-")
	if lastDash == -1 {
		return 0
	}

	idStr := base[lastDash+1:]
	tmdbID, err := strconv.Atoi(idStr)
	if err != nil {
		return 0
	}

	return tmdbID
}
