// movie_reset.go — CLI command to reset movie CLI data, caches, and database.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var (
	resetForce      bool
	resetYes        bool
	resetKeepConfig bool
	resetAll        bool
	resetDryRun     bool
)

var movieResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset movie database, caches, and output folders",
	Long: `Reset clears the SQLite database, cached thumbnails, JSON sidecars,
and .movie-output folders across known scan locations.

Media files are NEVER touched or deleted.`,
	Run: runMovieReset,
}

func init() {
	movieResetCmd.Flags().BoolVarP(&resetForce, "force", "f", false, "Force reset without confirmation")
	movieResetCmd.Flags().BoolVarP(&resetYes, "yes", "y", false, "Confirm reset automatically")
	movieResetCmd.Flags().BoolVar(&resetKeepConfig, "keep-config", true, "Preserve user configuration settings")
	movieResetCmd.Flags().BoolVar(&resetAll, "all", false, "Wipe global user caches in home directory as well")
	movieResetCmd.Flags().BoolVar(&resetDryRun, "dry-run", false, "Preview items that would be deleted")
}

func runMovieReset(cmd *cobra.Command, args []string) {
	opts := ResetOptions{
		IsForce:      resetForce || resetYes,
		IsKeepConfig: resetKeepConfig,
		IsAll:        resetAll,
		IsDryRun:     resetDryRun,
	}

	database, dbErr := db.Open()
	if dbErr != nil {
		errlog.Error("Database open failed during reset: %v", dbErr)
		return
	}

	targets := discoverResetTargets(database, opts)
	if opts.IsDryRun {
		printResetDryRun(targets)
		database.Close()
		return
	}

	if !opts.IsForce {
		hasConfirmed := confirmResetInteractive(targets)
		if !hasConfirmed {
			fmt.Println("Reset aborted.")
			database.Close()
			return
		}
	}

	database.Close()
	wipedCount := executeResetWipe(targets)
	reinitResetDatabase()

	printResetSummary(wipedCount, opts)
}
