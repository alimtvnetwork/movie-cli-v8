// movie_popout.go — movie popout: extract nested video files to root directory.
//
// Discovers video files inside subfolders of a target directory and moves
// them up to the root level with clean filenames. After the moves, any
// subfolder that has no remaining media is COMPACTED into <root>/.temp/
// (a non-destructive replacement for the old delete-prompt). Each move and
// each compact is tracked in move_history and action_history for full
// undo/redo support.
//
// Default behavior (no [directory] argument):
//
//	popout uses the current working directory. It will NEVER silently exit
//	just because no path was given — see mem://constraints/cwd-default-rule.
//
// Flags:
//
//	--dry-run         Preview without moving anything
//	--no-rename       Keep original filename
//	--depth N         Max subfolder depth (default 3)
//	--auto-compact    Skip the y/N prompt and compact non-media folders into
//	                  <root>/.temp/ automatically. Default: prompt with y/N.
package cmd

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/cleaner"
	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

// popoutTempDir is the destination subfolder used to hold compacted
// (non-media-bearing or originally empty) folders after popout. Kept here as
// a constant so tests and the compaction module agree.
const popoutTempDir = ".temp"

var (
	popoutDryRun      bool
	popoutNoRename    bool
	popoutAutoCompact bool
	popoutDepth       int
)

var moviePopoutCmd = &cobra.Command{
	Use:   "popout [directory]",
	Short: "Extract video files from subfolders to root directory",
	Long: `Finds video files nested inside subfolders and moves them up to the
parent directory with clean filenames. Useful for downloaded movies
that come wrapped in folders with extras, samples, and subtitles.

If no [directory] is given, the current working directory is used.

After the media files are popped out, any subfolder that no longer
contains any media (or was originally empty) is compacted into
<root>/.temp/ — kept around so undo can restore everything. Use
--auto-compact to skip the confirmation prompt.

Example:
  movie popout                    # uses current working directory
  movie popout ~/Downloads        # explicit path
  movie popout --auto-compact     # no prompts

All moves and compactions are tracked for undo support
(movie undo --batch <id>).`,
	Args: cobra.MaximumNArgs(1),
	Run:  runMoviePopout,
}

func init() {
	moviePopoutCmd.Flags().BoolVar(&popoutDryRun, "dry-run", false, "Preview only, no file moves")
	moviePopoutCmd.Flags().BoolVar(&popoutNoRename, "no-rename", false, "Keep original filename")
	moviePopoutCmd.Flags().BoolVar(&popoutAutoCompact, "auto-compact", false,
		"Skip prompt; auto-move non-media folders into <root>/.temp/")
	moviePopoutCmd.Flags().IntVar(&popoutDepth, "depth", 3, "Max subfolder depth to search")
}

// popoutItem represents a video file discovered in a subfolder.
type popoutItem struct {
	srcPath   string
	destPath  string
	cleanName string
	subDir    string
	result    cleaner.Result
	size      int64
}

// popoutFolderInfo holds info about a subfolder for the cleanup phase.
type popoutFolderInfo struct {
	name      string
	path      string
	files     []string
	totalSize int64
}

func runMoviePopout(cmd *cobra.Command, args []string) {
	database, openErr := db.Open()
	if openErr != nil {
		errlog.Error(msgDatabaseError, openErr)
		return
	}
	defer database.Close()

	scanner := bufio.NewScanner(os.Stdin)
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		errlog.Error("Cannot determine home directory: %v", homeErr)
		return
	}

	// Universal cwd-default rule. NEVER prompt silently here.
	rootDir, resolveErr := ResolveTargetDir(args, home)
	if resolveErr != nil {
		errlog.Error("Cannot resolve target directory: %v", resolveErr)
		return
	}
	fmt.Printf("📂 Target directory: %s\n", rootDir)

	if !validateDirectory(rootDir) {
		return
	}

	mc := MoveContext{Scanner: scanner, Database: database, Home: home, SourceDir: rootDir}
	runPopoutPipeline(mc, rootDir)
}

