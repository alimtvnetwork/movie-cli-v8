// help_formatter.go — ANSI terminal styling and custom Cobra help formatter.
package cmd

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiCyan    = "\033[1;36m"
	ansiGreen   = "\033[1;32m"
	ansiYellow  = "\033[33m"
	ansiDim     = "\033[2m"
	ansiWhite   = "\033[1;37m"
	ansiMagenta = "\033[1;35m"
)

func isColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}

	fd := os.Stdout.Fd()
	isTerm := isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)

	if !isTerm {
		return false
	}

	if !initVirtualTerminal() {
		return false
	}

	return true
}

func colorText(text, colorCode string, isColor bool) string {
	if isColor {
		return colorCode + text + ansiReset
	}

	return text
}

func formatCobraHelp(c *cobra.Command, args []string) {
	if c.Parent() == nil {
		fmt.Print(renderRootHelp(c))

		return
	}

	isColor := isColorEnabled()

	fmt.Printf("\n%s\n", colorText(c.Short, ansiCyan, isColor))

	if c.Long != "" {
		fmt.Printf("\n%s\n", c.Long)
	}

	fmt.Printf("\n%s\n", colorText("Usage:", ansiCyan, isColor))
	fmt.Printf("  %s\n", colorText(c.UseLine(), ansiGreen, isColor))

	if c.HasAvailableFlags() {
		fmt.Printf("\n%s\n", colorText("Flags:", ansiCyan, isColor))
		fmt.Print(c.Flags().FlagUsages())
	}

	renderUsageFooter(os.Stdout, isColor)
}

func setupColorfulHelp(root *cobra.Command) {
	root.SetHelpFunc(formatCobraHelp)
}
