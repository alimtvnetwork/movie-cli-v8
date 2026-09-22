// movie_report.go — generates library summary report and HTML dashboard.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var (
	reportOutputDir string
	reportNoOpen    bool
)

var movieReportCmd = &cobra.Command{
	Use:   "report [output-dir]",
	Short: "Generate library summary report and HTML dashboard",
	Long: `Generate a rich HTML report and display a library summary card.

The report includes collection inventory, Split-DB storage footprints,
genre distribution, and ratings overview.`,
	Run: runMovieReport,
}

func init() {
	movieReportCmd.Flags().StringVarP(&reportOutputDir, "out", "o", ".movie-output",
		"output directory for report.html")
	movieReportCmd.Flags().BoolVar(&reportNoOpen, "no-open", false,
		"do not automatically open report in browser")
}

func runMovieReport(cmd *cobra.Command, args []string) {
	outDir := resolveReportOutputDir(args)

	database, err := db.Open()

	if err != nil {
		errlog.Error(msgDatabaseError, err)

		return
	}

	defer database.Close()

	items, listErr := database.ListMedia(0, 10000)

	if listErr != nil {
		errlog.Error(msgDatabaseError, listErr)

		return
	}

	if len(items) == 0 {
		fmt.Println("📭 No media in library. Run 'movie scan <folder>' first.")

		return
	}

	executeReportGeneration(database, items, outDir)
}

func resolveReportOutputDir(args []string) string {
	if len(args) > 0 {
		if args[0] != "" {
			return args[0]
		}
	}

	if reportOutputDir != "" {
		return reportOutputDir
	}

	return ".movie-output"
}

func executeReportGeneration(database *db.DB, items []db.Media, outDir string) {
	_ = os.MkdirAll(outDir, 0755)

	stats := buildReportScanStats(items, outDir)
	_ = writeHTMLReport(stats)

	printReportSummaryCard(database, stats, outDir)

	if !reportNoOpen {
		openScanReport(outDir)
	}
}

func buildReportScanStats(items []db.Media, outDir string) ScanStats {
	movies, tv := 0, 0

	for i := range items {
		if items[i].Type == string(db.MediaTypeMovie) {
			movies++
		}

		if items[i].Type != string(db.MediaTypeMovie) {
			tv++
		}
	}

	return ScanStats{
		ScanDir:   "Library",
		OutputDir: outDir,
		Items:     items,
		Total:     len(items),
		Movies:    movies,
		TV:        tv,
		Skipped:   0,
	}
}
