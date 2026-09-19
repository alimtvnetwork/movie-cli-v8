// Package trashbin provides safe, cross-platform file and directory deletion.
// It moves items to the operating system's Recycle Bin or Trash instead of permanently unlinking them.
package trashbin

import (
	"os"
	"path/filepath"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

// MoveToTrash moves a file or directory at targetPath to the OS trash bin / Recycle Bin.
// If the native OS trash mechanism fails, it safely falls back to a quarantine directory.
// It never permanently unlinks or executes os.RemoveAll on target files.
func MoveToTrash(targetPath string) error {
	if targetPath == "" {
		return appfault.New("target path cannot be empty")
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return appfault.Wrapf(err, "resolve absolute path for %s", targetPath)
	}

	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return nil
	}

	trashErr := moveToTrashOS(absPath)
	if trashErr != nil {
		fallbackErr := fallbackQuarantine(absPath)
		if fallbackErr != nil {
			return appfault.Wrapf(fallbackErr, "trash failed (%v); quarantine fallback also failed for %s", trashErr, absPath)
		}

		return nil
	}

	return nil
}
