// movie_contextmenu.go — `movie add-contextmenu` / `remove-contextmenu` /
// `contextmenu-status` commands.
//
// Installs OS-native right-click "Movie ▸ Scan / Rescan / Report / Stats"
// entries that open a terminal in the clicked folder and run the matching
// `movie` command. Per-OS install logic lives in build-tagged sibling files.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

// ContextMenuLabel is the parent menu label shown in the OS shell.
const ContextMenuLabel = "Movie"

// ContextMenuEntry describes one submenu item.
type ContextMenuEntry struct {
	Key     string // "scan" | "rescan" | "report" | "stats"
	Label   string // user-visible label
	Command string // movie subcommand
}

// contextMenuEntries lists the four submenu items in display order.
var contextMenuEntries = []ContextMenuEntry{
	{Key: "scan", Label: "Scan with Movie", Command: "scan"},
	{Key: "rescan", Label: "Rescan with Movie", Command: "rescan"},
	{Key: "report", Label: "Open Movie Report", Command: "scan"},
	{Key: "stats", Label: "Show Movie Stats", Command: "stats"},
}

var addContextMenuCmd = &cobra.Command{
	Use:   "add-contextmenu",
	Short: "Add Movie submenu to the OS file-manager right-click menu",
	Long: `Installs an OS-native context-menu entry (Windows registry,
Linux .desktop actions, or macOS Automator Quick Actions) that lets you
right-click a folder and run "Scan", "Rescan", "Report", or "Stats".`,
	Run: runAddContextMenu,
}

var removeContextMenuCmd = &cobra.Command{
	Use:   "remove-contextmenu",
	Short: "Remove the Movie submenu from the OS context menu",
	Run:   runRemoveContextMenu,
}

var contextMenuStatusCmd = &cobra.Command{
	Use:   "contextmenu-status",
	Short: "Show whether the Movie context menu is installed",
	Run:   runContextMenuStatus,
}

func runAddContextMenu(cmd *cobra.Command, args []string) {
	exe, err := os.Executable()
	if err != nil {
		errlog.Error("Cannot resolve own executable path: %v", err)
		return
	}
	if installErr := installContextMenu(exe); installErr != nil {
		errlog.Error("Install failed: %v", installErr)
		return
	}
	fmt.Println("✅ Movie context menu installed.")
	printPostInstallHint()
}

func runRemoveContextMenu(cmd *cobra.Command, args []string) {
	if err := uninstallContextMenu(); err != nil {
		errlog.Error("Uninstall failed: %v", err)
		return
	}
	fmt.Println("✅ Movie context menu removed.")
}

func runContextMenuStatus(cmd *cobra.Command, args []string) {
	installed, detail := contextMenuStatus()
	if installed {
		fmt.Printf("✅ Installed\n   %s\n", detail)
		return
	}
	fmt.Printf("❌ Not installed\n   %s\n", detail)
}
