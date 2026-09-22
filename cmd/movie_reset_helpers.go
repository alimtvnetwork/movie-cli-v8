// movie_reset_helpers.go — helper routines for system reset operations.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// ResetOptions groups user choices for the reset command.
type ResetOptions struct {
	IsForce      bool `json:"is_force"`
	IsKeepConfig bool `json:"is_keep_config"`
	IsAll        bool `json:"is_all"`
	IsDryRun     bool `json:"is_dry_run"`
}

// ResetTarget describes an item targeted for deletion during reset.
type ResetTarget struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	HasTarget   bool   `json:"has_target"`
}

func discoverResetTargets(database *db.DB, opts ResetOptions) []ResetTarget {
	var targets []ResetTarget
	basePath := database.BasePath

	targets = append(targets, collectDataTargets(basePath, opts.IsKeepConfig)...)
	targets = append(targets, collectScanTargets(database)...)
	if opts.IsAll {
		targets = append(targets, collectHomeTargets(opts.IsKeepConfig)...)
	}

	return filterExistingTargets(targets)
}

func collectDataTargets(basePath string, isKeepConfig bool) []ResetTarget {
	targets := []ResetTarget{
		{Path: filepath.Join(basePath, "movie.db"), Description: "Primary SQLite database"},
		{Path: filepath.Join(basePath, "movie.db-wal"), Description: "Primary SQLite WAL file"},
		{Path: filepath.Join(basePath, "movie.db-shm"), Description: "Primary SQLite SHM file"},
		{Path: filepath.Join(basePath, "cache.db"), Description: "Ephemeral Cache database"},
		{Path: filepath.Join(basePath, "cache.db-wal"), Description: "Ephemeral Cache WAL file"},
		{Path: filepath.Join(basePath, "cache.db-shm"), Description: "Ephemeral Cache SHM file"},
		{Path: filepath.Join(basePath, "thumbnails"), Description: "Cached thumbnail images"},
		{Path: filepath.Join(basePath, "json"), Description: "JSON sidecar metadata"},
		{Path: filepath.Join(basePath, "log"), Description: "System error logs"},
	}

	if !isKeepConfig {
		targets = append(targets, ResetTarget{
			Path:        filepath.Join(basePath, "config"),
			Description: "User configuration files",
		})
	}

	return targets
}

func collectScanTargets(database *db.DB) []ResetTarget {
	var targets []ResetTarget
	seen := make(map[string]bool)

	addFolder := func(dir string) {
		if dir == "" {
			return
		}

		cleanDir := filepath.Clean(dir)
		if seen[cleanDir] {
			return
		}

		seen[cleanDir] = true

		targets = append(targets, ResetTarget{
			Path:        filepath.Join(cleanDir, ".movie-output"),
			Description: "Scan output directory",
		})
		targets = append(targets, ResetTarget{
			Path:        filepath.Join(cleanDir, ".movie"),
			Description: "Legacy scan folder",
		})
	}

	addFolder(".")
	folders, _ := database.ListDistinctScanFolders()
	for _, f := range folders {
		addFolder(f)
	}

	return targets
}

func collectHomeTargets(isKeepConfig bool) []ResetTarget {
	var targets []ResetTarget
	home, err := os.UserHomeDir()
	if err != nil {
		return targets
	}

	targets = append(targets, ResetTarget{
		Path:        filepath.Join(home, ".movie", "cache"),
		Description: "Global home cache directory",
	})
	targets = append(targets, ResetTarget{
		Path:        filepath.Join(home, ".movie-output"),
		Description: "Home scan output folder",
	})

	if !isKeepConfig {
		targets = append(targets, ResetTarget{
			Path:        filepath.Join(home, ".movie"),
			Description: "Global user movie settings",
		})
	}

	return targets
}

func filterExistingTargets(targets []ResetTarget) []ResetTarget {
	var verified []ResetTarget
	for _, t := range targets {
		_, statErr := os.Stat(t.Path)
		t.HasTarget = (statErr == nil)
		verified = append(verified, t)
	}

	return verified
}

func executeResetWipe(targets []ResetTarget) int {
	wiped := 0
	for _, t := range targets {
		if t.HasTarget {
			if rmErr := os.RemoveAll(t.Path); rmErr == nil {
				wiped++
			}
		}
	}

	return wiped
}

func reinitResetDatabase() {
	newDB, err := db.Open()
	if err == nil {
		newDB.Close()
	}
}

func confirmResetInteractive(targets []ResetTarget) bool {
	existingCount := 0

	for _, t := range targets {
		if t.HasTarget {
			existingCount++
		}
	}

	isColor := isColorEnabled()

	fmt.Println()
	fmt.Println(colorText("  ┌──────────────────────────────────────────────────────────┐", ansiYellow, isColor))
	fmt.Println(colorText("  │  ⚠️  WARNING: SYSTEM RESET                                │", ansiYellow, isColor))
	fmt.Println(colorText("  └──────────────────────────────────────────────────────────┘", ansiYellow, isColor))
	fmt.Printf("\n  System reset will permanently wipe %d targets in Split-DB stores:\n", existingCount)
	fmt.Println("    • movie.db (Primary persistent library)")
	fmt.Println("    • cache.db (Ephemeral cache store)")
	fmt.Println("    • thumbnails, logs, and sidecar metadata")
	fmt.Println("\n  Your actual media files on disk will NEVER be deleted.")
	fmt.Print("\n  Are you sure you want to proceed? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	ans, _ := reader.ReadString('\n')
	ans = strings.TrimSpace(strings.ToLower(ans))

	return ans == "y" || ans == "yes"
}

func printResetDryRun(targets []ResetTarget) {
	isColor := isColorEnabled()
	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println()
	fmt.Println(colorText("  ── System Reset — Dry Run Preview ──", ansiCyan, isColor))

	for _, t := range targets {
		status := colorText("[missing]", ansiDim, isColor)

		if t.HasTarget {
			status = colorText("[wipe]   ", ansiYellow, isColor)
		}

		fmt.Printf("  %s %s %-30s %s\n", bullet, status, colorText(t.Description+":", ansiDim, isColor), t.Path)
	}

	fmt.Printf("\n  %s Dry run complete. No files were removed.\n\n", colorText("[ok]", ansiGreen, isColor))
}

func printResetSummary(wipedCount int, opts ResetOptions) {
	isColor := isColorEnabled()

	fmt.Println()
	fmt.Printf("  %s System reset complete. Successfully wiped %d targets.\n", colorText("[ok]", ansiGreen, isColor), wipedCount)

	if opts.IsKeepConfig {
		fmt.Printf("  %s Preserved user configuration settings.\n", colorText("•", ansiCyan, isColor))
	}

	fmt.Printf("  %s SQLite Split-DB stores reinitialized with fresh schemas.\n\n", colorText("•", ansiCyan, isColor))
}
