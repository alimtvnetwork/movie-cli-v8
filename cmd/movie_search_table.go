// movie_search_table.go — table-formatted output for movie search
package cmd

import (
	"fmt"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

const (
	searchColNum    = 3
	searchColTitle  = 35
	searchColYear   = 6
	searchColType   = 8
	searchColRating = 8
	searchColID     = 8
)

// printSearchResultsTable outputs TMDb search results as a formatted table.
func printSearchResultsTable(results []tmdb.SearchResult) {
	isColor := isColorEnabled()

	fmt.Println()
	printSearchResultsHeader(isColor)
	printSearchResultsRows(results, isColor)
	printSearchResultsDivider("┴")

	count := minInt(len(results), 15)

	fmt.Printf("\n  Showing %d result(s)\n\n", count)
}

func printSearchResultsHeader(isColor bool) {
	fmt.Printf("  %s │ %s │ %s │ %s │ %s │ %s\n",
		formatSearchHeaderCell("#", searchColNum, isColor),
		formatSearchHeaderCell("Title", searchColTitle, isColor),
		formatSearchHeaderCell("Year", searchColYear, isColor),
		formatSearchHeaderCell("Type", searchColType, isColor),
		formatSearchHeaderCell("Rating", searchColRating, isColor),
		formatSearchHeaderCell("TMDb ID", searchColID, isColor))

	printSearchResultsDivider("┼")
}

func formatSearchHeaderCell(text string, width int, isColor bool) string {
	padded := fmt.Sprintf("%-*s", width, text)

	return colorText(padded, ansiCyan, isColor)
}

func printSearchResultsDivider(mid string) {
	fmt.Printf("  %s─%s─%s─%s─%s─%s─%s─%s─%s─%s─%s\n",
		strings.Repeat("─", searchColNum), mid,
		strings.Repeat("─", searchColTitle), mid,
		strings.Repeat("─", searchColYear), mid,
		strings.Repeat("─", searchColType), mid,
		strings.Repeat("─", searchColRating), mid,
		strings.Repeat("─", searchColID))
}

func printSearchResultsRows(results []tmdb.SearchResult, isColor bool) {
	maxRows := minInt(len(results), 15)

	for i := 0; i < maxRows; i++ {
		printSearchTableRow(i+1, results[i], isColor)
	}
}

func printSearchTableRow(idx int, item tmdb.SearchResult, isColor bool) {
	title := truncate(item.GetDisplayTitle(), searchColTitle)
	year := item.GetYear()

	if year == "" {
		year = "  -   "
	}

	mediaType := db.TypeLabel(item.MediaType)
	rating := "  -   "

	if item.VoteAvg > 0 {
		rating = fmt.Sprintf("⭐ %4.1f", item.VoteAvg)
	}

	fmt.Printf("  %-*d │ %-*s │ %-*s │ %-*s │ %-*s │ %*d\n",
		searchColNum, idx,
		searchColTitle, title,
		searchColYear, year,
		searchColType, mediaType,
		searchColRating, rating,
		searchColID, item.ID)
}
