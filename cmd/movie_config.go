// movie_config.go — movie config
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var movieConfigCmd = &cobra.Command{
	Use:   "config [get|set] [key] [value]",
	Short: "Manage movie CLI configuration",
	Long: `View or update configuration settings.

Keys:
  movies_dir     - Default movies directory
  tv_dir         - Default TV shows directory
  archive_dir    - Default archive directory
  scan_dir       - Default scan directory
  tmdb_api_key   - TMDb API key
  tmdb_token     - TMDb access token
  page_size      - Items per page in list view

Examples:
  movie config                           # Show all
  movie config get movies_dir            # Get one
  movie config set movies_dir ~/Movies   # Set one
  movie config set tmdb_api_key abc123   # Set API key
  movie config set tmdb_token eyJ...     # Set access token`,
	Run: runMovieConfig,
}

func runMovieConfig(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)

		return
	}

	defer database.Close()

	if len(args) == 0 {
		showAllConfig(database)

		return
	}

	action := args[0]
	isColor := isColorEnabled()

	switch action {
	case "get":
		handleConfigGet(database, args, isColor)

	case "set":
		handleConfigSet(database, args, isColor)

	default:
		errlog.Error("Unknown action: %s. Use 'get' or 'set'.", action)
	}
}

func handleConfigGet(database *db.DB, args []string, isColor bool) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "❌ Usage: movie config get <key>")

		return
	}

	key := args[1]
	val, getErr := database.GetConfig(key)
	bullet := colorText("●", ansiCyan, isColor)

	if getErr != nil {
		fmt.Printf("  %s %s = %s\n", bullet, colorText(key, ansiDim, isColor), colorText("(not set)", ansiDim, isColor))

		return
	}

	fmt.Printf("  %s %s = %s\n", bullet, colorText(key, ansiDim, isColor), colorText(val, ansiWhite, isColor))
}

func handleConfigSet(database *db.DB, args []string, isColor bool) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "❌ Usage: movie config set <key> <value>")

		return
	}

	key, value := args[1], args[2]
	if setErr := database.SetConfig(key, value); setErr != nil {
		errlog.Error("Config set error: %v", setErr)

		return
	}

	fmt.Printf("  %s Set %s = %s\n", colorText("[ok]", ansiGreen, isColor), colorText(key, ansiWhite, isColor), colorText(value, ansiCyan, isColor))
}

func showAllConfig(database *db.DB) {
	isColor := isColorEnabled()
	bullet := colorText("●", ansiCyan, isColor)

	fmt.Println()
	fmt.Println(colorText("  ┌──────────────────────────────────────────────────────────┐", ansiCyan, isColor))
	fmt.Println(colorText("  │  ⚙️  Movie CLI Configuration Settings                     │", ansiCyan, isColor))
	fmt.Println(colorText("  └──────────────────────────────────────────────────────────┘", ansiCyan, isColor))
	fmt.Println()

	fmt.Println(colorText("  ── Active Settings ──", ansiCyan, isColor))

	keys := []string{"MoviesDir", "TvDir", "ArchiveDir", "ScanDir", "TmdbApiKey", "TmdbToken", "PageSize"}
	for _, key := range keys {
		val := formatConfigValue(database, key, isColor)

		fmt.Printf("  %s %-16s %s\n", bullet, colorText(key+":", ansiDim, isColor), val)
	}

	fmt.Println()
}

func formatConfigValue(database *db.DB, key string, isColor bool) string {
	val, err := database.GetConfig(key)
	if err != nil {
		return colorText("(not set)", ansiDim, isColor)
	}

	hasSecret := (key == "TmdbApiKey" || key == "TmdbToken") && len(val) > 8
	if hasSecret {
		masked := val[:4] + "••••••••" + val[len(val)-4:]

		return colorText(masked, ansiYellow, isColor)
	}

	return colorText(val, ansiWhite, isColor)
}
