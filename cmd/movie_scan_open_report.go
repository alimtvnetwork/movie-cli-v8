// movie_scan_open_report.go — auto-open the generated report.html in the
// system browser after a scan or rescan completes.
//
// SHARED: openScanReport(outputDir), pathToFileURL(path).
// Callers: movie scan (post-scan finalize), movie rescan (regenerate path).
// Do NOT re-implement file:// URL building or the --no-open gate elsewhere
// — extend this file so all entry points honour the same opt-out flag.
//
// Skipped when:
//   - --dry-run    (nothing was generated)
//   - --json       (machine output should not spawn UI)
//   - --rest       (REST flow already opens http://localhost in browser)
//   - --no-open    (explicit user opt-out)
//   - report.html  is missing on disk
package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// scanNoOpen disables the auto-open behaviour when set via --no-open.
var scanNoOpen bool

// openScanReport opens the generated report.html for the given scan output dir.
// Safe to call unconditionally; it returns early when auto-open should be skipped.
func openScanReport(outputDir string) {
	if scanNoOpen {
		return
	}
	reportPath := filepath.Join(outputDir, "report.html")
	if _, err := os.Stat(reportPath); err != nil {
		return
	}
	fileURL := pathToFileURL(reportPath)
	fmt.Printf("\n🌐 Opening report in your browser...\n   %s\n", reportPath)
	openBrowser(fileURL)
}

// pathToFileURL converts an absolute filesystem path into a file:// URL that
// browsers accept across Windows, macOS and Linux.
func pathToFileURL(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = p
	}
	if runtime.GOOS == "windows" {
		abs = strings.ReplaceAll(abs, "\\", "/")
		return "file:///" + (&url.URL{Path: abs}).String()
	}
	return "file://" + (&url.URL{Path: abs}).String()
}
