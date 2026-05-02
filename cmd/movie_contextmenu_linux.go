// movie_contextmenu_linux.go — install/remove a Nautilus/Nemo/Caja
// "file-manager actions" .desktop file with one Action per submenu entry.
// Works for any file manager that respects the FileManager-Actions spec.
//go:build linux

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

const linuxDesktopFileName = "movie-cli.desktop"

func linuxActionsDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "file-manager", "actions")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "file-manager", "actions")
}

func linuxDesktopPath() string {
	return filepath.Join(linuxActionsDir(), linuxDesktopFileName)
}

func installContextMenu(exePath string) error {
	dir := linuxActionsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return apperror.Wrap("mkdir actions dir", err)
	}
	content := buildLinuxDesktopFile(exePath)
	if err := os.WriteFile(linuxDesktopPath(), []byte(content), 0o644); err != nil {
		return apperror.Wrap("write .desktop file", err)
	}
	return nil
}

func uninstallContextMenu() error {
	err := os.Remove(linuxDesktopPath())
	if err != nil && !os.IsNotExist(err) {
		return apperror.Wrap("remove .desktop file", err)
	}
	return nil
}

func contextMenuStatus() (bool, string) {
	p := linuxDesktopPath()
	if _, err := os.Stat(p); err != nil {
		return false, "not found: " + p
	}
	return true, "installed: " + p
}

func printPostInstallHint() {
	fmt.Println("ℹ️  Restart Nautilus/Nemo (e.g. `nautilus -q`) to load the new menu.")
}

// buildLinuxDesktopFile returns the .desktop file body wiring all entries.
func buildLinuxDesktopFile(exePath string) string {
	var b strings.Builder
	writeLinuxHeader(&b)
	for _, e := range contextMenuEntries {
		writeLinuxAction(&b, exePath, e)
	}
	return b.String()
}

func writeLinuxHeader(b *strings.Builder) {
	keys := make([]string, 0, len(contextMenuEntries))
	for _, e := range contextMenuEntries {
		keys = append(keys, e.Key)
	}
	fmt.Fprintf(b, "[Desktop Entry]\nType=Menu\nName=%s\nIcon=video-x-generic\n", ContextMenuLabel)
	fmt.Fprintf(b, "ItemsList=%s\n\n", strings.Join(keys, ";"))
}

func writeLinuxAction(b *strings.Builder, exePath string, e ContextMenuEntry) {
	cmdLine := fmt.Sprintf(
		`bash -c 'export MOVIE_TRIGGER=contextmenu MOVIE_CONTEXTMENU_ENTRY=%s; cd "%%f" && "%s" %s . || (echo Press enter to close; read)'`,
		e.Key, exePath, e.Command,
	)
	fmt.Fprintf(b, "[X-Action-Profile %s]\nName=%s\nExec=%s\nMimeTypes=inode/directory\n\n",
		e.Key, e.Label, cmdLine)
}
