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

type detailRow struct {
	label string
	value string
}

func buildDetailTableRows(m *db.Media, maxWidth int) []detailRow {
	rows := []detailRow{
		{"Title", m.Title},
		{"Year", fmt.Sprintf("%d", m.Year)},
		{"Type", m.Type},
	}

	if m.TmdbID > 0 {
		rows = append(rows, detailRow{"TMDb ID", fmt.Sprintf("%d", m.TmdbID)})
	}
	if m.ImdbID != "" {
		rows = append(rows, detailRow{"IMDb ID", m.ImdbID})
	}

	rows = append(rows, detailRow{"Rating", formatRating(m.TmdbRating, m.ImdbRating)})

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
