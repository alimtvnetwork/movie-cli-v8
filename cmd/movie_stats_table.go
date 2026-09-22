// movie_stats_table.go — table-formatted output for movie stats
package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

const statsLabelW = 20
const statsValueW = 40

// printStatsTable outputs library statistics as a formatted table.
func printStatsTable(database *db.DB, counts StatsCounts) {
	isColor := isColorEnabled()

	fmt.Println()
	printStatsTableHeader(isColor)
	printStatsTableCounts(counts.Movies, counts.TV, counts.Total)
	printStatsTableStorage(database, counts.Total)
	printStatsTableSplitDB(database)
	printStatsTableGenres(database)
	printStatsTableRatings(database)
	printStatsTableFooter()
}

func printStatsTableHeader(isColor bool) {
	fmt.Printf("  ┌%s┬%s┐\n", strings.Repeat("─", statsLabelW+2), strings.Repeat("─", statsValueW+2))
	fmt.Printf("  │ %-*s │ %-*s │\n",
		statsLabelW, colorText("Metric", ansiCyan, isColor),
		statsValueW, colorText("Value", ansiCyan, isColor))
	fmt.Printf("  ├%s┼%s┤\n", strings.Repeat("─", statsLabelW+2), strings.Repeat("─", statsValueW+2))
}

func printStatsTableSep(mid string) {
	fmt.Printf("  ├%s%s%s┤\n",
		strings.Repeat("─", statsLabelW+2), mid, strings.Repeat("─", statsValueW+2))
}

func printStatsTableFooter() {
	fmt.Printf("  └%s┴%s┘\n", strings.Repeat("─", statsLabelW+2), strings.Repeat("─", statsValueW+2))
	fmt.Println()
}

func printStatsTableCounts(totalMovies, totalTV, total int) {
	printStatsRow("Total Movies", fmt.Sprintf("%d", totalMovies))
	printStatsRow("Total TV Shows", fmt.Sprintf("%d", totalTV))
	printStatsRow("Total", fmt.Sprintf("%d", total))
}

func printStatsTableStorage(database *db.DB, total int) {
	totalSize, largestSize, smallestSize, sizeErr := database.FileSizeStats()
	if sizeErr != nil || totalSize <= 0 {
		return
	}

	printStatsRow("Total Size", db.HumanSize(totalSize))
	printStatsRow("Largest File", db.HumanSize(largestSize))
	printStatsRow("Smallest File", db.HumanSize(smallestSize))

	if total > 0 {
		printStatsRow("Average Size", db.HumanSize(totalSize/float64(total)))
	}
}

func printStatsTableSplitDB(database *db.DB) {
	status := database.GetSplitDBStatus()

	printStatsTableSep("┼")
	printStatsRow("Primary DB", fmt.Sprintf("%s (%s, %s)", status.MasterTier.Name, status.MasterTier.SizeFormatted, status.MasterTier.JournalMode))
	printStatsRow("Cache DB", fmt.Sprintf("%s (%s, %s)", status.CacheTier.Name, status.CacheTier.SizeFormatted, status.CacheTier.JournalMode))
}

func printStatsTableGenres(database *db.DB) {
	sorted := sortedGenreCounts(database, 10)
	if len(sorted) == 0 {
		return
	}

	printStatsTableSep("┼")

	for _, g := range sorted {
		printStatsRow(g.name, fmt.Sprintf("%d", g.count))
	}
}

func printStatsTableRatings(database *db.DB) {
	avgImdb, avgTmdb := computeAvgRatings(database)
	hasRatings := avgImdb > 0 || avgTmdb > 0

	if !hasRatings {
		return
	}

	printStatsTableSep("┼")

	if avgImdb > 0 {
		printStatsRow("Avg IMDb Rating", fmt.Sprintf("⭐ %.1f / 10", avgImdb))
	}

	if avgTmdb > 0 {
		printStatsRow("Avg TMDb Rating", fmt.Sprintf("⭐ %.1f / 10", avgTmdb))
	}
}

func printStatsRow(label, value string) {
	fmt.Printf("  │ %-*s │ %-*s │\n", statsLabelW, label, statsValueW, value)
}

// sortedGenreCounts returns genres sorted by count.
func sortedGenreCounts(database *db.DB, limit int) []genreCount {
	genres, genreErr := database.TopGenres(limit)
	if genreErr != nil || len(genres) == 0 {
		return nil
	}

	var sorted []genreCount

	for n, c := range genres {
		sorted = append(sorted, genreCount{n, c})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	return sorted
}

// computeAvgRatings returns average IMDb and TMDb ratings.
func computeAvgRatings(database *db.DB) (avgImdb, avgTmdb float64) {
	allMedia, listErr := database.ListMedia(0, 100000)
	if listErr != nil {
		return 0, 0
	}

	var sumImdb, sumTmdb float64
	var cntImdb, cntTmdb int

	for i := range allMedia {
		if allMedia[i].ImdbRating > 0 {
			sumImdb += allMedia[i].ImdbRating
			cntImdb++
		}

		if allMedia[i].TmdbRating > 0 {
			sumTmdb += allMedia[i].TmdbRating
			cntTmdb++
		}
	}

	if cntImdb > 0 {
		avgImdb = sumImdb / float64(cntImdb)
	}

	if cntTmdb > 0 {
		avgTmdb = sumTmdb / float64(cntTmdb)
	}

	return
}
