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
	printUIBanner(targetURL)

	if !uiNoOpen {
		go openBrowser(targetURL)
	}

	handler := logMiddleware(mux)
	addr := fmt.Sprintf("%s:%d", uiHost, uiPort)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if srvErr := server.ListenAndServe(); srvErr != nil && srvErr != http.ErrServerClosed {
			errlog.Error("UI server error: %v", srvErr)
		}
	}()

	<-stopChan
	fmt.Println("\n  🛑 Shutting down web UI...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if shutErr := server.Shutdown(shutdownCtx); shutErr != nil {
		errlog.Warn("Server shutdown warning: %v", shutErr)
	}
}

func printUIBanner(url string) {
	fmt.Println()
	fmt.Println("  ╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║              🎬 movie-cli Interactive Web UI                  ║")
	fmt.Println("  ╚═══════════════════════════════════════════════════════════════╝")
	fmt.Printf("   🌐 Dashboard running at: %s\n", url)
	fmt.Println("   💡 Press Ctrl+C to stop the server")
	fmt.Println()
}
