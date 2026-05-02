// movie_ls.go — movie ls
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var (
	lsFormat  string
	lsAll     bool
	lsMissing bool
)

var movieLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List scanned movies and TV shows from your library",
	Long: `Lists movies and TV shows from your library.

Filter behavior (mutually exclusive — pick at most one):
  (default)   Only SCANNED items — rows that have a known file path on disk
              (added via 'movie scan'). Metadata-only entries are hidden.
  --all       Every item in the database, including metadata-only entries
              that have no associated file path (e.g. discovered via TMDb,
              imported manually, or whose file was later removed).
  --missing   Only items WITHOUT a file path (metadata-only / never scanned
              or whose original file is gone). Useful to spot orphaned
              entries or rows that need a rescan.

If both --all and --missing are passed, --missing wins (more specific filter).

Navigation: press N for next page, P for previous, Q to quit.

Output formats:
  --format default  Interactive paged view (default).
  --format json     Emit all matching items as JSON to stdout (for piping).
  --format table    Emit all matching items as a formatted table (no pager).`,
	Run: runMovieLs,
}

func init() {
	movieLsCmd.Flags().StringVar(&lsFormat, "format", "default",
		"output format: default, json, or table")
	movieLsCmd.Flags().BoolVar(&lsAll, "all", false,
		"include every item, even metadata-only entries with no file path")
	movieLsCmd.Flags().BoolVar(&lsMissing, "missing", false,
		"only show items WITHOUT a file path (metadata-only / unscanned)")
}

// lsJSONItem represents a single media item in JSON output.
type lsJSONItem struct {
	Title      string  `json:"title"`
	CleanTitle string  `json:"clean_title"`
	Type       string  `json:"type"`
	ImdbID     string  `json:"imdb_id,omitempty"`
	Genre      string  `json:"genre,omitempty"`
	Director   string  `json:"director,omitempty"`
	Language   string  `json:"language,omitempty"`
	FilePath   string  `json:"file_path,omitempty"`
	ID         int64   `json:"id"`
	FileSize   int64   `json:"file_size,omitempty"`
	TmdbRating float64 `json:"tmdb_rating,omitempty"`
	ImdbRating float64 `json:"imdb_rating,omitempty"`
	Popularity float64 `json:"popularity,omitempty"`
	Year       int     `json:"year,omitempty"`
	TmdbID     int     `json:"tmdb_id,omitempty"`
	Runtime    int     `json:"runtime,omitempty"`
}

func runMovieLs(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	switch lsFormat {
	case string(db.OutputFormatJSON):
		runMovieLsJSON(database)
	case string(db.OutputFormatTable):
		runMovieLsTable(database)
	default:
		runMovieLsInteractive(database)
	}
}

// resolveLsFilterMode maps the --all / --missing flags to a DB filter mode.
// SHARED: used by JSON, table, and interactive ls paths to keep filter
// semantics consistent. --missing wins over --all when both are set.
func resolveLsFilterMode() string {
	if lsMissing {
		return "missing"
	}
	if lsAll {
		return "all"
	}
	return "scanned"
}

func runMovieLsJSON(database *db.DB) {
	allMedia, err := database.ListMediaFiltered(0, 100000, resolveLsFilterMode())
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}

	items := make([]lsJSONItem, len(allMedia))
	for i := range allMedia {
		items[i] = toLsJSONItem(&allMedia[i])
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if encErr := enc.Encode(items); encErr != nil {
		errlog.Error("JSON encode error: %v", encErr)
	}
}

func toLsJSONItem(m *db.Media) lsJSONItem {
	return lsJSONItem{
		ID: m.ID, Title: m.Title, CleanTitle: m.CleanTitle,
		Year: m.Year, Type: m.Type, TmdbID: m.TmdbID, ImdbID: m.ImdbID,
		TmdbRating: m.TmdbRating, ImdbRating: m.ImdbRating, Popularity: m.Popularity,
		Genre: m.Genre, Director: m.Director, Runtime: m.Runtime,
		Language: m.Language, FilePath: m.CurrentFilePath, FileSize: m.FileSize,
	}
}

