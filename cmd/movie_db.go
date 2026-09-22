// movie_db.go — movie db: print resolved database path for debugging
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var movieDBCmd = &cobra.Command{
	Use:   "db",
	Short: "Show the resolved database path and status",
	Long:  `Prints the full resolved path to the SQLite database and data directory. Useful for debugging data location issues.`,
	Run:   runMovieDB,
}

func init() {
	movieDBCmd.AddCommand(movieDBVersionCmd)
}

func runMovieDB(cmd *cobra.Command, args []string) {
	database, err := db.Open()

	if err != nil {
		errlog.Error(msgDatabaseError, err)

		return
	}

	defer database.Close()

	isColor := isColorEnabled()
	status := database.GetSplitDBStatus()

	printDBArchitectureHeader(isColor)
	printTierCard(status.MasterTier, isColor)
	fmt.Println()
	printTierCard(status.CacheTier, isColor)
	fmt.Println()
	printDBArchitectureFooter(status, isColor)
	printDBCountsSummary(database)
}

func printDBArchitectureHeader(isColor bool) {
	fmt.Println()
	fmt.Println(colorText("  ── Movie CLI Database Architecture ──", ansiCyan, isColor))
	fmt.Println()
}

func printDBArchitectureFooter(status db.SplitDBStatus, isColor bool) {
	divider := "  ──────────────────────────────────────────────────────────────────────────────"
	fmt.Println(colorText(divider, ansiDim, isColor))
	fmt.Printf("  Total: %d database file(s), combined size: %s on disk\n\n", status.TotalFiles, status.TotalSize)
}

func printTierCard(tier db.SplitDBTierInfo, isColor bool) {
	bullet := colorText("●", ansiCyan, isColor)
	fmt.Printf("  %s %s:\n", bullet, colorText(tier.Type, ansiBold, isColor))
	printTierProperty("Name:", tier.Name, isColor)
	printTierProperty("Location:", tier.Location, isColor)
	printTierProperty("Size:", tier.SizeFormatted, isColor)
	printTierProperty("Journal Mode:", tier.JournalMode, isColor)

	tblLabel := fmt.Sprintf("%d tables", tier.TableCount)

	if tier.TableCount == 1 {
		tblLabel = "1 table"
	}

	printTierProperty("Tables:", tblLabel, isColor)
	printTierProperty("Purpose:", tier.Purpose, isColor)
}

func printTierProperty(label, val string, isColor bool) {
	lbl := colorText(fmt.Sprintf("%-14s", label), ansiDim, isColor)

	fmt.Printf("    %s %s\n", lbl, val)
}

func printDBCountsSummary(database *db.DB) {
	total, countErr := database.CountMedia("")

	if countErr == nil {
		movies, _ := database.CountMedia(string(db.MediaTypeMovie))
		tv, _ := database.CountMedia(string(db.MediaTypeTV))

		totalLookups, hits, lookupErr := database.CountImdbLookups()

		if lookupErr == nil {
			misses := totalLookups - hits
			fmt.Printf("  ● Library records: %d total (%d movies, %d TV shows)\n", total, movies, tv)
			fmt.Printf("  ● Cached lookups:  %d total (%d hits, %d misses)\n", totalLookups, hits, misses)
		} else {
			fmt.Printf("  ● Library records: %d total (%d movies, %d TV shows)\n", total, movies, tv)
		}
	}

	fmt.Println()
}
