// movie_cd_output.go — Output formatting and interactive display for movie cd.
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
)

func isTerminalOutput() bool {
	fd := os.Stdout.Fd()
	isTerm := isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)

	return isTerm
}

func isTerminalInput() bool {
	fd := os.Stdin.Fd()
	isTerm := isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)

	return isTerm
}

func printCdResult(res *CdTargetResult) {
	if res == nil {
		return
	}

	isTerm := isTerminalOutput()

	if isTerm {
		printJumpCard(res)
	}

	fmt.Println(res.TargetDirectory)
}

func printJumpCard(res *CdTargetResult) {
	fmt.Fprintln(os.Stderr, "  ╭──────────────────────────────────────────────────────────────╮")
	fmt.Fprintf(os.Stderr, "  │ 🚀 Target: %-49s │\n", truncateText(res.MatchName, 49))
	fmt.Fprintf(os.Stderr, "  │ Type:   %-52s │\n", truncateText(describeCdType(res), 52))
	fmt.Fprintf(os.Stderr, "  │ Path:   %-52s │\n", truncateText(res.TargetDirectory, 52))
	fmt.Fprintf(os.Stderr, "  │ Quick:  movie ui %-18s • movie ls             │\n", truncateText(res.MatchName, 18))
	fmt.Fprintln(os.Stderr, "  ╰──────────────────────────────────────────────────────────────╯")
}

func describeCdType(res *CdTargetResult) string {
	if res.ItemCount > 0 {
		return fmt.Sprintf("%s (%d items)", res.MatchType, res.ItemCount)
	}

	if res.MovieTitle != "" {
		return fmt.Sprintf("movie folder for %q", res.MovieTitle)
	}

	return res.MatchType
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	if maxLen <= 3 {
		return text[:maxLen]
	}

	return text[:maxLen-3] + "..."
}

func printCdSuggestions(suggestions []CdSuggestion) {
	if len(suggestions) == 0 {
		fmt.Fprintln(os.Stderr, "📭 No scanned folders or movies match your query.")
		fmt.Fprintln(os.Stderr, "💡 Run 'movie scan <folder>' to index media first.")

		return
	}

	fmt.Fprintln(os.Stderr, "📂 Available Jump Targets:")
	fmt.Fprintln(os.Stderr, "────────────────────────────────────────────────────────────────────────────")
	fmt.Fprintf(os.Stderr, " %-3s  %-12s  %-24s  %-12s  %s\n", "#", "TYPE", "NAME", "DETAIL", "PATH")

	for _, s := range suggestions {
		fmt.Fprintf(os.Stderr, " %2d.  %-12s  %-24s  %-12s  %s\n",
			s.Number, s.Type, truncateText(s.Name, 24), s.Detail, truncateText(s.Path, 35))
	}

	fmt.Fprintln(os.Stderr, "────────────────────────────────────────────────────────────────────────────")
	fmt.Fprintln(os.Stderr, "💡 Shortcuts:")
	fmt.Fprintln(os.Stderr, "  mcd <name|#|alias>       Jump to target folder directly")
	fmt.Fprintln(os.Stderr, "  movie ui <name|#|alias>  Open Web UI for target folder")
	fmt.Fprintln(os.Stderr, "  movie cd --setup         Install shell 'mcd' function")
}

func printCdSetupInstructions() {
	fmt.Println("🔧 Shell Navigation Setup ('mcd' shortcut)")
	fmt.Println("────────────────────────────────────────────────────────────────────────────")
	fmt.Println("Since CLI processes cannot change their parent shell directory directly,")
	fmt.Println("add this helper to your shell profile:")
	fmt.Println()
	fmt.Println("  PowerShell ($PROFILE):")
	fmt.Println("    function mcd { $p = (movie cd @args); if ($p) { Set-Location $p } }")
	fmt.Println()
	fmt.Println("  Bash / Zsh (~/.bashrc or ~/.zshrc):")
	fmt.Println("    mcd() { p=\"$(movie cd \"$@\")\"; [ -n \"$p\" ] && cd \"$p\"; }")
	fmt.Println()
	fmt.Println("  Fish (~/.config/fish/functions/mcd.fish):")
	fmt.Println("    function mcd; set -l p (movie cd $argv); and test -n \"$p\"; and cd $p; end")
	fmt.Println()
	fmt.Println("Examples after setup:")
	fmt.Println("  mcd movies         Jump directly to Movies folder")
	fmt.Println("  mcd tvshows        Jump directly to TV Shows folder")
	fmt.Println("  mcd 1              Jump to folder #1")
	fmt.Println("  mcd Inception      Jump to containing folder of Inception")
}

func promptCdSelection(suggestions []CdSuggestion) *CdSuggestion {
	if len(suggestions) == 0 {
		return nil
	}

	if !isTerminalInput() {
		return nil
	}

	fmt.Fprintf(os.Stderr, "Select number (1-%d) to navigate [or press Enter to cancel]: ", len(suggestions))

	var input string
	_, err := fmt.Fscanln(os.Stdin, &input)

	if err != nil {
		return nil
	}

	num, parseErr := strconv.Atoi(strings.TrimSpace(input))

	if parseErr != nil || num < 1 || num > len(suggestions) {
		return nil
	}

	selected := suggestions[num-1]

	return &selected
}
