// update.go — implements the `movie update` command.
// Uses the copy-and-handoff pattern from gitmap-v2 to bypass Windows file locks.
// See spec/13-self-update-app-update/ for full architecture documentation.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/doctor"
	"github.com/alimtvnetwork/movie-cli-v8/updater"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update movie-cli to the latest version",
	Long: `Updates movie-cli by pulling latest source, rebuilding, and deploying.

The update process:
	  1. Finds the source repository (flag, saved DB path, binary dir, CWD, or sibling clone)
  2. Creates a handoff copy of the binary (bypasses Windows file locks)
	  3. Saves the resolved repo path into the local database for future updates
	  4. Runs run.ps1 to pull, rebuild, and deploy
  5. Compares version before/after and shows changelog

If no local repo is found, it clones a fresh copy next to the binary.
Run 'movie update' again after bootstrap to build.`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo-path")
		runUpdateWithDoctor(repoPath)
	},
}

// runUpdateWithDoctor runs preflight diagnose, the update, then auto-fix
// when the preflight reported a fixable mismatch (path/version drift).
func runUpdateWithDoctor(repoPath string) {
	pre := runPreflight()
	exitOnUpdateError("Update failed", updater.Run(repoPath))
	if pre == nil || !pre.HasFixable() {
		return
	}
	autoFixPostUpdate()
}

func runPreflight() *doctor.Report {
	report, err := doctor.Preflight()
	if err != nil {
		fmt.Printf("⚠ Preflight diagnose skipped: %v\n", err)
		return nil
	}
	return report
}

func autoFixPostUpdate() {
	fmt.Println()
	fmt.Println("==> Auto-running `movie doctor --fix` (preflight detected fixable issues)")
	report, err := doctor.Diagnose()
	if err != nil {
		fmt.Fprintf(os.Stderr, "auto-fix: diagnose failed: %v\n", err)
		return
	}
	if !report.HasFixable() {
		fmt.Println("  Post-update state is already clean — nothing to fix.")
		return
	}
	if _, err := report.Fix(); err != nil {
		fmt.Fprintf(os.Stderr, "auto-fix: %v\n", err)
	}
}

var updateRunnerCmd = &cobra.Command{
	Use:    "update-runner",
	Hidden: true,
	Short:  "Internal worker for update handoff",
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo-path")
		if repoPath == "" {
			fmt.Fprintln(os.Stderr, "❌ --repo-path is required for update-runner")
			os.Exit(1)
		}

		targetBinary, _ := cmd.Flags().GetString("target-binary")
		if targetBinary == "" {
			fmt.Fprintln(os.Stderr, "❌ --target-binary is required for update-runner")
			os.Exit(1)
		}

		exitOnUpdateError("Update worker failed", updater.RunWorker(repoPath, targetBinary))
	},
}

var updateCleanupCmd = &cobra.Command{
	Use:   "update-cleanup",
	Short: "Remove leftover temp files from previous updates",
	Long: `Removes temporary artifacts created during the update process:
  - Handoff binary copies (movie-update-*.exe)
  - Backup binaries (*.bak)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cleaning update artifacts...")
		skipPath, _ := cmd.Flags().GetString("skip-path")
		cleaned, err := updater.Cleanup(skipPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cleanup failed: %v\n", err)
			os.Exit(1)
		}
		if cleaned > 0 {
			fmt.Printf("Cleaned %d artifact(s)\n", cleaned)
			return
		}
		fmt.Println("No update artifacts found")
	},
}

func init() {
	updateCmd.Flags().String("repo-path", "", "Path to the source repository")
	updateRunnerCmd.Flags().String("repo-path", "", "Path to the source repository")
	updateRunnerCmd.Flags().String("target-binary", "", "Original executable path to redeploy")
	updateCleanupCmd.Flags().String("skip-path", "", "Path to skip during cleanup")
}

func exitOnUpdateError(label string, err error) {
	if err == nil {
		return
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}

	fmt.Fprintf(os.Stderr, "%s: %v\n", label, err)
	os.Exit(1)
}
