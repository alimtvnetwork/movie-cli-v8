// movie_contextmenu_windows.go — install/remove HKCU registry keys for the
// Windows Explorer "Movie ▸ ..." right-click submenu on directories and on
// the directory background. Uses HKCU so no admin elevation is needed.
//go:build windows

package cmd

import (
	"fmt"
	"os/exec"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

const winRegRoot = `HKCU\Software\Classes\Directory\shell\Movie`
const winRegRootBg = `HKCU\Software\Classes\Directory\Background\shell\Movie`

func installContextMenu(exePath string) error {
	for _, root := range []string{winRegRoot, winRegRootBg} {
		if err := installWinSubmenuRoot(root); err != nil {
			return err
		}
	}
	for _, e := range contextMenuEntries {
		if err := installWinEntry(exePath, e); err != nil {
			return err
		}
	}
	return nil
}

func installWinSubmenuRoot(root string) error {
	if err := regAdd(root, "MUIVerb", "REG_SZ", ContextMenuLabel); err != nil {
		return err
	}
	return regAdd(root, "subcommands", "REG_SZ", "")
}

func installWinEntry(exePath string, e ContextMenuEntry) error {
	cmdLine := buildWinCommand(exePath, e)
	for _, parent := range []string{winRegRoot, winRegRootBg} {
		key := fmt.Sprintf(`%s\shell\%s`, parent, e.Key)
		if err := regAdd(key, "MUIVerb", "REG_SZ", e.Label); err != nil {
			return err
		}
		if err := regAdd(key+`\command`, "", "REG_SZ", cmdLine); err != nil {
			return err
		}
	}
	return nil
}

// buildWinCommand returns the cmd.exe invocation that the registry will run.
// %V is the clicked folder; we cd there, set MOVIE_TRIGGER, and pause on error.
func buildWinCommand(exePath string, e ContextMenuEntry) string {
	return fmt.Sprintf(
		`cmd.exe /c "set MOVIE_TRIGGER=contextmenu&& set MOVIE_CONTEXTMENU_ENTRY=%s&& cd /d "%%V" && "%s" %s . & if errorlevel 1 pause"`,
		e.Key, exePath, e.Command,
	)
}

func uninstallContextMenu() error {
	for _, root := range []string{winRegRoot, winRegRootBg} {
		_ = regDelete(root) // ignore not-found
	}
	return nil
}

func contextMenuStatus() (bool, string) {
	out, err := exec.Command("reg", "query", winRegRoot).CombinedOutput()
	if err != nil {
		return false, "registry key absent"
	}
	return true, string(out)
}

func printPostInstallHint() {
	fmt.Println("ℹ️  Right-click any folder in Explorer → \"Movie\" submenu.")
}

func regAdd(key, name, valueType, data string) error {
	args := []string{"add", key, "/f", "/t", valueType}
	if name != "" {
		args = append(args, "/v", name)
	} else {
		args = append(args, "/ve")
	}
	if data != "" {
		args = append(args, "/d", data)
	}
	if out, err := exec.Command("reg", args...).CombinedOutput(); err != nil {
		return apperror.Wrapf(err, "reg add %s: %s", key, string(out))
	}
	return nil
}

func regDelete(key string) error {
	if out, err := exec.Command("reg", "delete", key, "/f").CombinedOutput(); err != nil {
		return apperror.Wrapf(err, "reg delete %s: %s", key, string(out))
	}
	return nil
}
