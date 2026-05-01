// movie_move.go — movie move
package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
)

var moveAllFlag bool

var movieMoveCmd = &cobra.Command{
	Use:   "move [directory] | <selector> <dest>",
	Short: "Move video files (interactive) or bulk-move by selector",
	Long: `Two modes:

  1. Interactive (0–1 args): browse a directory and pick files to move.
       movie move
       movie move /Volumes/Drive
       movie move ~/Downloads --all

  2. Selector (2 args): bulk-move matches of an id/title/expression to a dest.
       movie move 42 ~/Movies
       movie move "Inception" ~/Movies
       movie move "rating < 5 AND year >= 2010" ~/Archive
       movie move -g Horror ~/Movies/Horror`,
	Args: cobra.MaximumNArgs(2),
	Run:  runMovieMove,
}

func init() {
	movieMoveCmd.Flags().BoolVar(&moveAllFlag, "all", false, "Move all video files in the directory at once")
	movieMoveCmd.Flags().StringVarP(&moveGenreSugar, "genre", "g", "", "Sugar: equivalent to selector \"g = <name>\"")
	movieMoveCmd.Flags().BoolVarP(&moveAssumeYes, "yes", "y", false, "Skip confirmation prompt (selector mode)")
}

func runMovieMove(cmd *cobra.Command, args []string) {
	if isSelectorMoveInvocation(args) {
		runSelectorMove(args)
		return
	}
	runInteractiveMoveCmd(args)
}

func runInteractiveMoveCmd(args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	scanner := bufio.NewScanner(os.Stdin)
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		errlog.Error("Cannot determine home directory: %v", homeErr)
		return
	}

	sourceDir, resolveErr := ResolveTargetDir(args, home)
	if resolveErr != nil {
		errlog.Error("Cannot resolve source directory: %v", resolveErr)
		return
	}
	fmt.Printf("📂 Source directory: %s\n", sourceDir)

	mc := MoveContext{Database: database, Scanner: scanner, Home: home}
	files, valid := validateAndListVideos(sourceDir)
	if !valid {
		return
	}

	mc.SourceDir = sourceDir
	mc.Files = files
	if moveAllFlag {
		runBatchMove(mc)
		return
	}
	runInteractiveMove(mc)
}

// validateAndListVideos checks the directory and returns its video files.
func validateAndListVideos(sourceDir string) ([]os.FileInfo, bool) {
	if !validateDirectory(sourceDir) {
		return nil, false
	}
	files, listErr := listVideoFiles(sourceDir)
	if listErr != nil {
		errlog.Error("%v", listErr)
		return nil, false
	}
	if len(files) == 0 {
		fmt.Printf("📭 No video files found in: %s\n", sourceDir)
		return nil, false
	}
	return files, true
}