// runPopoutPipeline is the orchestration step extracted from runMoviePopout
// to keep that function under the 15-line project rule.
func runPopoutPipeline(mc MoveContext, rootDir string) {
	items := discoverNestedVideos(rootDir, popoutDepth)
	allSubdirs := discoverAllSubdirs(rootDir, popoutDepth)

	if len(items) == 0 {
		fmt.Printf("📭 No nested video files found in: %s\n", rootDir)
		// Still offer compaction for empty/non-media folders even when
		// there are no media files to pop out — that's a valid use case
		// (cleaning up a folder of leftover folders).
		offerCompactionForLeftovers(mc, rootDir, allSubdirs)
		return
	}

	printPopoutPreview(items)

	if popoutDryRun {
		fmt.Println("\n  (dry-run mode — no files moved)")
		return
	}

	executeAndCompactPopout(mc, items, rootDir, allSubdirs)
}

func validateDirectory(path string) bool {
	info, statErr := os.Stat(path)
	if statErr != nil {
		errlog.Error("Cannot access directory: %v", statErr)
		return false
	}
	if !info.IsDir() {
		errlog.Error("Path is not a directory: %s", path)
		return false
	}
	return true
}

// executeAndCompactPopout runs the move phase, then hands off to the
// compaction phase which replaces the old destructive "remove folder" flow.
func executeAndCompactPopout(mc MoveContext, items []popoutItem, rootDir string, allSubdirs []string) {
	if !confirmPopout(mc.Scanner, len(items)) {
		return
	}

	batchID := generateBatchID()
	success, failed := executePopout(mc.Database, items, batchID)
	printPopoutResult(success, failed, batchID)

	if success == 0 {
		printPopoutSummary(success, failed, 0, batchID)
		return
	}

	fmt.Println()
	cc := CleanupContext{Scanner: mc.Scanner, Database: mc.Database, BatchID: batchID}
	compacted := compactNonMediaFolders(cc, rootDir, allSubdirs, popoutAutoCompact)
	printPopoutSummary(success, failed, compacted, batchID)
}

// offerCompactionForLeftovers handles the "no media to pop out, but folders
// still exist" case. We always prompt here (no auto-compact unless flag) so
// the user isn't surprised.
func offerCompactionForLeftovers(mc MoveContext, rootDir string, allSubdirs []string) {
	if len(allSubdirs) == 0 {
		return
	}
	batchID := generateBatchID()
	cc := CleanupContext{Scanner: mc.Scanner, Database: mc.Database, BatchID: batchID}
	compacted := compactNonMediaFolders(cc, rootDir, allSubdirs, popoutAutoCompact)
	printPopoutSummary(0, 0, compacted, batchID)
}

func printPopoutPreview(items []popoutItem) {
	fmt.Printf("\n🎬 Movie Popout — %d files found in subfolders\n\n", len(items))
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i := range items {
		yearStr := formatYearSuffix(items[i].result.Year)
		fmt.Printf("\n  %d. %s%s  [%s]\n", i+1, items[i].result.CleanTitle, yearStr, humanSize(items[i].size))
		fmt.Printf("     From: %s\n", items[i].srcPath)
		fmt.Printf("     To:   %s\n", items[i].destPath)
	}
	fmt.Println("\n  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func formatYearSuffix(year int) string {
	if year <= 0 {
		return ""
	}
	return fmt.Sprintf(" (%d)", year)
}

func confirmPopout(scanner *bufio.Scanner, count int) bool {
	fmt.Printf("\n  Pop out all %d files? [y/N]: ", count)
	if !scanner.Scan() {
		return false
	}
	confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return confirm == "y" || confirm == "yes"
}

func printPopoutResult(success, failed int, batchID string) {
	fmt.Println()
	if failed == 0 {
		fmt.Printf("  ✅ All %d files popped out successfully!\n", success)
	}
	if failed > 0 {
		fmt.Printf("  ⚠️  %d moved, %d failed\n", success, failed)
	}
	fmt.Printf("  📋 Batch: %s\n", batchID[:8])
}

// generateBatchID creates a simple random hex ID for grouping related actions.
func generateBatchID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
