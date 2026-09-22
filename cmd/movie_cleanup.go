// movie_cleanup.go — find and remove stale DB entries where files no longer exist.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var cleanupDryRun bool

var movieCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Find stale entries where files no longer exist",
	Long: `Scan the library for media entries whose file path no longer exists
on disk. By default, shows a preview (dry run). Use --remove to delete
stale entries from the database.

Examples:
  movie cleanup              # preview stale entries (dry run)
  movie cleanup --remove     # delete stale entries from DB`,
	Run: runMovieCleanup,
}

func init() {
	movieCleanupCmd.Flags().BoolVar(&cleanupDryRun, "remove", false,
		"actually delete stale entries (default is dry-run preview)")
}

func runMovieCleanup(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)

		return
	}

	defer database.Close()

	stale, err := database.FindStaleEntries(10000)
	if err != nil {
		errlog.Error("Error scanning for stale entries: %v", err)

		return
	}

	isColor := isColorEnabled()

	if len(stale) == 0 {
		fmt.Printf("  %s No stale entries found — all files exist on disk.\n\n", colorText("[ok]", ansiGreen, isColor))

		return
	}

	printStaleEntries(stale, isColor)

	if !cleanupDryRun {
		printCleanupDryRunNotice(isColor)

		return
	}

	if !confirmCleanupDelete(len(stale)) {
		return
	}

	deleted := deleteStaleEntries(database, stale)

	fmt.Printf("\n  %s Deleted %d stale entries from movie.db.\n\n", colorText("[ok]", ansiGreen, isColor), deleted)
}

func printStaleEntries(stale []db.StaleEntry, isColor bool) {
	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println()
	fmt.Printf("  %s\n", colorText(fmt.Sprintf("── Stale Database Entries (%d found) ──", len(stale)), ansiCyan, isColor))

	for i := range stale {
		m := stale[i].Media
		titleLine := colorText(fmt.Sprintf("%s (%d)", m.Title, m.Year), ansiWhite, isColor)

		fmt.Printf("  %s [ID %-4d] %s\n", bullet, m.ID, titleLine)
		fmt.Printf("     %s  %s\n", colorText("Missing:", ansiDim, isColor), stale[i].FilePath)
		fmt.Printf("     %s    %s\n", colorText("Store:", ansiDim, isColor), colorText("movie.db (Primary Library)", ansiDim, isColor))
	}

	fmt.Println()
}

func printCleanupDryRunNotice(isColor bool) {
	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println(colorText("  ── Preview Mode ──", ansiCyan, isColor))
	fmt.Printf("  %s %-12s %s\n", bullet, colorText("Status:", ansiDim, isColor), colorText("Dry run (no changes made)", ansiYellow, isColor))
	fmt.Printf("  %s %-12s %s\n\n", bullet, colorText("Action:", ansiDim, isColor), "Run 'movie cleanup --remove' to delete these entries from movie.db")
}

func confirmCleanupDelete(count int) bool {
	fmt.Printf("  ⚠️  Delete %d stale entries from movie.db? [y/N] ", count)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	return answer == "y" || answer == "yes"
}

func deleteStaleEntries(database *db.DB, stale []db.StaleEntry) int {
	deleted := 0

	for i := range stale {
		if err := database.DeleteMedia(stale[i].Media.ID); err != nil {
			errlog.Warn("Failed to delete ID %d: %v", stale[i].Media.ID, err)

			continue
		}

		deleted++
	}

	return deleted
}
