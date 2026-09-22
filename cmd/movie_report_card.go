// movie_report_card.go — renders GitMap-styled summary card for movie report.
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func printReportSummaryCard(database *db.DB, stats ScanStats, outDir string) {
	isColor := isColorEnabled()

	fmt.Println()
	fmt.Println(colorText("  ┌──────────────────────────────────────────────────────────┐", ansiCyan, isColor))
	fmt.Println(colorText("  │   🎬 MOVIE CLI — Library Catalog & Report Summary        │", ansiCyan, isColor))
	fmt.Println(colorText("  └──────────────────────────────────────────────────────────┘", ansiCyan, isColor))
	fmt.Println()

	printReportCatalogRow(stats, isColor)
	printReportSplitDBRow(database, isColor)
	printReportGenresRow(database, isColor)
	printReportRatingsRow(database, isColor)
	printReportHTMLPathRow(outDir, isColor)
	fmt.Println()
}

func printReportCatalogRow(stats ScanStats, isColor bool) {
	bullet := colorText("●", ansiCyan, isColor)
	label := colorText("Catalog:", ansiDim, isColor)
	val := fmt.Sprintf("%d items (%d movies, %d TV series)", stats.Total, stats.Movies, stats.TV)

	fmt.Printf("  %s %-16s %s\n", bullet, label, colorText(val, ansiWhite, isColor))
}

func printReportSplitDBRow(database *db.DB, isColor bool) {
	status := database.GetSplitDBStatus()
	bullet := colorText("●", ansiCyan, isColor)
	label := colorText("Split-DB:", ansiDim, isColor)
	val := fmt.Sprintf("movie.db (%s) | cache.db (%s)", status.MasterTier.SizeFormatted, status.CacheTier.SizeFormatted)

	fmt.Printf("  %s %-16s %s\n", bullet, label, colorText(val, ansiWhite, isColor))
}

func printReportGenresRow(database *db.DB, isColor bool) {
	genres, _ := database.TopGenres(3)

	if len(genres) == 0 {
		return
	}

	var parts []string

	for name, count := range genres {
		parts = append(parts, fmt.Sprintf("%s (%d)", name, count))
	}

	bullet := colorText("●", ansiCyan, isColor)
	label := colorText("Top Genres:", ansiDim, isColor)

	fmt.Printf("  %s %-16s %s\n", bullet, label, colorText(strings.Join(parts, ", "), ansiWhite, isColor))
}

func printReportRatingsRow(database *db.DB, isColor bool) {
	avgImdb, avgTmdb := computeAvgRatings(database)

	if avgImdb <= 0 {
		if avgTmdb <= 0 {
			return
		}
	}

	bullet := colorText("●", ansiCyan, isColor)
	label := colorText("Top Ratings:", ansiDim, isColor)
	val := fmt.Sprintf("⭐ %.1f / 10 avg", (avgImdb+avgTmdb)/2.0)

	if avgImdb > 0 {
		if avgTmdb <= 0 {
			val = fmt.Sprintf("⭐ %.1f / 10 avg", avgImdb)
		}
	}

	fmt.Printf("  %s %-16s %s\n", bullet, label, colorText(val, ansiYellow, isColor))
}

func printReportHTMLPathRow(outDir string, isColor bool) {
	reportFile := filepath.Join(outDir, "report.html")
	absPath, err := filepath.Abs(reportFile)

	if err != nil {
		absPath = reportFile
	}

	bullet := colorText("●", ansiCyan, isColor)
	label := colorText("HTML Report:", ansiDim, isColor)

	fmt.Printf("  %s %-16s %s\n", bullet, label, colorText(absPath, ansiCyan, isColor))
}
