// movie_rest_report.go — HTML report rendering and media CRUD handlers.
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
	"github.com/alimtvnetwork/movie-cli-v8/pkg/trashbin"
	"github.com/alimtvnetwork/movie-cli-v8/templates"
)

func handleMediaByID(w http.ResponseWriter, r *http.Request, database *db.DB) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/media/")
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		writeRestError(w, http.StatusBadRequest, "INVALID_MEDIA_ID", "invalid media id")

		return
	}

	switch r.Method {
	case http.MethodGet:
		handleMediaGet(w, database, id)
	case http.MethodDelete:
		handleMediaDelete(w, r, database, id)
	case http.MethodPatch:
		handleMediaPatch(MediaPatchRequest{Writer: w, Request: r, Database: database, ID: id})
	default:
		writeRestError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func handleMediaGet(w http.ResponseWriter, database *db.DB, id int64) {
	m, getErr := database.GetMediaByID(id)

	if getErr != nil || m == nil {
		writeRestError(w, http.StatusNotFound, "MEDIA_NOT_FOUND", "media item not found")

		return
	}

	writeJSON(w, m)
}

func handleMediaDelete(w http.ResponseWriter, r *http.Request, database *db.DB, id int64) {
	media, getErr := database.GetMediaByID(id)

	if getErr != nil || media == nil {
		writeRestError(w, http.StatusNotFound, "MEDIA_NOT_FOUND", "media item not found")

		return
	}

	isImmediate := r.URL.Query().Get("immediate") == "true" || r.URL.Query().Get("immediate") == "1"

	if isImmediate {
		filePath := media.CurrentFilePath

		if filePath == "" {
			filePath = media.OriginalFilePath
		}

		if filePath != "" {
			_ = trashbin.MoveToTrash(filePath)
		}

		_ = database.SoftDeleteMedia(id)
		removeRmSidecar(media)

		snap, _ := db.MediaToJSON(media)
		_, _ = database.InsertActionSimple(db.ActionSimpleInput{
			FileAction: db.FileActionDelete,
			MediaID:    media.ID,
			Snapshot:   snap,
			Detail:     fmt.Sprintf("Moved to trash via Web UI: %s (%d)", media.Title, media.Year),
		})

		writeJSON(w, map[string]interface{}{
			"status":  "deleted",
			"id":      id,
			"title":   media.Title,
			"message": "Moved to OS trash bin and removed from library",
		})

		return
	}

	snapBytes, _ := json.Marshal(media)
	rec := &db.StagedActionRecord{
		ActionType:    db.StagedDelete,
		MediaId:       sql.NullInt64{Int64: id, Valid: true},
		SourcePath:    media.CurrentFilePath,
		MediaSnapshot: string(snapBytes),
		Status:        db.StatusPending,
	}

	stagedID, err := database.InsertStagedAction(rec)

	if err != nil {
		writeRestError(w, http.StatusInternalServerError, "INSERT_STAGED_FAILED", err.Error())

		return
	}

	writeJSON(w, map[string]interface{}{
		"status":           "staged_delete",
		"staged_action_id": stagedID,
		"message":          "Media deletion staged. Review and click Accept All to move to trash.",
	})
}

func handleMediaPatch(req MediaPatchRequest) {
	var updates map[string]interface{}

	if decErr := json.NewDecoder(req.Request.Body).Decode(&updates); decErr != nil {
		writeRestError(req.Writer, http.StatusBadRequest, "INVALID_JSON", "malformed JSON payload")

		return
	}

	for key, val := range updates {
		applyMediaUpdate(MediaUpdateField{Database: req.Database, ID: req.ID, Key: key, Val: val})
	}

	m, _ := req.Database.GetMediaByID(req.ID)
	writeJSON(req.Writer, m)
}

func applyMediaUpdate(field MediaUpdateField) {
	switch field.Key {
	case "genre":
		if genreStr, ok := field.Val.(string); ok {
			_ = field.Database.ReplaceMediaGenres(field.ID, genreStr)
		}
	case "title", "director", "description", "tagline":
		if _, execErr := field.Database.Exec("UPDATE Media SET "+field.Key+" = ?, UpdatedAt = datetime('now') WHERE MediaId = ?", field.Val, field.ID); execErr != nil {
			errlog.Error("DB update error for media %d field %s: %v", field.ID, field.Key, execErr)
		}
	}
}

// serveHTMLReport renders the HTML report template with live data from the database.
func serveHTMLReport(w http.ResponseWriter, database *db.DB, port int) {
	tmplBytes, err := templates.FS.ReadFile("report.html")
	if err != nil {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	tmpl, parseErr := template.New("report").Parse(string(tmplBytes))
	if parseErr != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		return
	}

	items, listErr := database.ListMedia(0, 10000)
	if listErr != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	data := buildReportData(database, items, port)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		errlog.Error("template render error: %v", err)
	}
}

func buildReportData(database *db.DB, items []db.Media, port int) htmlReportData {
	movies, tv := 0, 0

	for i := range items {
		m := &items[i]

		if m.Type == string(db.MediaTypeMovie) {
			movies++
		}

		if m.Type != string(db.MediaTypeMovie) {
			tv++
		}
	}

	reportItems := buildHTMLReportItems(items)

	return htmlReportData{
		ScannedFolder: "Library",
		ScannedAt:     "Live",
		TotalFiles:    len(items),
		Movies:        movies,
		TVShows:       tv,
		Port:          port,
		Items:         reportItems,
	}
}

// logMiddleware wraps an http.Handler and logs every request via errlog.
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		errlog.Info("[REST] %s %s → %d (%s)", r.Method, r.URL.Path, lrw.statusCode, duration)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
