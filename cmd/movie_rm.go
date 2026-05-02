// movie_rm.go — `movie rm` / `movie remove` / `movie delete`
//
// Soft-deletes media rows by ID, title, or condition expression.
// Spec: spec/08-app/10-remove-move-rescan/remove-command/01-spec.md
//
// Resolution modes (auto-detected from arg shape):
//  1. Numeric arg          → single MediaId
//  2. Title (no operator)  → resolveMediaByQuery
//  3. Quoted expression    → BuildConditionSQL → bulk match
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var (
	rmAssumeYes bool
	rmPurge     bool
)

var movieRmCmd = &cobra.Command{
	Use:     "rm <id|title|\"expr\">",
	Aliases: []string{"remove", "delete"},
	Short:   "Soft-delete media by id, title, or condition expression",
	Long: `Soft-delete media. Aliases: remove, delete.

Examples:
  movie rm 42
  movie rm "Inception"
  movie rm "rating < 5 AND year >= 2010"
  movie rm "g = Horror" --yes
  movie rm 42 --purge          # also delete the on-disk video file

The --purge flag removes the underlying file from disk in addition to
the soft-delete. The DB row is still recoverable via 'movie undo'
(but the file itself is gone — undo cannot restore it).
`,
	Args: cobra.MinimumNArgs(1),
	Run:  runMovieRm,
}

func init() {
	movieRmCmd.Flags().BoolVarP(&rmAssumeYes, "yes", "y", false, "Skip confirmation prompt")
	movieRmCmd.Flags().BoolVar(&rmPurge, "purge", false, "Also delete the on-disk file (irreversible)")
}

func runMovieRm(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	query := strings.Join(args, " ")
	ids, err := resolveRmTargets(database, query)
	if err != nil {
		errlog.Error("rm resolve: %v", err)
		return
	}
	if len(ids) == 0 {
		fmt.Println("📭 No matching media.")
		return
	}
	previewRmTargets(database, ids)
	if !confirmRm(len(ids)) {
		return
	}
	applyRm(database, ids)
}

// resolveRmTargets picks the correct resolution strategy for the query.
func resolveRmTargets(database *db.DB, query string) ([]int64, error) {
	if isConditionExpression(query) {
		return resolveByCondition(database, query)
	}
	m, err := resolveMediaByQuery(database, query)
	if err != nil {
		return nil, err
	}
	return []int64{m.ID}, nil
}

func isConditionExpression(q string) bool {
	for _, op := range []string{"<", ">", "=", "!="} {
		if strings.Contains(q, op) {
			return true
		}
	}
	return false
}

func resolveByCondition(database *db.DB, expr string) ([]int64, error) {
	where, args, err := BuildConditionSQL(expr)
	if err != nil {
		return nil, apperror.Wrap("parse expression", err)
	}
	return database.QueryMediaIDsByWhere(where, args)
}
