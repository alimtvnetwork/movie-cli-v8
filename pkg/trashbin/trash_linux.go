//go:build linux

package trashbin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

func moveToTrashOS(absPath string) error {
	if gioErr := moveViaGio(absPath); gioErr == nil {
		return nil
	}

	if putErr := moveViaTrashPut(absPath); putErr == nil {
		return nil
	}

	return moveViaXDGTrashSpec(absPath)
}

func moveViaGio(absPath string) error {
	cmd := exec.Command("gio", "trash", absPath)
	return cmd.Run()
}

func moveViaTrashPut(absPath string) error {
	cmd := exec.Command("trash-put", absPath)
	return cmd.Run()
}

func moveViaXDGTrashSpec(absPath string) error {
	trashDir := os.Getenv("XDG_DATA_HOME")
	if trashDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return appfault.Wrap(err, "resolve user home directory")
		}
		trashDir = filepath.Join(homeDir, ".local", "share", "Trash")
	} else {
		trashDir = filepath.Join(trashDir, "Trash")
	}

	filesDir := filepath.Join(trashDir, "files")
	infoDir := filepath.Join(trashDir, "info")

	if mkErr := os.MkdirAll(filesDir, 0755); mkErr != nil {
		return appfault.Wrap(mkErr, "create trash files directory")
	}

	if mkErr := os.MkdirAll(infoDir, 0755); mkErr != nil {
		return appfault.Wrap(mkErr, "create trash info directory")
	}

	baseName := filepath.Base(absPath)
	destPath := filepath.Join(filesDir, baseName)
	infoPath := filepath.Join(infoDir, baseName+".trashinfo")

	if _, statErr := os.Stat(destPath); statErr == nil {
		timestamp := time.Now().Format("20060102-150405")
		destPath = filepath.Join(filesDir, fmt.Sprintf("%s-%s", timestamp, baseName))
		infoPath = filepath.Join(infoDir, fmt.Sprintf("%s-%s.trashinfo", timestamp, baseName))
	}

	infoContent := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		absPath,
		time.Now().Format("2006-01-02T15:04:05"),
	)

	if writeErr := os.WriteFile(infoPath, []byte(infoContent), 0644); writeErr != nil {
		return appfault.Wrap(writeErr, "write trashinfo file")
	}

	if renErr := os.Rename(absPath, destPath); renErr != nil {
		_ = os.Remove(infoPath)
		return appfault.Wrap(renErr, "move file into trash files directory")
	}

	return nil
}
