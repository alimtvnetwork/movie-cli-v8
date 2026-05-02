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

	if len(stale) == 0 {
		fmt.Println("✅ No stale entries found — all files exist on disk.")
		return
	}

	printStaleEntries(stale)

	if !cleanupDryRun {
		fmt.Printf("\n📋 Dry run — no changes made. Use --remove to delete these entries.\n")
		return
	}

	if !confirmCleanupDelete(len(stale)) {
		return
	}

	deleted := deleteStaleEntries(database, stale)
	fmt.Printf("✅ Deleted %d stale entries from the database.\n", deleted)
}

func printStaleEntries(stale []db.StaleEntry) {
	fmt.Printf("🔍 Found %d stale entries (file missing from disk):\n", len(stale))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i := range stale {
		fmt.Printf("  %d. [ID %d] %s (%d)\n", i+1, stale[i].Media.ID, stale[i].Media.Title, stale[i].Media.Year)
		fmt.Printf("     Missing: %s\n", stale[i].FilePath)
	}
}

func confirmCleanupDelete(count int) bool {
	fmt.Printf("\n⚠️  Delete %d stale entries from the database? [y/N] ", count)
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
