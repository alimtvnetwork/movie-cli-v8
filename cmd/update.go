// update.go — implements the `movie update` command.
// Uses the GitMap Self-Update Gold Standard: canonical remote installer by default,
// two-phase handoff with rename-first deploy for source rebuilds.
// See 02-spec/13-generic-cli/22-self-update-gold-standard.md.
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
	Long: `Updates movie-cli to the latest version.

By default, downloads and executes the canonical remote installer to update
the binary in place without requiring a local source checkout or git repository.

Use --source-rebuild to rebuild from local source using run.ps1.`,
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo-path")
		isSourceRebuild, _ := cmd.Flags().GetBool("source-rebuild")
		runUpdateWithDoctor(repoPath, isSourceRebuild)
	},
}

// runUpdateWithDoctor runs preflight diagnose, the update, then auto-fix
// when the preflight reported a fixable mismatch (path/version drift).
func runUpdateWithDoctor(repoPath string, isSourceRebuild bool) {
	pre := runPreflight()

	if !isSourceRebuild {
		installDir := updater.ResolveCurrentInstallDir()

		errRemote := updater.RunRemoteUpdate(installDir)
		if errRemote == nil {
			if pre != nil {
				if pre.HasFixable() {
					autoFixPostUpdate()
				}
			}

			return
		}

		fmt.Fprintln(os.Stderr, "  ⚠ Remote installer failed — falling back to source rebuild...")
	}

	exitOnUpdateError("Update failed", updater.Run(repoPath))

	if pre == nil {
		return
	}

	if !pre.HasFixable() {
		return
	}

	autoFixPostUpdate()
}

func runPreflight() *doctor.Report {
	report, err := doctor.Preflight()
	if err != nil {
		fmt.Printf("  ⚠ Preflight diagnose skipped: %v\n", err)

		return nil
	}

	return report
}

func autoFixPostUpdate() {
	fmt.Println()
	fmt.Println("  ■ Auto-running `movie doctor --fix` (preflight detected fixable issues)")

	report, err := doctor.Diagnose()
	if err != nil {
		fmt.Fprintf(os.Stderr, "auto-fix: diagnose failed: %v\n", err)

		return
	}

	if !report.HasFixable() {
		fmt.Println("  ✓ Post-update state is already clean — nothing to fix.")

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
  - Backup binaries (*.old, *.bak)`,
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
	updateCmd.Flags().Bool("source-rebuild", false, "Rebuild from local source checkout")
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
