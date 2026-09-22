// movie_ui.go — movie ui: starts the local web UI and opens it in the browser.
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var (
	uiPort   int
	uiHost   string
	uiNoOpen bool
)

var movieUiCmd = &cobra.Command{
	Use:     "ui [folder-or-alias-or-number]",
	Aliases: []string{"dashboard", "gui", "web"},
	Short:   "Launch the interactive movie library web UI",
	Long: `Starts the local HTTP server and automatically opens the browser
to the interactive movie-cli dashboard.

You can launch the web UI scoped to a specific scanned folder from anywhere:
  movie ui                      # Launch UI for entire library
  movie ui movies               # Launch UI scoped to 'movies' folder
  movie ui tvshows              # Launch UI scoped to 'tvshows' folder
  movie ui 1                    # Launch UI scoped to folder #1
  movie ui "D:\Downloads"       # Launch UI for specified directory

Examples:
  movie ui                      # Starts on port 8086 and opens default browser
  movie ui --port 9000          # Starts on port 9000
  movie ui --no-open            # Starts server without launching browser`,
	Args: cobra.MaximumNArgs(1),
	Run:  runMovieUI,
}

func init() {
	movieUiCmd.Flags().IntVarP(&uiPort, "port", "p", 8086, "port to listen on")
	movieUiCmd.Flags().StringVar(&uiHost, "host", "127.0.0.1", "host address to bind")
	movieUiCmd.Flags().BoolVar(&uiNoOpen, "no-open", false, "do not open default browser automatically")
}

func runMovieUI(cmd *cobra.Command, args []string) {
	database, err := db.Open()

	if err != nil {
		errlog.Error(msgDatabaseError, err)

		return
	}

	defer database.Close()

	scopedTarget := resolveUIFolderTarget(database, args)

	initRestLogger(database)
	mux := buildRESTMux(database)

	baseURL := fmt.Sprintf("http://%s:%d", uiHost, uiPort)
	targetURL := baseURL

	if scopedTarget != nil {
		targetURL = fmt.Sprintf("%s?folder=%s", baseURL, url.QueryEscape(scopedTarget.TargetDirectory))
	}

	printUIBanner(targetURL, database, scopedTarget)

	if !uiNoOpen {
		go openBrowser(targetURL)
	}

	startHTTPServer(mux)
}

func resolveUIFolderTarget(database *db.DB, args []string) *CdTargetResult {
	if len(args) == 0 {
		return nil
	}

	targetRes, suggestions, resErr := resolveCdTarget(database, args[0], "")

	if resErr != nil {
		fmt.Fprintf(os.Stderr, "❌ Target not found: %v\n", resErr)
		fmt.Fprintln(os.Stderr, "💡 Run 'movie cd' or 'movie ls --folders' to see available folders.")
		os.Exit(1)
	}

	if targetRes != nil {
		return targetRes
	}

	if len(suggestions) > 0 {
		fmt.Fprintln(os.Stderr, "⚠️ Ambiguous folder query. Available matches:")
		printCdSuggestions(suggestions)
		os.Exit(1)
	}

	return nil
}

func startHTTPServer(mux http.Handler) {
	addr := fmt.Sprintf("%s:%d", uiHost, uiPort)
	server := &http.Server{
		Addr:    addr,
		Handler: logMiddleware(mux),
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if srvErr := server.ListenAndServe(); srvErr != nil && srvErr != http.ErrServerClosed {
			errlog.Error("UI server error: %v", srvErr)
		}
	}()

	<-stopChan
	isColor := isColorEnabled()
	fmt.Printf("\n  %s %s\n\n", colorText("🛑", ansiYellow, isColor), colorText("Shutting down web UI server gracefully...", ansiDim, isColor))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if shutErr := server.Shutdown(shutdownCtx); shutErr != nil {
		errlog.Warn("Server shutdown warning: %v", shutErr)
	}
}

func printUIBanner(targetURL string, database *db.DB, scoped *CdTargetResult) {
	isColor := isColorEnabled()
	bullet := colorText("●", ansiCyan, isColor)
	status := database.GetSplitDBStatus()

	fmt.Println()
	fmt.Println(colorText("  ┌──────────────────────────────────────────────────────────┐", ansiCyan, isColor))
	fmt.Println(colorText("  │   🎬 MOVIE CLI — Interactive Web Dashboard               │", ansiCyan, isColor))
	fmt.Println(colorText("  └──────────────────────────────────────────────────────────┘", ansiCyan, isColor))
	fmt.Println()
	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Dashboard URL:", ansiDim, isColor), colorText(targetURL, ansiCyan, isColor))

	if scoped != nil {
		fmt.Printf("  %s %-16s %s (%s, %d items)\n",
			bullet,
			colorText("Scoped Folder:", ansiDim, isColor),
			colorText(scoped.TargetDirectory, ansiCyan, isColor),
			scoped.MatchName,
			scoped.ItemCount)
	}

	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Bind Address:", ansiDim, isColor), colorText(fmt.Sprintf("%s:%d", uiHost, uiPort), ansiWhite, isColor))
	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Process PID:", ansiDim, isColor), colorText(fmt.Sprintf("%d", os.Getpid()), ansiWhite, isColor))
	fmt.Printf("  %s %-16s %s (%s)\n", bullet, colorText("Primary DB:", ansiDim, isColor), colorText("movie.db", ansiWhite, isColor), status.MasterTier.SizeFormatted)
	fmt.Printf("  %s %-16s %s (%s)\n", bullet, colorText("Cache DB:", ansiDim, isColor), colorText("cache.db", ansiWhite, isColor), status.CacheTier.SizeFormatted)
	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Stop Server:", ansiDim, isColor), colorText("Ctrl+C", ansiYellow, isColor))
	fmt.Println()
}
