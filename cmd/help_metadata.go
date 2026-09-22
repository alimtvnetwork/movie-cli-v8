// help_metadata.go — renders GitMap binary and environment metadata card for root help.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/version"
)

func renderBinaryCard(b *strings.Builder, isColor bool) {
	masterDB, cacheDB := resolveDatabasePaths()
	bullet := colorText("●", ansiCyan, isColor)

	b.WriteString(colorText("  ┌──────────────────────────────────────────────────────────┐\n", ansiCyan, isColor))
	b.WriteString(colorText("  │   🎬 MOVIE CLI — System & Environment Runtime            │\n", ansiCyan, isColor))
	b.WriteString(colorText("  └──────────────────────────────────────────────────────────┘\n\n", ansiCyan, isColor))

	b.WriteString(fmt.Sprintf("  %s %-16s %s\n", bullet, colorText("Version:", ansiDim, isColor), version.Full()))
	b.WriteString(fmt.Sprintf("  %s %-16s %s\n", bullet, colorText("Commit SHA:", ansiDim, isColor), version.Commit))
	b.WriteString(fmt.Sprintf("  %s %-16s %s\n", bullet, colorText("Platform:", ansiDim, isColor), fmt.Sprintf("%s/%s (%s)", runtime.GOOS, runtime.GOARCH, runtime.Version())))
	b.WriteString(fmt.Sprintf("  %s %-16s %s\n", bullet, colorText("Primary DB:", ansiDim, isColor), masterDB))
	b.WriteString(fmt.Sprintf("  %s %-16s %s\n", bullet, colorText("Cache DB:", ansiDim, isColor), cacheDB))
	b.WriteString(fmt.Sprintf("  %s %-16s %s\n", bullet, colorText("Built:", ansiDim, isColor), version.BuildDate))
	b.WriteString(fmt.Sprintf("  %s %-16s %s\n\n", bullet, colorText("Docs URL:", ansiDim, isColor), colorText("https://github.com/alimtvnetwork/movie-cli-v8", ansiWhite, isColor)))
}

func resolveDatabasePaths() (string, string) {
	exePath, _ := os.Executable()
	exePath, _ = filepath.EvalSymlinks(exePath)

	if exePath == "" {
		exePath = "movie"
	}

	dataDir := filepath.Join(filepath.Dir(exePath), "data")
	masterDB := filepath.Join(dataDir, "movie.db")
	cacheDB := filepath.Join(dataDir, "cache.db")

	return masterDB, cacheDB
}
