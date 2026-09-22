// movie_info_table.go — table-formatted output for movie info
package cmd

import (
	"fmt"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// printMediaDetailTable outputs a media item as a formatted key-value table.
func printMediaDetailTable(m *db.Media) {
	labelWidth := 14
	valueWidth := 55
	isColor := isColorEnabled()

	fmt.Println()
	fmt.Printf("  ┌%s┬%s┐\n", strings.Repeat("─", labelWidth+2), strings.Repeat("─", valueWidth+2))

	fieldHdr := colorText(fmt.Sprintf(" %-*s ", labelWidth, "Field"), ansiCyan, isColor)
	valHdr := colorText(fmt.Sprintf(" %-*s ", valueWidth, "Value"), ansiCyan, isColor)
	fmt.Printf("  │%s│%s│\n", fieldHdr, valHdr)

	fmt.Printf("  ├%s┼%s┤\n", strings.Repeat("─", labelWidth+2), strings.Repeat("─", valueWidth+2))

	rows := buildDetailTableRows(m, valueWidth)

	for _, r := range rows {
		lbl := colorText(fmt.Sprintf(" %-*s ", labelWidth, r.label), ansiDim, isColor)
		val := fmt.Sprintf(" %-*s ", valueWidth, r.value)
		fmt.Printf("  │%s│%s│\n", lbl, val)
	}

	fmt.Printf("  └%s┴%s┘\n", strings.Repeat("─", labelWidth+2), strings.Repeat("─", valueWidth+2))
	fmt.Println()
}

// printMediaDetailCard renders an interactive GitMap card for movie info default output.
func printMediaDetailCard(m *db.Media, source string) {
	isColor := isColorEnabled()

	fmt.Println()
	printMediaCardHeader(m, isColor)
	printMediaCardOverview(m, isColor)
	printMediaCardCredits(m, isColor)
	printMediaCardStorage(m, source, isColor)
	printMediaCardSynopsis(m, isColor)
}

func printMediaCardHeader(m *db.Media, isColor bool) {
	typeBadge := colorText(fmt.Sprintf("[%s]", db.TypeLabel(m.Type)), ansiCyan, isColor)
	yearStr := ""

	if m.Year > 0 {
		yearStr = fmt.Sprintf(" (%d)", m.Year)
	}

	titleLine := fmt.Sprintf("  │  🎬 %s%s", m.Title, yearStr)
	cardTop := "  ┌──────────────────────────────────────────────────────────┐"
	cardBot := "  └──────────────────────────────────────────────────────────┘"

	fmt.Println(colorText(cardTop, ansiCyan, isColor))
	fmt.Printf("%-54s %s │\n", titleLine, typeBadge)
	fmt.Println(colorText(cardBot, ansiCyan, isColor))

	if m.Tagline != "" {
		fmt.Printf("  %s\n", colorText(fmt.Sprintf("\"%s\"", m.Tagline), ansiDim, isColor))
	}

	fmt.Println()
}

func printMediaCardOverview(m *db.Media, isColor bool) {
	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println(colorText("  ── Media Overview ──", ansiCyan, isColor))
	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Title:", ansiDim, isColor), colorText(m.Title, ansiWhite, isColor))

	if m.Year > 0 {
		fmt.Printf("  %s %-16s %d\n", bullet, colorText("Release Year:", ansiDim, isColor), m.Year)
	}

	if m.Genre != "" {
		fmt.Printf("  %s %-16s %s\n", bullet, colorText("Genre:", ansiDim, isColor), m.Genre)
	}

	if m.Runtime > 0 {
		fmt.Printf("  %s %-16s %d min\n", bullet, colorText("Runtime:", ansiDim, isColor), m.Runtime)
	}

	if m.Language != "" {
		fmt.Printf("  %s %-16s %s\n", bullet, colorText("Language:", ansiDim, isColor), strings.ToUpper(m.Language))
	}

	printMediaCardRatings(m, isColor)
	fmt.Println()
}

func printMediaCardRatings(m *db.Media, isColor bool) {
	hasRatings := m.ImdbRating > 0 || m.TmdbRating > 0

	if !hasRatings {
		return
	}

	bullet := colorText("●", ansiCyan, isColor)
	var ratings []string

	if m.ImdbRating > 0 {
		ratings = append(ratings, fmt.Sprintf("⭐ %s (IMDb)", colorText(fmt.Sprintf("%.1f", m.ImdbRating), ansiYellow, isColor)))
	}

	if m.TmdbRating > 0 {
		ratings = append(ratings, fmt.Sprintf("⭐ %s (TMDb)", colorText(fmt.Sprintf("%.1f", m.TmdbRating), ansiYellow, isColor)))
	}

	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Rating:", ansiDim, isColor), strings.Join(ratings, "  "))
}

func printMediaCardCredits(m *db.Media, isColor bool) {
	hasCredits := m.Director != "" || m.CastList != ""

	if !hasCredits {
		return
	}

	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println(colorText("  ── Cast & Crew ──", ansiCyan, isColor))

	if m.Director != "" {
		fmt.Printf("  %s %-16s %s\n", bullet, colorText("Director:", ansiDim, isColor), m.Director)
	}

	if m.CastList != "" {
		cast := truncate(m.CastList, 55)

		fmt.Printf("  %s %-16s %s\n", bullet, colorText("Cast:", ansiDim, isColor), cast)
	}

	fmt.Println()
}

func printMediaCardStorage(m *db.Media, source string, isColor bool) {
	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println(colorText("  ── Storage & Split-DB ──", ansiCyan, isColor))

	sourceChip := colorText("[local]", ansiGreen, isColor)
	sourceDesc := "Found in library"

	if source == "tmdb" {
		sourceChip = colorText("[tmdb]", ansiYellow, isColor)
		sourceDesc = "Hydrated from TMDb API"
	}

	fmt.Printf("  %s %-16s %s %s\n", bullet, colorText("Source:", ansiDim, isColor), sourceChip, sourceDesc)
	fmt.Printf("  %s %-16s %s (%s)\n", bullet, colorText("Split-DB Store:", ansiDim, isColor),
		colorText("movie.db", ansiWhite, isColor),
		colorText("Primary Library", ansiDim, isColor))

	if m.CurrentFilePath != "" {
		fmt.Printf("  %s %-16s %s\n", bullet, colorText("File Location:", ansiDim, isColor), m.CurrentFilePath)
	}

	if m.FileSizeMb > 0 {
		fmt.Printf("  %s %-16s %s\n", bullet, colorText("File Size:", ansiDim, isColor), db.HumanSize(m.FileSizeMb))
	}

	fmt.Println()
}

func printMediaCardSynopsis(m *db.Media, isColor bool) {
	hasSynopsis := m.Description != "" || m.TrailerURL != ""

	if !hasSynopsis {
		return
	}

	bullet := colorText("●", ansiCyan, isColor)

	if m.Description != "" {
		fmt.Println(colorText("  ── Synopsis ──", ansiCyan, isColor))
		fmt.Printf("  %s\n\n", m.Description)
	}

	if m.TrailerURL != "" {
		fmt.Printf("  %s %-16s %s\n\n", bullet, colorText("Trailer:", ansiDim, isColor), m.TrailerURL)
	}
}

type detailRow struct {
	label string
	value string
}

func buildDetailTableRows(m *db.Media, maxWidth int) []detailRow {
	rows := []detailRow{
		{"Title", m.Title},
		{"Year", fmt.Sprintf("%d", m.Year)},
		{"Type", db.TypeLabel(m.Type)},
	}

	if m.TmdbID > 0 {
		rows = append(rows, detailRow{"TMDb ID", fmt.Sprintf("%d", m.TmdbID)})
	}

	if m.ImdbID != "" {
		rows = append(rows, detailRow{"IMDb ID", m.ImdbID})
	}

	rows = append(rows, detailRow{"Rating", formatRating(m.TmdbRating, m.ImdbRating)})
	rows = append(rows, detailRow{"Store Tier", "movie.db (Primary Library)"})

	optionalFields := []struct {
		label string
		value string
		show  bool
	}{
		{"Genre", m.Genre, m.Genre != ""},
		{"Director", truncate(m.Director, maxWidth), m.Director != ""},
		{"Cast", truncate(m.CastList, maxWidth), m.CastList != ""},
		{"Runtime", fmt.Sprintf("%d min", m.Runtime), m.Runtime > 0},
		{"Language", m.Language, m.Language != ""},
		{"Tagline", truncate(m.Tagline, maxWidth), m.Tagline != ""},
		{"Trailer", m.TrailerURL, m.TrailerURL != ""},
		{"Budget", fmt.Sprintf("$%d", m.Budget), m.Budget > 0},
		{"Revenue", fmt.Sprintf("$%d", m.Revenue), m.Revenue > 0},
		{"File", truncate(m.CurrentFilePath, maxWidth), m.CurrentFilePath != ""},
		{"Description", truncate(m.Description, maxWidth), m.Description != ""},
	}

	for _, f := range optionalFields {
		if f.show {
			rows = append(rows, detailRow{f.label, f.value})
		}
	}

	return rows
}
