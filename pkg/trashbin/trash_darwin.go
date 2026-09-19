//go:build darwin

package trashbin

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

func moveToTrashOS(absPath string) error {
	escaped := strings.ReplaceAll(absPath, "\"", "\\\"")
	script := fmt.Sprintf("tell application \"Finder\" to delete POSIX file \"%s\"", escaped)
	cmd := exec.Command("osascript", "-e", script)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return appfault.Wrapf(err, "macos trash failed: %s (output: %s)", absPath, string(out))
	}

	return nil
}