func runMovieLsInteractive(database *db.DB) {
	pageSize := resolvePageSize(database)
	mode := resolveLsFilterMode()

	total, countErr := database.CountMediaFiltered(mode)
	if countErr != nil {
		errlog.Error(msgDatabaseError, countErr)
		return
	}
	if total == 0 {
		fmt.Println(emptyLsMessage(mode))
		return
	}

	offset := 0
	scanner := bufio.NewScanner(os.Stdin)

	for {
		media, listErr := database.ListMediaFiltered(offset, pageSize, mode)
		if listErr != nil {
			errlog.Error("Error: %v", listErr)
			return
		}

		pg := LsPage{Offset: offset, PageSize: pageSize, Total: total}
		printLsPage(database, media, pg)

		if !scanner.Scan() {
			break
		}
		offset = handleLsInput(scanner.Text(), pg, database)
		if offset < 0 {
			return
		}
	}
}

// emptyLsMessage returns a filter-aware empty-state hint.
func emptyLsMessage(mode string) string {
	switch mode {
	case "missing":
		return "✅ No metadata-only entries — every item has a file path."
	case "all":
		return "📭 Library is empty. Run 'movie scan <folder>' to add items."
	default:
		return "📭 No scanned media found. Run 'movie scan <folder>' first."
	}
}

func resolvePageSize(database *db.DB) int {
	pageSizeStr, cfgErr := database.GetConfig("PageSize")
	if cfgErr != nil && cfgErr.Error() != "sql: no rows in result set" {
		errlog.Warn("Config read error (page_size): %v", cfgErr)
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize <= 0 {
		pageSize = 20
	}
	return pageSize
}

func printLsPage(database *db.DB, media []db.Media, pg LsPage) {
	fmt.Print("\033[H\033[2J")
	page := (pg.Offset / pg.PageSize) + 1
	totalPages := (pg.Total + pg.PageSize - 1) / pg.PageSize
	fmt.Printf("🎬 Your Library — Page %d/%d (%d total)\n", page, totalPages, pg.Total)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if page == 1 {
		printScanFolders(database)
	}

	for i := range media {
		printLsMediaRow(pg.Offset+i+1, &media[i])
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Print("  [N] Next  [P] Previous  [Q] Quit  [1-9] View details → ")
}

func printScanFolders(database *db.DB) {
	scanFolders, _ := database.ListDistinctScanFolders()
	if len(scanFolders) == 0 {
		return
	}
	fmt.Print("  📂 Scanned: ")
	for i, f := range scanFolders {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(f)
		if i >= 2 && len(scanFolders) > 3 {
			fmt.Printf(" (+%d more)", len(scanFolders)-3)
			break
		}
	}
	fmt.Println()
	fmt.Println()
}

func printLsMediaRow(num int, m *db.Media) {
	yearStr := ""
	if m.Year > 0 {
		yearStr = fmt.Sprintf("(%d)", m.Year)
	}
	rating := bestRating(m)
	typeIcon := db.TypeIcon(m.Type)
	fmt.Printf("  %3d. %-40s %-6s  ⭐ %-4s  %s %s\n",
		num, m.CleanTitle, yearStr, rating, typeIcon, capitalize(m.Type))
}

func bestRating(m *db.Media) string {
	if m.TmdbRating > 0 {
		return fmt.Sprintf("%.1f", m.TmdbRating)
	}
	if m.ImdbRating > 0 {
		return fmt.Sprintf("%.1f", m.ImdbRating)
	}
	return "N/A"
}

func handleLsInput(input string, pg LsPage, database *db.DB) int {
	switch {
	case input == "n" || input == "N":
		if pg.Offset+pg.PageSize < pg.Total {
			return pg.Offset + pg.PageSize
		}
		fmt.Println("  ⚠️  Already on last page")
		return pg.Offset
	case input == "p" || input == "P":
		if pg.Offset-pg.PageSize >= 0 {
			return pg.Offset - pg.PageSize
		}
		fmt.Println("  ⚠️  Already on first page")
		return pg.Offset
	case input == "q" || input == "Q":
		fmt.Println("👋 Bye!")
		return -1
	default:
		if num, parseErr := strconv.Atoi(input); parseErr == nil && num > 0 && num <= pg.Total {
			showMediaDetail(database, int64(num))
			fmt.Print("\nPress Enter to continue...")
		}
		return pg.Offset
	}
}
