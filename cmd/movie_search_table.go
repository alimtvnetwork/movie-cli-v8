// movie_search_table.go — table-formatted output for movie search
package cmd

import (
	"fmt"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

// printSearchResultsTable outputs TMDb search results as a formatted table.
func printSearchResultsTable(results []tmdb.SearchResult) {
	fmt.Println()
	fmt.Printf("  %-3s │ %-35s │ %-6s │ %-8s │ %-6s │ %-6s\n",
		"#", "Title", "Year", "Type", "Rating", "TMDb ID")
	fmt.Printf("  %s─┼─%s─┼─%s─┼─%s─┼─%s─┼─%s\n",
		strings.Repeat("─", 3),
		strings.Repeat("─", 35),
		strings.Repeat("─", 6),
		strings.Repeat("─", 8),
		strings.Repeat("─", 6),
		strings.Repeat("─", 6))

	for i := range results {
		if i >= 15 {
			break
		}
		title := truncate(results[i].GetDisplayTitle(), 35)
		year := results[i].GetYear()
		if year == "" {
			year = "  -   "
		}
		mediaType := db.TypeLabel(results[i].MediaType)
		rating := "   -  "
		if results[i].VoteAvg > 0 {
			rating = fmt.Sprintf("%5.1f ", results[i].VoteAvg)
		}
		fmt.Printf("  %-3d │ %-35s │ %-6s │ %-8s │ %s│ %6d\n",
			i+1, title, year, mediaType, rating, results[i].ID)
	}

	fmt.Printf("  %s─┴─%s─┴─%s─┴─%s─┴─%s─┴─%s\n",
		strings.Repeat("─", 3),
		strings.Repeat("─", 35),
		strings.Repeat("─", 6),
		strings.Repeat("─", 8),
		strings.Repeat("─", 6),
		strings.Repeat("─", 6))
	fmt.Println()
}
