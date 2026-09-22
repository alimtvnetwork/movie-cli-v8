// movie_ui.go — movie ui: starts the local web UI and opens it in the browser.
package cmd

import (
	"context"
	"fmt"
	"net/http"
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
	Use:     "ui",
	Aliases: []string{"dashboard", "gui", "web"},
	Short:   "Launch the interactive movie library web UI",
	Long: `Starts the local HTTP server and automatically opens the browser
to the interactive movie-cli dashboard.

Examples:
  movie ui               # Starts on port 8086 and opens default browser
  movie ui --port 9000   # Starts on port 9000
  movie ui --no-open     # Starts server without launching browser`,
	Run: runMovieUI,
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

	initRestLogger(database)
	mux := buildRESTMux(database)

	targetURL := fmt.Sprintf("http://%s:%d", uiHost, uiPort)
	printUIBanner(targetURL, database)

	if !uiNoOpen {
		go openBrowser(targetURL)
	}

	startHTTPServer(mux)
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

func printUIBanner(targetURL string, database *db.DB) {
	isColor := isColorEnabled()
	bullet := colorText("●", ansiCyan, isColor)
	status := database.GetSplitDBStatus()

	fmt.Println()
	fmt.Println(colorText("  ┌──────────────────────────────────────────────────────────┐", ansiCyan, isColor))
	fmt.Println(colorText("  │   🎬 MOVIE CLI — Interactive Web Dashboard               │", ansiCyan, isColor))
	fmt.Println(colorText("  └──────────────────────────────────────────────────────────┘", ansiCyan, isColor))
	fmt.Println()
	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Dashboard URL:", ansiDim, isColor), colorText(targetURL, ansiCyan, isColor))
	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Bind Address:", ansiDim, isColor), colorText(fmt.Sprintf("%s:%d", uiHost, uiPort), ansiWhite, isColor))
	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Process PID:", ansiDim, isColor), colorText(fmt.Sprintf("%d", os.Getpid()), ansiWhite, isColor))
	fmt.Printf("  %s %-16s %s (%s)\n", bullet, colorText("Primary DB:", ansiDim, isColor), colorText("movie.db", ansiWhite, isColor), status.MasterTier.SizeFormatted)
	fmt.Printf("  %s %-16s %s (%s)\n", bullet, colorText("Cache DB:", ansiDim, isColor), colorText("cache.db", ansiWhite, isColor), status.CacheTier.SizeFormatted)
	fmt.Printf("  %s %-16s %s\n", bullet, colorText("Stop Server:", ansiDim, isColor), colorText("Ctrl+C", ansiYellow, isColor))
	fmt.Println()
}
