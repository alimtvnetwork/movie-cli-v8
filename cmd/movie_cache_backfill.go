// movie_cache_backfill.go — `movie cache imdb backfill` subcommand.
//
// Walks every cached HIT row whose TmdbId is 0 (legacy v2 rows from before
// migration v3, plus partial hits where the /find call never resolved) and
// re-runs TMDb /find?external_source=imdb_id for each one. On success the
// cache row is updated in place with the resolved TmdbId + MediaType so the
// next run becomes a fully warm hit and skips both DuckDuckGo AND /find.
//
// Failures are kept as partial hits (TmdbId stays 0); they will be retried
// on the next backfill or on the next normal scan that happens to hit the
// same title.
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

// pause between /find calls so we don't burst against TMDb's rate limit.
const cacheBackfillRequestDelay = 250 * time.Millisecond

var (
	cacheBackfillLimit  int
	cacheBackfillDryRun bool
)

var movieCacheImdbBackfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Resolve cached IMDb hits that don't yet have a TmdbId via TMDb /find",
	Long: `Walks every cached HIT row whose TmdbId is 0 (legacy v2 rows or
partial hits where the /find call never resolved) and re-runs TMDb
/find?external_source=imdb_id for each one. On success the cache row is
updated in place with the resolved TmdbId + MediaType so the next run
becomes a fully warm hit.

Use --limit to cap how many rows are processed per run.
Use --dry-run to print what would be resolved without writing anything.`,
	Run: runCacheImdbBackfill,
}

func init() {
	movieCacheImdbBackfillCmd.Flags().IntVarP(&cacheBackfillLimit, "limit", "n", 0,
		"max rows to process (0 = all)")
	movieCacheImdbBackfillCmd.Flags().BoolVar(&cacheBackfillDryRun, "dry-run", false,
		"print resolutions without updating the cache")
	movieCacheImdbCmd.AddCommand(movieCacheImdbBackfillCmd)
}

func runCacheImdbBackfill(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	creds := readTmdbCredentials(database)
	if !creds.HasAuth() {
		errlog.Error("TMDb is not configured. Run: movie config set tmdb_api_key YOUR_KEY")
		return
	}

	rows, listErr := database.ListImdbLookupsUnresolved()
	if listErr != nil {
		errlog.Error("Failed to list unresolved cache rows: %v", listErr)
		return
	}
	rows = capBackfillRows(rows, cacheBackfillLimit)

	if len(rows) == 0 {
		fmt.Println("\n✅ Nothing to backfill — every cached hit already has a TmdbId.")
		fmt.Println()
		return
	}

	client := tmdb.NewClientWithToken(creds.ApiKey, creds.Token)
	stats := backfillCacheRows(database, client, rows)
	printBackfillSummary(stats, len(rows))
}

func capBackfillRows(rows []db.ImdbCacheEntry, limit int) []db.ImdbCacheEntry {
	if limit > 0 && len(rows) > limit {
		return rows[:limit]
	}
	return rows
}

type backfillStats struct {
	resolved int
	failed   int
}

func backfillCacheRows(database *db.DB, client *tmdb.Client, rows []db.ImdbCacheEntry) backfillStats {
	stats := backfillStats{}
	mode := "Backfilling"
	if cacheBackfillDryRun {
		mode = "Dry-run"
	}
	fmt.Printf("\n🔄 %s %d cached IMDb hit(s) without a TmdbId...\n\n", mode, len(rows))

	for i, row := range rows {
		processBackfillRow(database, client, row, i+1, len(rows), &stats)
		if i < len(rows)-1 {
			time.Sleep(cacheBackfillRequestDelay)
		}
	}
	return stats
}

func processBackfillRow(database *db.DB, client *tmdb.Client, row db.ImdbCacheEntry,
	idx, total int, stats *backfillStats) {
	results := client.LookupByImdbId(row.ImdbID)
	if len(results) == 0 {
		fmt.Printf("  [%d/%d] ❌ %s (%d) — IMDb %s did not resolve via /find\n",
			idx, total, row.CleanTitle, row.Year, row.ImdbID)
		stats.failed++
		return
	}

	best := results[0]
	fmt.Printf("  [%d/%d] ✅ %s (%d) — IMDb %s → TMDb %d (%s)\n",
		idx, total, row.CleanTitle, row.Year, row.ImdbID, best.ID, best.MediaType)

	if cacheBackfillDryRun {
		stats.resolved++
		return
	}

	storeErr := database.SetImdbLookup(row.CleanTitle, row.Year, row.ImdbID, best.ID, best.MediaType)
	if storeErr != nil {
		errlog.Warn("could not update cache row for '%s' (%d): %v", row.CleanTitle, row.Year, storeErr)
		stats.failed++
		return
	}
	stats.resolved++
}

func printBackfillSummary(stats backfillStats, total int) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if cacheBackfillDryRun {
		fmt.Printf("  Would resolve: %d / %d\n", stats.resolved, total)
		fmt.Printf("  Would fail:    %d\n", stats.failed)
		fmt.Println("  (dry-run — cache not modified)")
	} else {
		fmt.Printf("  Resolved: %d / %d\n", stats.resolved, total)
		fmt.Printf("  Failed:   %d (still partial; will retry next backfill)\n", stats.failed)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}
