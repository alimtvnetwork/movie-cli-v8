// movie_rest_export.go — exports the current dashboard list filter results
// as CSV or JSON.
//
//	GET /api/dashboard/export?format=csv|json&<dashboard list filters>
//
// Reuses the same query parsing and filtering pipeline as
// /api/dashboard/list so the exported rows always match what the user sees.
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

const (
	exportFormatCSV  = "csv"
	exportFormatJSON = "json"
)

func handleDashboardExport(w http.ResponseWriter, r *http.Request, database *db.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cards, err := collectExportCards(r, database)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "" {
		format = exportFormatCSV
	}

	switch format {
	case exportFormatJSON:
		writeExportJSON(w, cards)
	case exportFormatCSV:
		writeExportCSV(w, cards)
	default:
		http.Error(w, "format must be csv or json", http.StatusBadRequest)
	}
}

func collectExportCards(r *http.Request, database *db.DB) ([]dashboardCard, error) {
	q := parseDashboardListQuery(r)
	q.Limit = len64Max
	q.Offset = 0

	items, err := database.ListAllMedia()
	if err != nil {
		return nil, err
	}

	cards := buildDashboardCards(database, items)
	cards = applyDashboardFilters(cards, q)
	sortDashboardCards(cards, q.Sort)
	return cards, nil
}

const len64Max = 1 << 30

func writeExportJSON(w http.ResponseWriter, cards []dashboardCard) {
	filename := exportFilename(exportFormatJSON)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"total": len(cards),
		"items": cards,
	}); err != nil {
		errlog.Error("export json encode error: %v", err)
	}
}

func writeExportCSV(w http.ResponseWriter, cards []dashboardCard) {
	filename := exportFilename(exportFormatCSV)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	cw := csv.NewWriter(w)
	defer cw.Flush()

	if err := cw.Write(exportCSVHeader()); err != nil {
		errlog.Error("export csv header error: %v", err)
		return
	}
	for i := range cards {
		if err := cw.Write(cardToCSVRow(&cards[i])); err != nil {
			errlog.Error("export csv row error: %v", err)
			return
		}
	}
}

func exportCSVHeader() []string {
	return []string{
		"id", "title", "type", "year", "runtime",
		"tmdb_id", "tmdb_rating", "genre", "director",
		"cast_list", "tags", "watched",
	}
}

func cardToCSVRow(c *dashboardCard) []string {
	return []string{
		strconv.FormatInt(c.ID, 10),
		c.Title,
		c.Type,
		strconv.Itoa(c.Year),
		strconv.Itoa(c.Runtime),
		strconv.Itoa(c.TmdbID),
		strconv.FormatFloat(c.TmdbRating, 'f', 2, 64),
		c.Genre,
		c.Director,
		c.CastList,
		strings.Join(c.Tags, "|"),
		strconv.FormatBool(c.Watched),
	}
}

func exportFilename(format string) string {
	stamp := time.Now().UTC().Format("20060102-150405")
	return fmt.Sprintf("movie-dashboard-%s.%s", stamp, format)
}
