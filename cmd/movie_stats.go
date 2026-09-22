// movie_stats.go — movie stats
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var statsFormat string

var movieStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show library statistics",
	Long: `Display total counts, top genres, and average ratings.

Use --format json to output stats as JSON to stdout for piping.
Use --format table to output stats as a formatted table.`,
	Run: runMovieStats,
}

func init() {
	movieStatsCmd.Flags().StringVar(&statsFormat, "format", "default",
		"output format: default, table, or json")
}

// statsJSONOutput is the JSON structure for --format json.
type statsJSONOutput struct {
	Storage     *statsStorage `json:"storage,omitempty"`
	TopGenres   []statsGenre  `json:"top_genres,omitempty"`
	AvgImdb     float64       `json:"avg_imdb_rating,omitempty"`
	AvgTmdb     float64       `json:"avg_tmdb_rating,omitempty"`
	TotalMovies int           `json:"total_movies"`
	TotalTV     int           `json:"total_tv_shows"`
	Total       int           `json:"total"`
}

type statsStorage struct {
	TotalHuman   string `json:"total_human"`
	TotalSize    int64  `json:"total_bytes"`
	LargestFile  int64  `json:"largest_file_bytes"`
	SmallestFile int64  `json:"smallest_file_bytes"`
	AverageSize  int64  `json:"average_bytes"`
}

type statsGenre struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func runMovieStats(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	cwd, _ := os.Getwd()
	RecordContextMenuClick(database, cwd)

	totalMovies, _ := database.CountMedia(string(db.MediaTypeMovie))
	totalTV, _ := database.CountMedia(string(db.MediaTypeTV))
	total, err := database.CountMedia("")
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}

	if total == 0 {
		printEmptyStats()
		return
	}

	counts := StatsCounts{Movies: totalMovies, TV: totalTV, Total: total}
	switch statsFormat {
	case string(db.OutputFormatJSON):
		printStatsJSON(database, counts)
	case string(db.OutputFormatTable):
		printStatsTable(database, counts)
	default:
		printStatsDefault(database, counts)
	}
}

func printEmptyStats() {
	if statsFormat == "json" {
		fmt.Println("{}")
		return
	}
	fmt.Println("📭 No media in library. Run 'movie scan <folder>' first.")
}

func printStatsJSON(database *db.DB, counts StatsCounts) {
	out := statsJSONOutput{
		TotalMovies: counts.Movies, TotalTV: counts.TV, Total: counts.Total,
	}
	out.Storage = buildStatsStorageJSON(database, counts.Total)
	out.TopGenres = buildStatsGenresJSON(database)
	out.AvgImdb, out.AvgTmdb = computeAvgRatings(database)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if encErr := enc.Encode(out); encErr != nil {
		errlog.Error("JSON encode error: %v", encErr)
	}
}

func buildStatsStorageJSON(database *db.DB, total int) *statsStorage {
	totalSize, largestSize, smallestSize, sizeErr := database.FileSizeStats()
	if sizeErr != nil || totalSize <= 0 {
		return nil
	}
	return &statsStorage{
		TotalSize:    int64(totalSize * 1024 * 1024),
		TotalHuman:   db.HumanSize(totalSize),
		LargestFile:  int64(largestSize * 1024 * 1024),
		SmallestFile: int64(smallestSize * 1024 * 1024),
		AverageSize:  int64(totalSize*1024*1024) / int64(total),
	}
}

