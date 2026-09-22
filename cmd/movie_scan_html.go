// movie_scan_html.go — generates report.html from the embedded template after a scan.
package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/cleaner"
	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
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

// htmlReportVersion represents an alternate file version of a movie item.
type htmlReportVersion struct {
	FilePath   string  `json:"file_path"`
	FileName   string  `json:"file_name"`
	ID         int64   `json:"id"`
	FileSizeMb float64 `json:"file_size_mb"`
	Year       int     `json:"year"`
}

// htmlReportItem represents a single media item in the HTML report.
type htmlReportItem struct {
	Title         string
	CleanTitle    string
	Type          string
	Genre         string
	Director      string
	CastList      string
	Description   string
	Tagline       string
	ThumbnailPath string
	BackdropPath  string
	FilePath      string
	FileName      string
	TrailerURL    string
	ImdbID        string
	VersionsJSON  string
	GenreList     []string
	Versions      []htmlReportVersion
	VersionCount  int
	ID            int64
	TmdbRating    float64
	ImdbRating    float64
	FileSizeMb    float64
	Year          int
	TmdbID        int
	Runtime       int
}

// writeHTMLReport generates report.html in the output directory.
func writeHTMLReport(stats ScanStats) error {
	tmplBytes, err := templates.FS.ReadFile("report.html")
	if err != nil {
		return appfault.Wrap("read template", err)
	}

	tmpl, err := template.New("report").Parse(string(tmplBytes))
	if err != nil {
		return appfault.Wrap("parse template", err)
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
		return appfault.Wrap("create file", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return appfault.Wrap("execute template", err)
	}
	return nil
}

func buildHTMLReportItems(media []db.Media) []htmlReportItem {
	itemMap := make(map[string]*htmlReportItem)
	orderedKeys := make([]string, 0, len(media))

	for i := range media {
		m := &media[i]
		key := makeMediaGroupingKey(m)

		ver := htmlReportVersion{
			ID:         m.ID,
			FilePath:   resolveMediaFilePath(m),
			FileName:   m.OriginalFileName,
			FileSizeMb: m.FileSizeMb,
			Year:       m.Year,
		}

		existing, found := itemMap[key]

		if found {
			existing.Versions = append(existing.Versions, ver)
			existing.VersionCount = len(existing.Versions)

			isExistingMissingMeta := existing.TmdbID == 0 || existing.ThumbnailPath == ""
			isNewHasMeta := m.TmdbID > 0 || m.ThumbnailPath != ""

			if isExistingMissingMeta && isNewHasMeta {
				updated := toHTMLReportItem(m)
				updated.Versions = existing.Versions
				updated.VersionCount = len(existing.Versions)
				*existing = updated
			}

			continue
		}

		item := toHTMLReportItem(m)
		item.Versions = []htmlReportVersion{ver}
		item.VersionCount = 1

		itemMap[key] = &item
		orderedKeys = append(orderedKeys, key)
	}

	result := make([]htmlReportItem, 0, len(orderedKeys))

	for _, k := range orderedKeys {
		it := *itemMap[k]

		if b, err := json.Marshal(it.Versions); err == nil {
			it.VersionsJSON = string(b)
		}

		result = append(result, it)
	}

	return result
}

func resolveMediaFilePath(m *db.Media) string {
	if m.CurrentFilePath != "" {
		return m.CurrentFilePath
	}

	return m.OriginalFilePath
}

func makeMediaGroupingKey(m *db.Media) string {
	if m.TmdbID > 0 {
		return fmt.Sprintf("tmdb:%d", m.TmdbID)
	}

	slug := cleaner.ToSlug(m.CleanTitle)

	if slug == "" {
		slug = cleaner.ToSlug(m.Title)
	}

	if slug == "" {
		return fmt.Sprintf("file:%s", m.OriginalFilePath)
	}

	return fmt.Sprintf("title:%s", slug)
}

func toHTMLReportItem(m *db.Media) htmlReportItem {
	filePath := resolveMediaFilePath(m)
	thumb := normalizeExistingThumb(m.ThumbnailPath)

	if thumb == "" && m.TmdbID > 0 {
		thumb = constructTmdbThumbCandidate(nil, m)
	}

	return htmlReportItem{
		ID:            m.ID,
		TmdbID:        m.TmdbID,
		Title:         m.Title,
		CleanTitle:    m.CleanTitle,
		Year:          m.Year,
		Type:          m.Type,
		Genre:         m.Genre,
		GenreList:     splitGenreList(m.Genre),
		Director:      m.Director,
		CastList:      m.CastList,
		Description:   m.Description,
		Tagline:       m.Tagline,
		TmdbRating:    m.TmdbRating,
		ImdbRating:    m.ImdbRating,
		Runtime:       m.Runtime,
		ThumbnailPath: thumb,
		BackdropPath:  m.BackdropPath,
		FilePath:      filePath,
		FileName:      m.OriginalFileName,
		FileSizeMb:    m.FileSizeMb,
		TrailerURL:    m.TrailerURL,
		ImdbID:        m.ImdbID,
	}
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
