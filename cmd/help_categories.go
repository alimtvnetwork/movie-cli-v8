// help_categories.go — command categorization and root help text generator.
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/version"
)

type helpCommand struct {
	Name string
	Desc string
}

type helpGroup struct {
	Title    string
	Commands []helpCommand
}

func getHelpGroups() []helpGroup {
	return []helpGroup{
		{
			Title: "── Core Media Operations ──",
			Commands: []helpCommand{
				{"movie scan [folder]", "Scan folder, match TMDb, save metadata & HTML report"},
				{"movie report [folder]", "Generate library summary report and HTML dashboard"},
				{"movie ls", "List indexed movies and TV series"},
				{"movie info <title>", "Query TMDb and inspect media metadata"},
				{"movie search <query>", "Search TMDb for movies or TV series"},
			},
		},
		{
			Title: "── Library Management & Files ──",
			Commands: []helpCommand{
				{"movie ui", "Launch web dashboard and open in browser"},
				{"movie move", "Move organized media to destination directories"},
				{"movie rename", "Batch-rename messy media filenames"},
				{"movie rm <title>", "Stage media deletion to OS trash bin"},
				{"movie undo / redo", "Undo or redo previous file operations"},
				{"movie cleanup", "Prune stale database entries for removed files"},
				{"movie popout", "Extract nested video files to root directory"},
			},
		},
		{
			Title: "── Discovery, Taxonomy & Play ──",
			Commands: []helpCommand{
				{"movie play <title>", "Play media item in default player"},
				{"movie watch", "Watch directory for incoming media files"},
				{"movie suggest", "Get recommendations from your library"},
				{"movie tag", "Manage custom tags for your collection"},
			},
		},
		{
			Title: "── Storage, REST API & Admin ──",
			Commands: []helpCommand{
				{"movie rest", "Start headless REST API service"},
				{"movie reset", "Safely wipe database, caches, and output folders"},
				{"movie config", "Configure TMDb API key and CLI preferences"},
				{"movie db", "Inspect SQLite multi-tier database statistics"},
				{"movie stats", "View comprehensive library metrics"},
				{"movie logs", "Inspect system and error logs"},
				{"movie doctor", "Diagnose environment and network dependencies"},
				{"movie update", "Update movie-cli to latest release"},
			},
		},
	}
}

func renderRootHelp(c *cobra.Command) string {
	isColor := isColorEnabled()
	var b strings.Builder

	b.WriteString(colorText(fmt.Sprintf("\nmovie-cli %s", version.Short()), ansiCyan, isColor))
	b.WriteString(" — Organize, enrich, and enjoy your movie collection\n\n")

	b.WriteString(colorText("Usage:\n", ansiCyan, isColor))
	b.WriteString("  " + colorText("movie", ansiGreen, isColor) + " [command] [flags]\n\n")

	renderHelpGroups(&b, isColor)
	renderFlagsSection(&b, isColor)
	renderBinaryCard(&b, isColor)

	return b.String()
}

func renderHelpGroups(b *strings.Builder, isColor bool) {
	for _, g := range getHelpGroups() {
		b.WriteString(colorText("  "+g.Title+"\n", ansiCyan, isColor))

		for _, cmd := range g.Commands {
			cmdStr := fmt.Sprintf("    %-26s", cmd.Name)
			b.WriteString(colorText(cmdStr, ansiGreen, isColor))
			b.WriteString(" " + colorText(cmd.Desc, ansiDim, isColor) + "\n")
		}

		b.WriteString("\n")
	}
}

func renderFlagsSection(b *strings.Builder, isColor bool) {
	b.WriteString(colorText("  ── Global Options ──\n", ansiCyan, isColor))
	b.WriteString("    " + colorText("-h, --help", ansiYellow, isColor) + "                 " + colorText("Show help for command", ansiDim, isColor) + "\n")
	b.WriteString("    " + colorText("-v, --version", ansiYellow, isColor) + "              " + colorText("Print version and exit", ansiDim, isColor) + "\n\n")
}