func buildStatsGenresJSON(database *db.DB) []statsGenre {
	genres, genreErr := database.TopGenres(10)
	if genreErr != nil || len(genres) == 0 {
		return nil
	}
	var out []statsGenre
	for n, c := range genres {
		out = append(out, statsGenre{Name: n, Count: c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}

func printStatsDefault(database *db.DB, counts StatsCounts) {
	isColor := isColorEnabled()

	fmt.Println()
	fmt.Println(colorText("  ┌──────────────────────────────────────────────────────────┐", ansiCyan, isColor))
	fmt.Println(colorText("  │  📊 Movie CLI Library Statistics                         │", ansiCyan, isColor))
	fmt.Println(colorText("  └──────────────────────────────────────────────────────────┘", ansiCyan, isColor))
	fmt.Println()

	printStatsDefaultCounts(counts.Movies, counts.TV, counts.Total, isColor)
	printStatsDefaultStorage(database, counts.Total, isColor)
	printStatsDefaultSplitDB(database, isColor)
	printStatsDefaultGenres(database, isColor)
	printStatsDefaultRatings(database, isColor)
}

func printStatsDefaultCounts(totalMovies, totalTV, total int, isColor bool) {
	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println(colorText("  ── Library Overview ──", ansiCyan, isColor))
	fmt.Printf("  %s %-18s %s\n", bullet, colorText("Total Movies:", ansiDim, isColor), colorText(fmt.Sprintf("%d", totalMovies), ansiWhite, isColor))
	fmt.Printf("  %s %-18s %s\n", bullet, colorText("Total TV Shows:", ansiDim, isColor), colorText(fmt.Sprintf("%d", totalTV), ansiWhite, isColor))
	fmt.Printf("  %s %-18s %s\n", bullet, colorText("Total Media:", ansiDim, isColor), colorText(fmt.Sprintf("%d", total), ansiBold, isColor))
	fmt.Println()
}

func printStatsDefaultStorage(database *db.DB, total int, isColor bool) {
	totalSize, largestSize, smallestSize, sizeErr := database.FileSizeStats()
	if sizeErr != nil {
		errlog.Warn("File size stats error: %v", sizeErr)

		return
	}

	if totalSize <= 0 {
		return
	}

	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println(colorText("  ── Storage & Filesystem ──", ansiCyan, isColor))
	fmt.Printf("  %s %-18s %s\n", bullet, colorText("Total Size:", ansiDim, isColor), colorText(db.HumanSize(totalSize), ansiWhite, isColor))
	fmt.Printf("  %s %-18s %s\n", bullet, colorText("Largest File:", ansiDim, isColor), colorText(db.HumanSize(largestSize), ansiWhite, isColor))
	fmt.Printf("  %s %-18s %s\n", bullet, colorText("Smallest File:", ansiDim, isColor), colorText(db.HumanSize(smallestSize), ansiWhite, isColor))

	if total > 0 {
		avgSize := totalSize / float64(total)

		fmt.Printf("  %s %-18s %s\n", bullet, colorText("Average Size:", ansiDim, isColor), colorText(db.HumanSize(avgSize), ansiWhite, isColor))
	}

	fmt.Println()
}

func printStatsDefaultSplitDB(database *db.DB, isColor bool) {
	status := database.GetSplitDBStatus()
	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println(colorText("  ── SQLite Split-DB Stores ──", ansiCyan, isColor))
	fmt.Printf("  %s %-18s %s (%s, %s)\n", bullet, colorText("Primary Library:", ansiDim, isColor),
		colorText(status.MasterTier.Name, ansiWhite, isColor),
		status.MasterTier.SizeFormatted,
		colorText(status.MasterTier.JournalMode+" mode", ansiGreen, isColor))
	fmt.Printf("  %s %-18s %s (%s, %s)\n", bullet, colorText("Ephemeral Cache:", ansiDim, isColor),
		colorText(status.CacheTier.Name, ansiWhite, isColor),
		status.CacheTier.SizeFormatted,
		colorText(status.CacheTier.JournalMode+" mode", ansiGreen, isColor))
	fmt.Printf("  %s %-18s %s\n", bullet, colorText("Total Footprint:", ansiDim, isColor),
		colorText(status.TotalSize, ansiWhite, isColor))
	fmt.Println()
}

func printStatsDefaultGenres(database *db.DB, isColor bool) {
	sorted := sortedGenreCounts(database, 10)
	if len(sorted) == 0 {
		return
	}

	fmt.Println(colorText("  ── Top Genres ──", ansiCyan, isColor))

	for _, g := range sorted {
		barLen := minInt(g.count, 25)
		bar := colorText(strings.Repeat("█", barLen), ansiCyan, isColor)

		fmt.Printf("     %-18s %s %s\n", g.name, bar, colorText(fmt.Sprintf("%d", g.count), ansiDim, isColor))
	}

	fmt.Println()
}

func printStatsDefaultRatings(database *db.DB, isColor bool) {
	avgImdb, avgTmdb := computeAvgRatings(database)
	hasRatings := avgImdb > 0 || avgTmdb > 0

	if !hasRatings {
		return
	}

	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println(colorText("  ── Ratings Overview ──", ansiCyan, isColor))

	if avgImdb > 0 {
		fmt.Printf("  %s %-18s ⭐ %s / 10\n", bullet, colorText("Average IMDb:", ansiDim, isColor), colorText(fmt.Sprintf("%.1f", avgImdb), ansiYellow, isColor))
	}

	if avgTmdb > 0 {
		fmt.Printf("  %s %-18s ⭐ %s / 10\n", bullet, colorText("Average TMDb:", ansiDim, isColor), colorText(fmt.Sprintf("%.1f", avgTmdb), ansiYellow, isColor))
	}

	fmt.Println()
}
