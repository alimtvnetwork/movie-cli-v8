// movie_scan_html.go — generates report.html from the embedded template after a scan.
package cmd

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/templates"
)

const defaultRESTPort = 8086

// htmlReportData is the data passed to the HTML template.
type htmlReportData struct {
	ScannedFolder string
	ScannedAt     string
	Items         []htmlReportItem
	TotalFiles    int
	Movies        int
	TVShows       int
	Skipped       int
	Port          int
}

// htmlReportItem represents a single media item in the HTML report.
type htmlReportItem struct {
	Title         string
	Type          string
	Genre         string
	Director      string
	CastList      string
	Description   string
	Tagline       string
	ThumbnailPath string
	GenreList     []string
	ID            int64
	TmdbRating    float64
	ImdbRating    float64
	Year          int
	Runtime       int
}

// writeHTMLReport generates report.html in the output directory.
func writeHTMLReport(stats ScanStats) error {
	tmplBytes, err := templates.FS.ReadFile("report.html")
	if err != nil {
		return apperror.Wrap("read template", err)
	}

	tmpl, err := template.New("report").Parse(string(tmplBytes))
	if err != nil {
		return apperror.Wrap("parse template", err)
	}

	data := htmlReportData{
		ScannedFolder: stats.ScanDir,
		ScannedAt:     time.Now().Format("2006-01-02 15:04:05"),
		TotalFiles:    stats.Total,
		Movies:        stats.Movies,
		TVShows:       stats.TV,
		Skipped:       stats.Skipped,
		Port:          defaultRESTPort,
		Items:         buildHTMLReportItems(stats.Items),
	}

	outPath := filepath.Join(stats.OutputDir, "report.html")
	f, err := os.Create(outPath)
	if err != nil {
		return apperror.Wrap("create file", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return apperror.Wrap("execute template", err)
	}
	return nil
}

func buildHTMLReportItems(media []db.Media) []htmlReportItem {
	items := make([]htmlReportItem, 0, len(media))
	for i := range media {
		m := &media[i]
		items = append(items, htmlReportItem{
			ID: m.ID, Title: m.Title, Year: m.Year, Type: m.Type,
			Genre: m.Genre, GenreList: splitGenreList(m.Genre),
			Director: m.Director, CastList: m.CastList,
			Description: m.Description, Tagline: m.Tagline,
			TmdbRating: m.TmdbRating, ImdbRating: m.ImdbRating,
			Runtime: m.Runtime, ThumbnailPath: m.ThumbnailPath,
		})
	}
	return items
}

func splitGenreList(genre string) []string {
	if genre == "" {
		return nil
	}
	var genres []string
	for _, g := range strings.Split(genre, ",") {
		g = strings.TrimSpace(g)
		if g != "" {
			genres = append(genres, g)
		}
	}
	return genres
}
