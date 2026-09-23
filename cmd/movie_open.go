// movie_open.go — movie open [target] — open folder or movie in OS file explorer and navigate shell.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var movieOpenCmd = &cobra.Command{
	Use:     "open [folder-or-movie]",
	Aliases: []string{"start", "explore", "o"},
	Short:   "Open a scanned folder or movie directory in system file explorer",
	Long: `Resolves a scanned media folder or movie directory, opens it in your operating
system's default file manager (File Explorer on Windows, Finder on macOS, or
xdg-open on Linux), and moves your terminal shell to that path.

Examples:
  movie open                     Open current folder in File Explorer
  movie open movie               Open the 'movies' root folder
  movie start movie              Alias for 'movie open'
  movie open 1                   Open scanned folder #1
  movie open "Inception"         Open containing directory of Inception`,
	Args: cobra.MaximumNArgs(1),
	Run:  runMovieOpen,
}

func init() {}

func runMovieOpen(cmd *cobra.Command, args []string) {
	database, err := db.Open()

	if err != nil {
		errlog.Error(msgDatabaseError, err)

		return
	}

	defer database.Close()

	query := "."

	if len(args) > 0 {
		query = args[0]
	}

	executeOpenTarget(database, query)
}

func executeOpenTarget(database *db.DB, query string) {
	if query == "." {
		cwd, _ := os.Getwd()
		openAndReport(cwd, "current directory")

		return
	}

	res, suggestions, err := resolveCdTarget(database, query, "")

	if err != nil {
		handleOpenFallback(query, err)

		return
	}

	if res != nil {
		openAndReport(res.TargetDirectory, res.MatchName)

		return
	}

	if len(suggestions) > 0 {
		handleOpenSuggestions(suggestions)

		return
	}

	fmt.Fprintln(os.Stderr, "📭 No scanned folders found. Run 'movie scan <folder>' first.")
}

func handleOpenFallback(query string, err error) {
	if _, statErr := os.Stat(query); statErr == nil {
		absPath, _ := filepath.Abs(query)
		openAndReport(absPath, query)

		return
	}

	fmt.Fprintf(os.Stderr, "❌ Target not found: %v\n", err)
	fmt.Fprintln(os.Stderr, "💡 Run 'movie cd --list' or 'movie ls --folders' to see available folders.")
	os.Exit(1)
}

func handleOpenSuggestions(suggestions []CdSuggestion) {
	printCdSuggestions(suggestions)

	chosen := promptCdSelection(suggestions)

	if chosen != nil {
		openAndReport(chosen.Path, chosen.Name)
	}
}

func openAndReport(targetPath, label string) {
	writeHandoffPath(targetPath)

	if err := openDirectoryInOS(targetPath); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️ Failed to open file manager: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "📂 Opened %s in File Explorer\n", label)
	}

	fmt.Println(targetPath)
}
