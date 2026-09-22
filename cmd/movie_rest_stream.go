// movie_rest_stream.go — video streaming range requests and external player execution for REST API.
package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func handleVideoStream(w http.ResponseWriter, r *http.Request, database *db.DB, id int64) {
	media, err := database.GetMediaByID(id)

	if err != nil || media == nil {
		writeRestError(w, http.StatusNotFound, "MEDIA_NOT_FOUND", "media item not found")

		return
	}

	filePath := resolveMediaStreamPath(media)

	if len(filePath) == 0 {
		writeRestError(w, http.StatusNotFound, "PATH_MISSING", "media file path missing")

		return
	}

	if _, statErr := os.Stat(filePath); statErr != nil {
		writeRestError(w, http.StatusNotFound, "FILE_NOT_FOUND", "media file not found on disk")

		return
	}

	setVideoStreamContentType(w, filePath)
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeFile(w, r, filePath)
}

func resolveMediaStreamPath(m *db.Media) string {
	if len(m.CurrentFilePath) > 0 {
		return m.CurrentFilePath
	}

	return m.OriginalFilePath
}

func setVideoStreamContentType(w http.ResponseWriter, path string) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".mp4", ".m4v":
		w.Header().Set("Content-Type", "video/mp4")
	case ".mkv":
		w.Header().Set("Content-Type", "video/x-matroska")
	case ".webm":
		w.Header().Set("Content-Type", "video/webm")
	case ".avi":
		w.Header().Set("Content-Type", "video/x-msvideo")
	case ".mov":
		w.Header().Set("Content-Type", "video/quicktime")
	}
}

func handleExternalPlay(w http.ResponseWriter, r *http.Request, database *db.DB, id int64) {
	media, err := database.GetMediaByID(id)

	if err != nil || media == nil {
		writeRestError(w, http.StatusNotFound, "MEDIA_NOT_FOUND", "media item not found")

		return
	}

	filePath := resolveMediaStreamPath(media)

	if len(filePath) == 0 {
		writeRestError(w, http.StatusNotFound, "PATH_MISSING", "media file path missing")

		return
	}

	if _, statErr := os.Stat(filePath); statErr != nil {
		writeRestError(w, http.StatusNotFound, "FILE_NOT_FOUND", "media file not found on disk")

		return
	}

	launchPlayer(filePath)

	writeJSON(w, map[string]interface{}{
		"status":  "playing",
		"id":      id,
		"file":    filePath,
		"title":   media.Title,
		"message": "Launched in default media player / VLC",
	})
}
