// movie_cache_forget.go — `movie cache imdb forget <cleanTitle> [year]`
// subcommand. Deletes one row from the ImdbLookupCache so the next scan
// re-resolves that single title from scratch (DuckDuckGo + /find) without
// nuking the entire cache or running with --no-cache for every title.
package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var movieCacheImdbForgetCmd = &cobra.Command{
	Use:   "forget <cleanTitle> [year]",
	Short: "Delete a single cached IMDb lookup so it is re-resolved on the next scan",
	Long: `Removes the cache row matching (cleanTitle, year) from the
ImdbLookupCache. The next scan that hits this title will re-run the full
search-fallback chain (DuckDuckGo + TMDb /find) instead of reusing the
stale or wrong cached resolution.

The cleanTitle MUST match the value stored in the cache (use
'movie cache imdb list' to find the exact spelling). Year defaults to 0
(used for cache rows where the filename had no year).

Examples:
  movie cache imdb forget "Scream" 2022
  movie cache imdb forget "The Batman" 2022
  movie cache imdb forget "Some Documentary"     # year defaults to 0`,
	Args: cobra.RangeArgs(1, 2),
	Run:  runCacheImdbForget,
}

func init() {
	movieCacheImdbCmd.AddCommand(movieCacheImdbForgetCmd)
}

func runCacheImdbForget(cmd *cobra.Command, args []string) {
	cleanTitle, year, parseErr := parseForgetArgs(args)
	if parseErr != nil {
		errlog.Error("%v", parseErr)
		return
	}

	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	removed, forgetErr := database.ForgetImdbLookup(cleanTitle, year)
	if forgetErr != nil {
		errlog.Error("Failed to forget IMDb cache row: %v", forgetErr)
		return
	}
	printForgetResult(cleanTitle, year, removed)
}

func parseForgetArgs(args []string) (string, int, error) {
	cleanTitle := args[0]
	year := 0
	if len(args) == 2 {
		parsed, convErr := strconv.Atoi(args[1])
		if convErr != nil {
			return "", 0, apperror.Wrap("year must be an integer", convErr)
		}
		year = parsed
	}
	return cleanTitle, year, nil
}

func printForgetResult(cleanTitle string, year int, removed int64) {
	fmt.Println()
	if removed == 0 {
		fmt.Printf("⚠️  No cache row found for '%s' (%d). Nothing to forget.\n",
			cleanTitle, year)
		fmt.Println("   Tip: run 'movie cache imdb list' to see exact spellings.")
		fmt.Println()
		return
	}
	fmt.Printf("🗑️  Forgot cache row for '%s' (%d). It will be re-resolved on the next scan.\n",
		cleanTitle, year)
	fmt.Println()
}
