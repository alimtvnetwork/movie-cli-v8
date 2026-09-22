// movie_ls_folders.go — Scanned root folder listings, shortcuts, and statistics for movie ls.
package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func runMovieLsFolders(database *db.DB) {
	stats, err := database.ListScanFoldersWithStats()

	if err != nil || len(stats) == 0 {
		fmt.Println("📭 No scanned folders found. Run 'movie scan <folder>' first.")

		return
	}

	printFullFoldersReport(stats)
}

func printFullFoldersReport(stats []db.ScanFolderStat) {
	fmt.Printf("📂 Scanned Root Folders (%d locations):\n", len(stats))
	fmt.Println("────────────────────────────────────────────────────────────────────────────")
	fmt.Printf(" %-3s  %-12s  %-8s  %-20s  %-20s\n", "#", "ALIAS", "ITEMS", "QUICK JUMP", "WEB UI SHORTCUT")

	for i, s := range stats {
		quickJump := fmt.Sprintf("mcd %s", s.Alias)
		quickUI := fmt.Sprintf("movie ui %s", s.Alias)
		fmt.Printf(" %2d.  %-12s  %-8d  %-20s  %-20s\n", i+1, s.Alias, s.ItemCount, quickJump, quickUI)
		fmt.Printf("      ↳ Path: %s\n", s.FolderPath)
	}

	fmt.Println("────────────────────────────────────────────────────────────────────────────")
	fmt.Println("💡 Shortcuts:")
	fmt.Println("  mcd <alias|#>            Change directory directly into folder")
	fmt.Println("  movie ui <alias|#>       Launch Web UI scoped to that folder")
	fmt.Println("  movie cd --setup         Show shell 'mcd' helper setup")
}

func printScanFolders(database *db.DB) {
	stats, err := database.ListScanFoldersWithStats()

	if err != nil || len(stats) == 0 {
		return
	}

	fmt.Printf("  📂 Root Folders (%d):\n", len(stats))

	for i, s := range stats {
		if i >= 4 {
			fmt.Printf("     ... and %d more (run 'movie ls --folders' to see all)\n", len(stats)-4)

			break
		}

		fmt.Printf("     [%d] %-10s (%d items)  →  mcd %-10s • movie ui %s\n",
			i+1, s.Alias, s.ItemCount, s.Alias, s.Alias)
	}

	fmt.Println()
}
