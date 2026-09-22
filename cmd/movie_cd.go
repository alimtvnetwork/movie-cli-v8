// movie_cd.go — movie cd [target] — GitMap-style quick navigation for folders and movies.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var (
	cdAliasFlag string
	cdSetupFlag bool
	cdListFlag  bool
)

var movieCdCmd = &cobra.Command{
	Use:     "cd [folder-or-movie]",
	Aliases: []string{"go"},
	Short:   "Navigate to a scanned folder or movie directory",
	Long: `Prints the full path of a scanned root folder or movie directory so you can
navigate to it instantly in your terminal.

Features:
  • Match by folder name or derived alias (e.g., 'movies', 'tvshows')
  • Match by movie title (e.g., 'Inception' jumps to its containing folder)
  • Match by numeric index from 'movie cd' or 'movie ls'
  • Explicit alias navigation via -A / --alias
  • Shell integration via 'mcd' helper (see 'movie cd --setup')

Examples:
  movie cd                       List all available scanned root folders
  movie cd movies                Print path of the 'movies' root folder
  movie cd 1                     Print path of folder #1
  movie cd "Inception"           Print directory containing Inception
  movie cd -A movies             Jump using explicit alias
  movie cd --setup               Show shell 'mcd' function configuration`,
	Args: cobra.MaximumNArgs(1),
	Run:  runMovieCd,
}

func init() {
	movieCdCmd.Flags().StringVarP(&cdAliasFlag, "alias", "A", "", "jump by folder alias")
	movieCdCmd.Flags().BoolVar(&cdSetupFlag, "setup", false, "display shell 'mcd' configuration")
	movieCdCmd.Flags().BoolVar(&cdListFlag, "list", false, "list all scanned folders and targets")
}

func runMovieCd(cmd *cobra.Command, args []string) {
	if cdSetupFlag {
		printCdSetupInstructions()

		return
	}

	database, err := db.Open()

	if err != nil {
		errlog.Error(msgDatabaseError, err)

		return
	}

	defer database.Close()

	query := ""

	if len(args) > 0 {
		query = args[0]
	}

	if query == "list" {
		cdListFlag = true
		query = ""
	}

	executeCdResolution(database, query, cdAliasFlag, cdListFlag)
}

func executeCdResolution(database *db.DB, query, alias string, isListRequested bool) {
	if isListRequested {
		suggestions := buildAllSuggestions(database)
		printCdSuggestions(suggestions)

		return
	}

	res, suggestions, err := resolveCdTarget(database, query, alias)

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Navigation target not found: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'movie cd' or 'movie ls --folders' to see available folders.")
		os.Exit(1)
	}

	if res != nil {
		printCdResult(res)

		return
	}

	if len(suggestions) > 0 {
		printCdSuggestions(suggestions)

		return
	}

	fmt.Fprintln(os.Stderr, "📭 No scanned folders found. Run 'movie scan <folder>' first.")
}
