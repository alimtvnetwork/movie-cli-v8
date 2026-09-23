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
	cdPickFlag  bool
	cdOpenFlag  bool
	cdStartFlag bool
)

// CdExecuteParams encapsulates parameters for cd navigation.
type CdExecuteParams struct {
	Database        *db.DB
	Query           string
	Alias           string
	IsListRequested bool
	IsPickRequested bool
	IsOpenRequested bool
}

var movieCdCmd = &cobra.Command{
	Use:     "cd [folder-or-movie]",
	Aliases: []string{"go"},
	Short:   "Navigate to a scanned folder or movie directory",
	Long: `Prints the full path of a scanned root folder or movie directory so you can
navigate to it instantly in your terminal. When the shell wrapper is loaded,
your terminal shell moves to the directory automatically.

Features:
  • Match by folder name or derived alias (e.g., 'movies', 'tvshows')
  • Match by movie title (e.g., 'Inception' jumps to its containing folder)
  • Match by numeric index from 'movie cd' or 'movie ls'
  • Explicit alias navigation via -A / --alias
  • Interactive picker when multiple targets match or no args are passed
  • Open in File Explorer via -o / --open or -s / --start
  • Shell integration via 'mcd' helper (see 'movie setup')

Examples:
  movie cd                       Interactive picker of all scanned root folders
  movie cd movies                Jump to the 'movies' root folder
  movie cd movies --start        Jump and open in File Explorer
  movie cd 1                     Jump to folder #1
  movie cd "Inception"           Jump to directory containing Inception
  movie cd -A movies             Jump using explicit alias
  movie cd --pick                Force interactive target picker
  movie cd --list                List all available targets without picker
  movie setup                    Install shell navigation wrapper`,
	Args: cobra.MaximumNArgs(1),
	Run:  runMovieCd,
}

func init() {
	movieCdCmd.Flags().StringVarP(&cdAliasFlag, "alias", "A", "", "jump by folder alias")
	movieCdCmd.Flags().BoolVar(&cdSetupFlag, "setup", false, "display shell 'mcd' configuration")
	movieCdCmd.Flags().BoolVar(&cdListFlag, "list", false, "list all scanned folders and targets")
	movieCdCmd.Flags().BoolVar(&cdPickFlag, "pick", false, "force interactive target picker")
	movieCdCmd.Flags().BoolVarP(&cdOpenFlag, "open", "o", false, "open directory in system file explorer")
	movieCdCmd.Flags().BoolVarP(&cdStartFlag, "start", "s", false, "open directory in system file explorer (alias for --open)")
}

func runMovieCd(cmd *cobra.Command, args []string) {
	if cdSetupFlag {
		printCdSetupInstructions()

		return
	}

	warnIfNoWrapper()

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

	isList := cdListFlag

	if query == "list" {
		isList = true
		query = ""
	}

	isOpen := cdOpenFlag

	if cdStartFlag {
		isOpen = true
	}

	params := &CdExecuteParams{
		Database:        database,
		Query:           query,
		Alias:           cdAliasFlag,
		IsListRequested: isList,
		IsPickRequested: cdPickFlag,
		IsOpenRequested: isOpen,
	}

	executeCdResolution(params)
}

func executeCdResolution(params *CdExecuteParams) {
	if params.IsListRequested {
		if !params.IsPickRequested {
			suggestions := buildAllSuggestions(params.Database)
			printCdSuggestions(suggestions)

			return
		}
	}

	if params.IsPickRequested {
		suggestions := buildAllSuggestions(params.Database)
		handleCdSuggestions(suggestions, params.IsOpenRequested)

		return
	}

	res, suggestions, err := resolveCdTarget(params.Database, params.Query, params.Alias)

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Navigation target not found: %v\n", err)
		fmt.Fprintln(os.Stderr, "💡 Run 'movie cd --list' or 'movie ls --folders' to see available folders.")
		os.Exit(1)
	}

	if res != nil {
		printCdResult(res, params.IsOpenRequested)

		return
	}

	if len(suggestions) > 0 {
		handleCdSuggestions(suggestions, params.IsOpenRequested)

		return
	}

	fmt.Fprintln(os.Stderr, "📭 No scanned folders found. Run 'movie scan <folder>' first.")
}

func handleCdSuggestions(suggestions []CdSuggestion, isOpenRequested bool) {
	printCdSuggestions(suggestions)

	chosen := promptCdSelection(suggestions)

	if chosen != nil {
		printCdResult(&CdTargetResult{
			TargetDirectory: chosen.Path,
			MatchName:       chosen.Name,
			MatchType:       chosen.Type,
		}, isOpenRequested)
	}
}
