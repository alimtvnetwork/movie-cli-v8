// movie_db_version.go — movie db version: prints applied schema migrations.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var movieDBVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current database schema version",
	Long:  `Prints all applied schema migrations with version numbers and timestamps.`,
	Run:   runMovieDBVersion,
}

func runMovieDBVersion(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	entries, err := database.ListSchemaVersions()
	if err != nil {
		errlog.Error("Schema version error: %v", err)
		return
	}

	fmt.Println()
	fmt.Println("📋 Schema Migrations")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(entries) == 0 {
		fmt.Println("  No migrations applied yet.")
		fmt.Println()
		return
	}

	for _, e := range entries {
		fmt.Printf("  v%-4d │ %-30s │ %s\n", e.Version, e.Description, e.AppliedAt)
	}

	latest := entries[len(entries)-1]
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Current schema: v%d\n", latest.Version)
	fmt.Println()
}
