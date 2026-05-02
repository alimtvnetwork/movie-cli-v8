// movie_rescan_failed.go — re-runs TMDb lookup ONLY for rows still missing TmdbId,
// using the new SearchWithFallback chain (progressive trim → DuckDuckGo→IMDb→TMDb /find).
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

var rescanFailedLimit int
var rescanFailedNoCache bool
var rescanFailedKeepLogs bool

var movieRescanFailedCmd = &cobra.Command{
	Use:   "rescan-failed",
	Short: "Re-fetch TMDb metadata for rows still missing a TmdbId (uses fallback chain)",
	Long: `Selects every Media row where TmdbId is NULL or 0 and re-runs the
TMDb lookup using the full SearchWithFallback chain:

  1. SearchMulti(title + year)
  2. Progressive trim — drop the trailing word and retry
  3. DuckDuckGo → IMDb id → TMDb /find (web fallback)

Examples:
  movie rescan-failed              Re-fetch every row missing TmdbId
  movie rescan-failed --limit 50   Process at most 50 entries
  movie rescan-failed --no-cache   Bypass the IMDb cache for this run only`,
	Run: runMovieRescanFailed,
}

func init() {
	movieRescanFailedCmd.Flags().IntVar(&rescanFailedLimit, "limit", 0,
		"max number of entries to process (0 = unlimited)")
	movieRescanFailedCmd.Flags().BoolVar(&rescanFailedNoCache, "no-cache", false,
		"bypass the IMDb lookup cache for this run (forces fresh DuckDuckGo + /find)")
	movieRescanFailedCmd.Flags().BoolVar(&rescanFailedKeepLogs, "keep-logs", false,
		"keep the previous run's logs instead of wiping .movie-output/logs/ on start")
}

func runMovieRescanFailed(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	creds := resolveScanTmdbCredentials(database)
	if !creds.HasAuth() {
		fmt.Fprintln(os.Stderr, "❌ No TMDb credentials configured. Run: movie config set tmdb_api_key YOUR_KEY")
		return
	}

	initRunLogger("", "rescan-failed", rescanFailedKeepLogs)
	defer errlog.Close()

	entries, err := fetchFailedEntries(database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database query error: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("✅ Every row already has a TmdbId. Nothing to re-resolve.")
		return
	}

	client := tmdb.NewClientWithToken(creds.ApiKey, creds.Token)
	attachImdbCacheUnless(client, database, rescanFailedNoCache, "rescan-failed")
	updated, failed := processRescanEntries(database, client, entries)
	printRescanFailedResult(updated, failed, len(entries))

	if updated > 0 {
		regenerateReports(database)
	}
}

func fetchFailedEntries(database *db.DB) ([]db.Media, error) {
	entries, err := database.GetMediaWithMissingTmdbID()
	if err != nil {
		return nil, err
	}
	if rescanFailedLimit > 0 && len(entries) > rescanFailedLimit {
		return entries[:rescanFailedLimit], nil
	}
	return entries, nil
}

func printRescanFailedResult(updated, failed, total int) {
	fmt.Printf("\n📊 Rescan-Failed Complete!\n")
	fmt.Printf("   Resolved:  %d (now have TmdbId)\n", updated)
	fmt.Printf("   Still failed: %d\n", failed)
	fmt.Printf("   Total scanned: %d\n\n", total)
}
