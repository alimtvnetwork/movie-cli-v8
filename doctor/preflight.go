// preflight.go — compact pre-update diagnose summary.
// Called from updater.Run() before handoff so users see PATH/deploy
// mismatches up front instead of after a full build cycle.
package doctor

import "fmt"

// Preflight runs the full diagnose pass and prints a compact warning
// banner when any non-OK finding exists. Returns the report so callers
// can decide whether to abort or continue.
func Preflight() (*Report, error) {
	report, err := Diagnose()
	if err != nil {
		return nil, err
	}
	if !hasNonOK(report) {
		fmt.Println("[ OK ] Preflight checks passed")
		return report, nil
	}
	printPreflightBanner(report)
	return report, nil
}

func printPreflightBanner(r *Report) {
	isColor := isDoctorColor()

	fmt.Println("==> Preflight checks (movie doctor)")
	fmt.Println("  --------------------------------------------------")

	for _, f := range r.Findings {
		if f.Severity == SeverityOK {
			continue
		}

		printFinding(f, isColor)
	}

	fmt.Println("  --------------------------------------------------")
	printPreflightFooter(r)
}

func printPreflightFooter(r *Report) {
	if r.HasErrors() {
		fmt.Println("  Preflight: errors detected — update will continue but may not take effect.")
		fmt.Println("  Run `movie doctor --fix` after the update to repair.")
		return
	}
	fmt.Println("  Preflight: warnings detected — update will continue.")
	fmt.Println("  Run `movie doctor --fix` after the update if issues persist.")
}

func hasNonOK(r *Report) bool {
	for _, f := range r.Findings {
		if f.Severity != SeverityOK {
			return true
		}
	}
	return false
}
