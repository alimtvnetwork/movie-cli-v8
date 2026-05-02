// fix.go — applies repair actions for fixable findings.
//
// Repair strategy by finding ID:
//   - path-mismatch / version-drift -> updater.SelfReplace
//   - stale-worker                  -> sweep matching files (best-effort)
//   - deploy-in-path                -> print PATH-edit instructions (no auto)
package doctor

import (
	"fmt"
	"os"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
	"github.com/alimtvnetwork/movie-cli-v8/updater"
)

// Fix walks fixable findings and applies their repair action.
// Returns the number of actions taken.
func (r *Report) Fix() (int, error) {
	fmt.Println("==> movie doctor --fix")
	applied := 0
	for _, f := range r.Findings {
		if !f.IsFixable || f.Severity == SeverityOK {
			continue
		}
		if applyFix(r, f) {
			applied++
		}
	}
	fmt.Printf("  Applied %d repair action(s).\n", applied)
	return applied, nil
}

func applyFix(r *Report, f Finding) bool {
	switch f.ID {
	case idPathMismatch, idVersionDrift:
		return runSelfReplace(r)
	case idStaleWorker:
		return sweepWorkers(r.DeployDir)
	case idDeployInPath:
		return printPathInstructions(r.DeployDir)
	}
	return false
}

func runSelfReplace(r *Report) bool {
	fmt.Printf("  [FIX ] self-replace %s -> %s\n", r.Source, r.Target)
	if err := updater.SelfReplace(r.Source, r.Target); err != nil {
		fmt.Printf("  [ERR ] self-replace failed: %v\n", apperror.Wrap("self-replace", err))
		return false
	}
	return true
}

func sweepWorkers(deployDir string) bool {
	workers := findStaleWorkers(deployDir)
	if len(workers) == 0 {
		return false
	}
	cleaned := 0
	for _, w := range workers {
		if removeBestEffort(w) {
			cleaned++
		}
	}
	fmt.Printf("  [FIX ] swept %d/%d stale worker(s)\n", cleaned, len(workers))
	return cleaned > 0
}

func removeBestEffort(path string) bool {
	if err := os.Remove(path); err != nil {
		fmt.Printf("    skip (locked): %s\n", path)
		return false
	}
	fmt.Printf("    removed: %s\n", path)
	return true
}

func printPathInstructions(deployDir string) bool {
	fmt.Println("  [FIX ] PATH repair (manual — auto-edit is intentionally not done)")
	fmt.Printf("    Add this to your User PATH: %s\n", deployDir)
	fmt.Println("    PowerShell one-liner:")
	fmt.Printf(`    [Environment]::SetEnvironmentVariable("Path",`+
		` [Environment]::GetEnvironmentVariable("Path","User") + ";%s", "User")`+"\n", deployDir)
	fmt.Println("    Then open a new terminal.")
	return true
}
