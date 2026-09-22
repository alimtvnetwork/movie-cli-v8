// SHARED HELPER: movie_scan_helpers_print.go — print helpers for scan footer (extracted from movie_scan_helpers.go).
// movie_scan_helpers_print.go — print helpers for scan footer (extracted from movie_scan_helpers.go).
//
// SHARED: scan summary footer rendering (counts, sizes, durations).
// Callers: movie scan, movie rescan, movie rescan-failed.
// Do NOT re-implement the footer in JSON/table renderers — they should
// call these helpers so plain output stays consistent across modes.
package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

func printScanCounts(stats ScanStats) {
	isColor := isColorEnabled()

	fmt.Println()
	fmt.Println(colorText("  ┌──────────────────────────────────────────────────────────┐", ansiDim, isColor))
	fmt.Println(colorText("  │   🎬 MOVIE CLI — Scan Summary Report                     │", ansiCyan, isColor))
	fmt.Println(colorText("  └──────────────────────────────────────────────────────────┘", ansiDim, isColor))
	fmt.Println()

	fmt.Printf("  %s %s:\n", colorText("●", ansiCyan, isColor), colorText("Library Additions", ansiBold, isColor))
	fmt.Printf("    %-18s %d\n", colorText("Total Files:", ansiDim, isColor), stats.Total)
	fmt.Printf("    %-18s %d\n", colorText("Movies:", ansiDim, isColor), stats.Movies)
	fmt.Printf("    %-18s %d\n", colorText("TV Series:", ansiDim, isColor), stats.TV)

	newCount := stats.Total - stats.Skipped

	if newCount > 0 {
		fmt.Printf("    %-18s %s\n", colorText("New Additions:", ansiDim, isColor), colorText(fmt.Sprintf("%d", newCount), ansiGreen, isColor))
	}

	if stats.Skipped > 0 {
		fmt.Printf("    %-18s %d (already in DB)\n", colorText("Existing:", ansiDim, isColor), stats.Skipped)
	}

	if stats.Removed > 0 {
		fmt.Printf("    %-18s %s\n", colorText("Removed:", ansiDim, isColor), colorText(fmt.Sprintf("%d", stats.Removed), ansiYellow, isColor))
	}

	fmt.Println()
	fmt.Printf("  %s %s:\n", colorText("●", ansiCyan, isColor), colorText("Split-DB Storage", ansiBold, isColor))
	fmt.Printf("    %-18s %s\n", colorText("Master Store:", ansiDim, isColor), "movie.db (Primary Library)")
	fmt.Printf("    %-18s %s\n", colorText("Cache Store:", ansiDim, isColor), "cache.db (Lookups & Search)")
	fmt.Printf("    %-18s %s\n", colorText("Journal Mode:", ansiDim, isColor), "WAL")
}

func printScanOutputFiles(stats ScanStats) {
	isColor := isColorEnabled()

	fmt.Println()
	fmt.Printf("  %s %s:\n", colorText("●", ansiCyan, isColor), colorText("Output Artifacts", ansiBold, isColor))
	fmt.Printf("    %-18s %s/\n", colorText("Target Folder:", ansiDim, isColor), stats.OutputDir)

	writeScanOutputSummary(stats)
	writeScanOutputHTML(stats)

	fmt.Printf("    %-18s %s\n", colorText("Media JSON:", ansiDim, isColor), "json/movie/, json/tv/")
	fmt.Printf("    %-18s %s\n", colorText("Thumbnails:", ansiDim, isColor), "thumbnails/")
}

func writeScanOutputSummary(stats ScanStats) {
	if err := writeScanSummary(stats); err != nil {
		errlog.Warn("Could not write summary.json: %v", err)

		return
	}

	isColor := isColorEnabled()
	fmt.Printf("    %-18s %s\n", colorText("Summary JSON:", ansiDim, isColor), "summary.json")
}

func writeScanOutputHTML(stats ScanStats) {
	if err := writeHTMLReport(stats); err != nil {
		errlog.Warn("Could not write report.html: %v", err)

		return
	}

	isColor := isColorEnabled()
	fmt.Printf("    %-18s %s\n", colorText("HTML Report:", ansiDim, isColor), "report.html")
}

func printScanGuidanceCard(stats ScanStats) {
	isColor := isColorEnabled()

	fmt.Println()
	fmt.Println(colorText("  ────────────────────────────────────────────────────────────", ansiDim, isColor))
	fmt.Println(colorText("  Quick Actions & Next Steps:", ansiCyan, isColor))
	fmt.Printf("    %-24s %s\n", colorText("movie ui", ansiYellow, isColor), "Launch local web dashboard in browser")
	fmt.Printf("    %-24s %s\n", colorText("movie ls", ansiYellow, isColor), "List indexed library movies and TV series")
	fmt.Printf("    %-24s %s\n", colorText("movie info <title>", ansiYellow, isColor), "Query TMDb and inspect media metadata")
	fmt.Printf("    %-24s %s\n", colorText("movie stats", ansiYellow, isColor), "View comprehensive library metrics")
	fmt.Printf("    %-24s %s\n", colorText("movie db", ansiYellow, isColor), "Inspect multi-tier Split-DB statistics")
	fmt.Println(colorText("  ────────────────────────────────────────────────────────────", ansiDim, isColor))
	fmt.Println()
}
