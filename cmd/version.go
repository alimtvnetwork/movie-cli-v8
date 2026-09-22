// version.go — implements the `movie version` command.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current version, commit, and build date",
	Long: `Display the full version information for the movie binary.

Shows the semantic version, git commit hash, build date, Go version,
architecture, data directories, and SQLite Split-DB mode.`,
	Run: runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	isColor := isColorEnabled()

	printVersionCard(isColor)
}

func printVersionCard(isColor bool) {
	exePath, _ := os.Executable()
	exePath, _ = filepath.EvalSymlinks(exePath)

	if exePath == "" {
		exePath = "movie"
	}

	dataDir := filepath.Join(filepath.Dir(exePath), "data")
	masterDB := filepath.Join(dataDir, "movie.db")
	cacheDB := filepath.Join(dataDir, "cache.db")

	fmt.Println()
	fmt.Println(colorText("  ────────────────────────────────────────────────────────────", ansiDim, isColor))
	fmt.Println(colorText("  movie-cli binary", ansiCyan, isColor))
	printMetaRow("Name:", "movie-cli", isColor)
	printMetaRow("Git URL:", "https://github.com/alimtvnetwork/movie-cli-v8", isColor)
	printMetaRow("Version:", version.Short(), isColor)
	printMetaRow("Commit SHA:", version.Commit, isColor)
	printMetaRow("Master DB:", masterDB+" (SQLite Split-DB)", isColor)
	printMetaRow("Cache DB:", cacheDB+" (WAL mode)", isColor)
	printMetaRow("Installed path:", exePath, isColor)
	printMetaRow("Architecture:", fmt.Sprintf("%s/%s (%s)", runtime.GOOS, runtime.GOARCH, runtime.Version()), isColor)
	printMetaRow("Built:", version.BuildDate, isColor)
	fmt.Println(colorText("  ────────────────────────────────────────────────────────────", ansiDim, isColor))
	fmt.Println()
}

func printMetaRow(label, val string, isColor bool) {
	bullet := colorText("●", ansiCyan, isColor)
	lbl := colorText(fmt.Sprintf("%-15s", label), ansiDim, isColor)

	fmt.Printf("  %s %s %s\n", bullet, lbl, val)
}
