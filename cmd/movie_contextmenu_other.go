// movie_contextmenu_other.go — fallback for OSes we don't support yet.
//go:build !windows && !linux && !darwin

package cmd

import (
	"fmt"
	"runtime"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

func installContextMenu(exePath string) error {
	return apperror.New(fmt.Sprintf("context menu not supported on %s", runtime.GOOS))
}

func uninstallContextMenu() error {
	return apperror.New(fmt.Sprintf("context menu not supported on %s", runtime.GOOS))
}

func contextMenuStatus() (bool, string) {
	return false, fmt.Sprintf("unsupported OS: %s", runtime.GOOS)
}

func printPostInstallHint() {}
