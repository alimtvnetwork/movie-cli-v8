package trashbin

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

func fallbackQuarantine(absPath string) error {
	parentDir := filepath.Dir(absPath)
	quarantineDir := filepath.Join(parentDir, ".trash")

	if mkErr := os.MkdirAll(quarantineDir, 0755); mkErr != nil {
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return appfault.Wrap("resolve user home for fallback quarantine", homeErr)
		}

		quarantineDir = filepath.Join(homeDir, ".movie", "trash")
		if secMkErr := os.MkdirAll(quarantineDir, 0755); secMkErr != nil {
			return appfault.Wrap("create secondary quarantine directory", secMkErr)
		}
	}

	baseName := filepath.Base(absPath)
	timestamp := time.Now().Format("20060102_150405")
	destPath := filepath.Join(quarantineDir, fmt.Sprintf("%s_%s", timestamp, baseName))

	if err := os.Rename(absPath, destPath); err != nil {
		return appfault.Wrapf(err, "quarantine move %s -> %s", absPath, destPath)
	}

	return nil
}
