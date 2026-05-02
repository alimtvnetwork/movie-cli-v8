// doctor.go — `movie doctor` command.
//
// Diagnostic for the updater pipeline. Surfaces deployPath/PATH mismatches,
// missing PATH entries, stale handoff workers, and version drift between
// the active binary and the deployed one — exactly the failure modes
// cataloged in spec/09-app-issues/08-updater-stale-handoff-loop-full-rca.md.
//
// Usage:
//
//	movie doctor          # diagnose only (human output)
//	movie doctor --fix    # diagnose + auto-repair fixable findings
//	movie doctor --json   # machine-readable JSON for CI/scripts
//
// Exit codes:
//
//	0 = all OK
//	2 = errors found
//	3 = fixable warnings only (no errors)
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/doctor"
)

var (
	doctorFix  bool
	doctorJson bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose updater/PATH issues; --fix to auto-repair, --json for CI",
	Long: `Surfaces the exact failure modes from the v2.97.0 → v2.121.0 stale-handoff
loop:

  - Active PATH binary differs from powershell.json deployPath
  - Deploy directory is missing from $PATH
  - Stale *-update-* handoff workers left on disk
  - Active binary version is older than the deployed one

By default, doctor only reports. Pass --fix to auto-repair: calls
self-replace for binary mismatches, sweeps stale workers, and prints
PATH-edit instructions (PATH editing is never automated).

Use --json to emit a stable machine-readable schema (movie-doctor/v1)
for CI pipelines. Exit codes: 0 = ok, 2 = errors, 3 = fixable warnings.`,
	Run: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) {
	report, err := doctor.Diagnose()
	if err != nil {
		fmt.Fprintf(os.Stderr, "doctor: %v\n", err)
		os.Exit(1)
	}
	if doctorJson {
		emitJson(report)
		return
	}
	report.Print()
	if !doctorFix {
		exitForReport(report)
	}
	if _, err := report.Fix(); err != nil {
		fmt.Fprintf(os.Stderr, "doctor --fix: %v\n", err)
		os.Exit(1)
	}
	rerunDiagnose()
}

func emitJson(report *doctor.Report) {
	if err := report.PrintJson(); err != nil {
		fmt.Fprintf(os.Stderr, "doctor --json: %v\n", err)
		os.Exit(1)
	}
	exitForReport(report)
}

func exitForReport(report *doctor.Report) {
	if report.HasErrors() {
		os.Exit(2)
	}
	if report.HasFixable() {
		os.Exit(3)
	}
	os.Exit(0)
}

func rerunDiagnose() {
	fmt.Println()
	fmt.Println("==> Re-running diagnose after fix...")
	report, err := doctor.Diagnose()
	if err != nil {
		fmt.Fprintf(os.Stderr, "doctor: %v\n", err)
		os.Exit(1)
	}
	report.Print()
	exitForReport(report)
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Auto-repair fixable findings (self-replace, sweep workers, print PATH instructions)")
	doctorCmd.Flags().BoolVar(&doctorJson, "json", false, "Emit report as JSON (schema: movie-doctor/v1) for scripting/CI")
}
