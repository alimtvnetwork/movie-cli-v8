// movie_search_json.go — JSON output for movie search
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

// searchJSONItem represents a single TMDb search result in JSON output.
type searchJSONItem struct {
	Title      string  `json:"title"`
	Year       string  `json:"year,omitempty"`
	Type       string  `json:"type"`
	Overview   string  `json:"overview,omitempty"`
	PosterPath string  `json:"poster_path,omitempty"`
	GenreIDs   []int   `json:"genre_ids,omitempty"`
	Rating     float64 `json:"rating"`
	Popularity float64 `json:"popularity"`
	Index      int     `json:"index"`
	TmdbID     int     `json:"tmdb_id"`
}

// printSearchResultsJSON outputs search results as a JSON array to stdout.
func printSearchResultsJSON(results []tmdb.SearchResult) {
	items := make([]searchJSONItem, 0, len(results))
	for i := range results {
		if i >= 15 {
			break
		}
		mediaType := results[i].MediaType
		if mediaType == "" {
			mediaType = string(db.MediaTypeMovie)
		}
		items = append(items, searchJSONItem{
			Index:      i + 1,
			Title:      results[i].GetDisplayTitle(),
			Year:       results[i].GetYear(),
			Type:       mediaType,
			TmdbID:     results[i].ID,
			Rating:     results[i].VoteAvg,
			Popularity: results[i].Popularity,
			Overview:   results[i].Overview,
			PosterPath: results[i].PosterPath,
			GenreIDs:   results[i].GenreIDs,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(items); err != nil {
		fmt.Fprintf(os.Stderr, "❌ JSON encode error: %v\n", err)
	}
}
