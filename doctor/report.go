// report.go — pretty-printer for the doctor report following GitMap UI standards.
package doctor

import (
	"fmt"
	"os"
	"strings"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiGreen  = "\033[1;32m"
	ansiYellow = "\033[1;33m"
	ansiRed    = "\033[1;31m"
	ansiCyan   = "\033[1;36m"
	ansiDim    = "\033[2m"
)

func isDoctorColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}

func colorDoc(text, colorCode string, isColor bool) string {
	if isColor {
		return colorCode + text + ansiReset
	}

	return text
}

// Print writes the human-readable report to stdout.
func (r *Report) Print() {
	isColor := isDoctorColor()

	fmt.Println()
	fmt.Println(colorDoc("  ┌──────────────────────────────────────────────────────────┐", ansiDim, isColor))
	fmt.Println(colorDoc("  │   🎬 MOVIE CLI — Environment & System Diagnostics        │", ansiCyan, isColor))
	fmt.Println(colorDoc("  └──────────────────────────────────────────────────────────┘", ansiDim, isColor))
	fmt.Println()

	for _, f := range r.Findings {
		printFinding(f, isColor)
	}

	printRepoSummary(r, isColor)
	printFooter(r, isColor)
}

func printFinding(f Finding, isColor bool) {
	tag := tagFor(f.Severity, isColor)
	compName := componentNameFor(f.ID)
	compFormatted := fmt.Sprintf("%-12s", compName)

	if isColor {
		compFormatted = ansiCyan + compFormatted + ansiReset
	}

	summary := resolveFindingSummary(f)

	fmt.Printf("  %s %s %s\n", tag, compFormatted, summary)

	if f.Detail != "" {
		if f.Severity != SeverityOK {
			for _, line := range strings.Split(f.Detail, "\n") {
				trimmed := strings.TrimSpace(line)

				if trimmed != "" {
					fmt.Printf("               %s\n", trimmed)
				}
			}
		}
	}

	if f.FixHint != "" {
		if f.Severity != SeverityOK {
			hintStr := fmt.Sprintf("hint: %s", f.FixHint)

			if isColor {
				hintStr = ansiYellow + hintStr + ansiReset
			}

			fmt.Printf("               %s\n", hintStr)
		}
	}
}

func resolveFindingSummary(f Finding) string {
	if f.Detail != "" {
		if f.Severity == SeverityOK {
			lines := strings.Split(f.Detail, "\n")

			return strings.TrimSpace(lines[0])
		}
	}

	return f.Title
}

func printFooter(r *Report, isColor bool) {
	fmt.Println()

	if r.HasErrors() {
		msg := colorDoc("Result: errors found. Run `movie doctor --fix` to attempt repair.", ansiRed, isColor)
		fmt.Printf("  %s\n\n", msg)

		return
	}

	if r.HasFixable() {
		msg := colorDoc("Result: warnings found. Run `movie doctor --fix` to clean up.", ansiYellow, isColor)
		fmt.Printf("  %s\n\n", msg)

		return
	}

	msg := colorDoc("All systems nominal.", ansiGreen, isColor)
	fmt.Printf("  %s\n\n", msg)
}

func printRepoSummary(r *Report, isColor bool) {
	tag := tagFor(SeverityOK, isColor)
	isCleanRepo := r.Repo.IsCurrent && r.Repo.IsClean

	if r.Repo.IsGitRepo {
		if !isCleanRepo {
			tag = tagFor(SeverityWarn, isColor)
		}
	}

	compFormatted := fmt.Sprintf("%-12s", "repo")

	if isColor {
		compFormatted = ansiCyan + compFormatted + ansiReset
	}

	fmt.Printf("  %s %s %s\n", tag, compFormatted, r.Repo.Summary)
}

func tagFor(sev Severity, isColor bool) string {
	if !isColor {
		switch sev {
		case SeverityOK:
			return "[ok]  "
		case SeverityErr:
			return "[err] "
		default:
			return "[warn]"
		}
	}

	switch sev {
	case SeverityOK:
		return ansiGreen + "[ok]  " + ansiReset
	case SeverityErr:
		return ansiRed + "[err] " + ansiReset
	default:
		return ansiYellow + "[warn]" + ansiReset
	}
}

func componentNameFor(id string) string {
	switch id {
	case "path-mismatch":
		return "deploy"
	case "deploy-in-path":
		return "PATH"
	case "stale-worker", "stale-workers":
		return "workers"
	case "version-drift":
		return "version"
	case "config-keys":
		return "tmdb-key"
	case "source-folder":
		return "scan-dir"
	case "rest-port":
		return "rest-port"
	case "split-db:master", "split-db":
		return "db:master"
	case "split-db:cache":
		return "db:cache"
	default:
		if id != "" {
			return id
		}

		return "system"
	}
}
