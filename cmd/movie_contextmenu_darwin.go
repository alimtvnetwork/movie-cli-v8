// movie_contextmenu_darwin.go — install/remove macOS Finder Quick Actions
// (Automator workflows) under ~/Library/Services/. The user must enable
// each service ONCE in System Settings → Keyboard → Services → Files & Folders.
// Auto-enabling is blocked by macOS TCC without code-signing.
//go:build darwin

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

func macServicesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Services")
}

func macWorkflowPath(key string) string {
	return filepath.Join(macServicesDir(), fmt.Sprintf("Movie - %s.workflow", key))
}

func installContextMenu(exePath string) error {
	if err := os.MkdirAll(macServicesDir(), 0o755); err != nil {
		return apperror.Wrap("mkdir Services", err)
	}
	for _, e := range contextMenuEntries {
		if err := installMacWorkflow(exePath, e); err != nil {
			return err
		}
	}
	return nil
}

func installMacWorkflow(exePath string, e ContextMenuEntry) error {
	wfDir := macWorkflowPath(e.Key)
	contentsDir := filepath.Join(wfDir, "Contents")
	if err := os.MkdirAll(contentsDir, 0o755); err != nil {
		return apperror.Wrap("mkdir workflow", err)
	}
	infoPlist := buildMacInfoPlist(e)
	if err := os.WriteFile(filepath.Join(contentsDir, "Info.plist"), []byte(infoPlist), 0o644); err != nil {
		return apperror.Wrap("write Info.plist", err)
	}
	doc := buildMacDocumentWflow(exePath, e)
	return os.WriteFile(filepath.Join(contentsDir, "document.wflow"), []byte(doc), 0o644)
}

func uninstallContextMenu() error {
	for _, e := range contextMenuEntries {
		if err := os.RemoveAll(macWorkflowPath(e.Key)); err != nil {
			return apperror.Wrap("remove workflow", err)
		}
	}
	return nil
}

func contextMenuStatus() (bool, string) {
	missing := 0
	for _, e := range contextMenuEntries {
		if _, err := os.Stat(macWorkflowPath(e.Key)); err != nil {
			missing++
		}
	}
	if missing == 0 {
		return true, "all 4 workflows present in " + macServicesDir()
	}
	return false, fmt.Sprintf("%d/4 workflows missing in %s", missing, macServicesDir())
}

func printPostInstallHint() {
	fmt.Println("ℹ️  One-time step: System Settings → Keyboard → Keyboard Shortcuts")
	fmt.Println("    → Services → Files & Folders → enable each \"Movie - ...\" entry.")
	fmt.Println("ℹ️  Destructive actions (Rescan) will prompt for typed 'y' confirmation")
	fmt.Println("    in the opened Terminal window before executing.")
}

func buildMacInfoPlist(e ContextMenuEntry) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>NSServices</key><array><dict>
<key>NSMenuItem</key><dict><key>default</key><string>Movie - %s</string></dict>
<key>NSMessage</key><string>runWorkflowAsService</string>
<key>NSSendFileTypes</key><array><string>public.folder</string></array>
</dict></array></dict></plist>`, e.Label)
}

// macConfirmPrompt returns a bash snippet that prints a confirmation banner
// and waits for the user to type "y" before running the action. Used for
// destructive entries on macOS where the Quick Action runs without any
// native confirm dialog. Non-destructive entries skip the prompt.
func macConfirmPrompt(e ContextMenuEntry) string {
	if !isDestructiveEntry(e.Key) {
		return ""
	}
	return fmt.Sprintf(
		`echo ''; echo '⚠️  About to run: movie %s in '\"$f\"; read -p 'Type y then Enter to continue (anything else cancels): ' a; [ \"$a\" = y ] || { echo 'cancelled'; exit 0; };`,
		e.Command,
	)
}

// isDestructiveEntry lists context-menu keys that mutate the library and
// therefore deserve a typed confirmation on macOS.
func isDestructiveEntry(key string) bool {
	return key == "rescan"
}

func buildMacDocumentWflow(exePath string, e ContextMenuEntry) string {
	confirm := macConfirmPrompt(e)
	script := fmt.Sprintf(
		`for f in "$@"; do osascript -e 'tell app "Terminal" to do script "export MOVIE_TRIGGER=contextmenu MOVIE_CONTEXTMENU_ENTRY=%s; cd \"'"$f"'\" && %s \"%s\" %s ."'; done`,
		e.Key, confirm, exePath, e.Command,
	)
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>AMApplicationBuild</key><string>492</string>
<key>AMApplicationVersion</key><string>2.10</string>
<key>actions</key><array><dict>
<key>action</key><dict><key>AMActionVersion</key><string>2.0.3</string>
<key>AMParameterProperties</key><dict><key>COMMAND_STRING</key><dict/><key>shell</key><dict/></dict>
<key>ActionParameters</key><dict>
<key>COMMAND_STRING</key><string>%s</string>
<key>inputMethod</key><integer>1</integer>
<key>shell</key><string>/bin/bash</string>
</dict>
<key>BundleIdentifier</key><string>com.apple.RunShellScript</string>
<key>CFBundleVersion</key><string>2.0.3</string>
</dict></dict></array></dict></plist>`, script)
}
