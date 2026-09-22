// movie_alias.go — movie alias: assign short names to scanned folders for quick access.
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var aliasApplyFlag bool

var movieAliasCmd = &cobra.Command{
	Use:     "alias [subcommand]",
	Aliases: []string{"a"},
	Short:   "Manage short alias names for scanned folders",
	Long: `Assign and manage short names for your scanned media folders for quick
navigation with 'movie cd' and 'movie ui'.

Subcommands:
  list       List all configured and auto-derived folder aliases (default)
  set        Create or update an alias: movie alias set <name> <folder-or-number>
  remove     Remove an alias: movie alias remove <name>
  show       Show details for a specific alias: movie alias show <name>
  suggest    Auto-suggest aliases for scanned root folders

Examples:
  movie alias                     List all aliases
  movie alias set mv 1            Set alias 'mv' to folder #1
  movie alias set tv "D:\TV"      Set alias 'tv' to D:\TV
  movie alias suggest --apply     Auto-register default aliases for all folders
  movie cd -A mv                  Jump to folder via alias`,
	Run: runAliasList,
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all folder aliases",
	Run:   runAliasList,
}

var aliasSetCmd = &cobra.Command{
	Use:   "set <alias> <folder-or-number>",
	Short: "Create or update an alias for a folder",
	Args:  cobra.ExactArgs(2),
	Run:   runAliasSet,
}

var aliasRemoveCmd = &cobra.Command{
	Use:     "remove <alias>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a folder alias",
	Args:    cobra.ExactArgs(1),
	Run:     runAliasRemove,
}

var aliasShowCmd = &cobra.Command{
	Use:   "show <alias>",
	Short: "Show details for an alias",
	Args:  cobra.ExactArgs(1),
	Run:   runAliasShow,
}

var aliasSuggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Auto-suggest aliases for scanned folders",
	Run:   runAliasSuggest,
}

func init() {
	aliasSuggestCmd.Flags().BoolVar(&aliasApplyFlag, "apply", false, "automatically save suggested aliases")

	movieAliasCmd.AddCommand(aliasListCmd)
	movieAliasCmd.AddCommand(aliasSetCmd)
	movieAliasCmd.AddCommand(aliasRemoveCmd)
	movieAliasCmd.AddCommand(aliasShowCmd)
	movieAliasCmd.AddCommand(aliasSuggestCmd)
}

func openDatabaseForAlias() *db.DB {
	database, err := db.Open()

	if err != nil {
		errlog.Error(msgDatabaseError, err)
		os.Exit(1)
	}

	return database
}

func runAliasList(cmd *cobra.Command, args []string) {
	database := openDatabaseForAlias()
	defer database.Close()

	aliases, _ := database.ListFolderAliases()
	stats, _ := database.ListScanFoldersWithStats()

	if len(aliases) == 0 && len(stats) == 0 {
		fmt.Println("📭 No scanned folders or aliases found. Run 'movie scan <folder>' first.")

		return
	}

	fmt.Println("🏷️  Folder Aliases & Navigation Shortcuts:")
	fmt.Println("────────────────────────────────────────────────────────────────────────────")
	fmt.Printf(" %-12s  %-8s  %-8s  %s\n", "ALIAS", "TYPE", "ITEMS", "FOLDER PATH")

	for _, a := range aliases {
		fmt.Printf(" %-12s  %-8s  %-8s  %s\n", a.AliasName, "manual", "—", a.FolderPath)
	}

	for _, s := range stats {
		fmt.Printf(" %-12s  %-8s  %-8d  %s\n", s.Alias, "auto", s.ItemCount, s.FolderPath)
	}

	fmt.Println("────────────────────────────────────────────────────────────────────────────")
	fmt.Println("💡 Quick Jump:  mcd <alias>   •   movie ui <alias>")
}

func runAliasSet(cmd *cobra.Command, args []string) {
	alias := strings.ToLower(strings.TrimSpace(args[0]))
	target := strings.TrimSpace(args[1])

	database := openDatabaseForAlias()
	defer database.Close()

	resolvedPath := target

	if num, err := strconv.Atoi(target); err == nil && num > 0 {
		stats, _ := database.ListScanFoldersWithStats()

		if num <= len(stats) {
			resolvedPath = stats[num-1].FolderPath
		}
	}

	if err := database.SetFolderAlias(alias, resolvedPath); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to save alias: %v\n", err)

		return
	}

	fmt.Printf("✔ Alias %q -> %s\n", alias, resolvedPath)
}

func runAliasRemove(cmd *cobra.Command, args []string) {
	alias := strings.ToLower(strings.TrimSpace(args[0]))

	database := openDatabaseForAlias()
	defer database.Close()

	if err := database.DeleteFolderAlias(alias); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to remove alias: %v\n", err)

		return
	}

	fmt.Printf("✔ Alias %q removed\n", alias)
}

func runAliasShow(cmd *cobra.Command, args []string) {
	alias := strings.ToLower(strings.TrimSpace(args[0]))

	database := openDatabaseForAlias()
	defer database.Close()

	path, err := database.GetFolderAlias(alias)

	if err == nil && path != "" {
		fmt.Printf("Alias:  %s\nType:   manual\nPath:   %s\n", alias, path)

		return
	}

	stats, _ := database.ListScanFoldersWithStats()

	for _, s := range stats {
		if strings.EqualFold(s.Alias, alias) {
			fmt.Printf("Alias:  %s\nType:   auto-derived\nItems:  %d\nPath:   %s\n", s.Alias, s.ItemCount, s.FolderPath)

			return
		}
	}

	fmt.Fprintf(os.Stderr, "❌ Alias %q not found\n", alias)
}

func runAliasSuggest(cmd *cobra.Command, args []string) {
	database := openDatabaseForAlias()
	defer database.Close()

	stats, err := database.ListScanFoldersWithStats()

	if err != nil || len(stats) == 0 {
		fmt.Println("📭 No scanned folders found to suggest aliases for.")

		return
	}

	fmt.Printf("Suggesting aliases for %d scanned root folder(s):\n", len(stats))

	createdCount := 0

	for _, s := range stats {
		derived := db.DeriveCleanAlias(s.FolderPath)
		fmt.Printf("  %-20s -> %-10s (%d items)\n", s.FolderPath, derived, s.ItemCount)

		if aliasApplyFlag {
			if saveErr := database.SetFolderAlias(derived, s.FolderPath); saveErr == nil {
				createdCount++
			}
		}
	}

	if aliasApplyFlag {
		fmt.Printf("\n✔ %d alias(es) saved successfully.\n", createdCount)

		return
	}

	fmt.Println("\n💡 To save these aliases, run: movie alias suggest --apply")
}
