// movie_cache.go — `movie cache imdb` command and its subcommands (list,
// clear, clear-misses) for inspecting and invalidating the ImdbLookupCache
// without touching the SQLite file directly.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var cacheImdbListLimit int

// movieCacheCmd is the parent grouping; cache backends other than imdb may
// be added later (e.g. tmdb response cache).
var movieCacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Inspect and invalidate persistent caches",
	Long:  "Group of commands to inspect and clear caches stored in the SQLite database.",
}

var movieCacheImdbCmd = &cobra.Command{
	Use:   "imdb",
	Short: "Manage the DuckDuckGo→IMDb lookup cache",
	Long: `The IMDb lookup cache stores the result of every DuckDuckGo→IMDb id
search performed by the search-fallback chain so repeated movie scan,
movie rescan, and movie rescan-failed runs do not re-hit the web.

Hits live for 180 days, misses for 7 days.`,
	Run: runCacheImdbSummary,
}

var movieCacheImdbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached IMDb lookup entries (most recent first)",
	Run:   runCacheImdbList,
}

var movieCacheImdbClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Delete every cached IMDb lookup (hits and misses)",
	Run:   runCacheImdbClear,
}

var movieCacheImdbClearMissesCmd = &cobra.Command{
	Use:   "clear-misses",
	Short: "Delete only cached misses so failed titles are retried on next scan",
	Run:   runCacheImdbClearMisses,
}

func init() {
	movieCacheImdbListCmd.Flags().IntVarP(&cacheImdbListLimit, "limit", "n", 50,
		"max rows to display (0 = all)")
	movieCacheImdbCmd.AddCommand(
		movieCacheImdbListCmd,
		movieCacheImdbClearCmd,
		movieCacheImdbClearMissesCmd,
	)
	movieCacheCmd.AddCommand(movieCacheImdbCmd)
}

func runCacheImdbSummary(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	printCacheImdbSummary(database)
	fmt.Println("\n  Subcommands: list | clear | clear-misses | forget | backfill")
}

func runCacheImdbList(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	entries, listErr := database.ListImdbLookups(cacheImdbListLimit)
	if listErr != nil {
		errlog.Error("Failed to list IMDb cache: %v", listErr)
		return
	}
	printCacheImdbSummary(database)
	printCacheImdbEntries(entries)
}

func runCacheImdbClear(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	removed, clearErr := database.ClearImdbLookups()
	if clearErr != nil {
		errlog.Error("Failed to clear IMDb cache: %v", clearErr)
		return
	}
	fmt.Printf("\n🧹 Cleared %d IMDb cache entries (hits + misses).\n\n", removed)
}

func runCacheImdbClearMisses(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	removed, clearErr := database.ClearImdbLookupMisses()
	if clearErr != nil {
		errlog.Error("Failed to clear IMDb cache misses: %v", clearErr)
		return
	}
	fmt.Printf("\n🧹 Cleared %d cached misses; hits preserved.\n\n", removed)
}

func printCacheImdbSummary(database *db.DB) {
	total, hits, err := database.CountImdbLookups()
	if err != nil {
		errlog.Warn("could not count IMDb cache: %v", err)
		return
	}
	fmt.Println()
	fmt.Println("📦 IMDb Lookup Cache")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Total entries: %d\n", total)
	fmt.Printf("  Hits:          %d\n", hits)
	fmt.Printf("  Misses:        %d\n", total-hits)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printCacheImdbEntries(entries []db.ImdbCacheEntry) {
	if len(entries) == 0 {
		fmt.Println("\n  (no cached entries)")
		fmt.Println()
		return
	}
	fmt.Printf("\n  Showing %d entries (most recent first):\n\n", len(entries))
	for _, e := range entries {
		printCacheImdbEntry(e)
	}
	fmt.Println()
}

func printCacheImdbEntry(e db.ImdbCacheEntry) {
	icon := "❌"
	imdb := "(miss)"
	if e.IsHit {
		icon = "✅"
		imdb = e.ImdbID
	}
	tmdb := "(unresolved)"
	if e.TmdbID > 0 {
		tmdb = fmt.Sprintf("%d (%s)", e.TmdbID, e.MediaType)
	}
	fmt.Printf("  %s %s (%d)\n", icon, e.CleanTitle, e.Year)
	fmt.Printf("       IMDb:  %s\n", imdb)
	fmt.Printf("       TMDb:  %s\n", tmdb)
	fmt.Printf("       When:  %s\n\n", e.LookedUpAt)
}
