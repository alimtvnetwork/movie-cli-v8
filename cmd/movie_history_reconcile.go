// movie_history_reconcile.go — `movie history reconcile`
//
// Audits ReconciliationHistory grouped by scanDir × ActionType.
// Spec: spec/08-app/10-remove-move-rescan/rescan-reconciliation/01-spec.md
//
// scanDir is derived from the linked Media.CurrentFilePath:
// the parent directory before `.movie-output`, falling back to the file's
// dirname. Rows without a MediaId (Converged summary) are bucketed under
// the literal label "(global)".
package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
)

const (
	reconcileGlobalBucket = "(global)"
	reconcileDefaultLimit = 500
)

var (
	reconcileLimit  int
	reconcileFilter string
)

var movieHistoryReconcileCmd = &cobra.Command{
	Use:     "reconcile",
	Aliases: []string{"recon", "rescan-audit"},
	Short:   "Audit SmartRescan results grouped by scanDir × action type",
	Long: `Lists ReconciliationHistory rows grouped by scan directory and
SmartRescan action type (HydratedFromJson, RemovedMissing, AddedNew, Converged).

Examples:
  movie history reconcile
  movie history reconcile --limit 1000
  movie history reconcile --dir ~/Movies
`,
	Run: runMovieHistoryReconcile,
}

func init() {
	movieHistoryReconcileCmd.Flags().IntVar(&reconcileLimit, "limit", reconcileDefaultLimit,
		"max ReconciliationHistory rows to scan")
	movieHistoryReconcileCmd.Flags().StringVar(&reconcileFilter, "dir", "",
		"only show rows whose scanDir matches this prefix")
	movieHistoryCmd.AddCommand(movieHistoryReconcileCmd)
}

// reconcileBucket aggregates counts for one (scanDir, action) pair.
type reconcileBucket struct {
	ScanDir string
	Action  string
	Latest  string
	Count   int
}

func runMovieHistoryReconcile(cmd *cobra.Command, args []string) {
	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	rows, listErr := database.ListReconciliation(reconcileLimit)
	if listErr != nil {
		errlog.Error("recon: list: %v", listErr)
		return
	}
	if len(rows) == 0 {
		fmt.Println("📭 No SmartRescan history yet.")
		return
	}
	buckets := groupReconcile(database, rows)
	if len(buckets) == 0 {
		fmt.Println("📭 No rows match the --dir filter.")
		return
	}
	printReconcileBuckets(buckets, len(rows))
}

func groupReconcile(database *db.DB, rows []db.ReconRecord) []reconcileBucket {
	agg := make(map[string]*reconcileBucket)
	for i := range rows {
		scanDir := scanDirForRow(database, &rows[i])
		if !matchesReconcileFilter(scanDir) {
			continue
		}
		key := scanDir + "\x1f" + db.ReconActionName(rows[i].ReconciliationActionTypeId)
		bucket := ensureBucket(agg, key, scanDir, db.ReconActionName(rows[i].ReconciliationActionTypeId))
		bucket.Count++
		if rows[i].OccurredAt > bucket.Latest {
			bucket.Latest = rows[i].OccurredAt
		}
	}
	return sortedBuckets(agg)
}

func ensureBucket(agg map[string]*reconcileBucket, key, scanDir, action string) *reconcileBucket {
	bucket, ok := agg[key]
	if ok {
		return bucket
	}
	bucket = &reconcileBucket{ScanDir: scanDir, Action: action}
	agg[key] = bucket
	return bucket
}

func matchesReconcileFilter(scanDir string) bool {
	if reconcileFilter == "" {
		return true
	}
	return strings.HasPrefix(scanDir, reconcileFilter)
}

func scanDirForRow(database *db.DB, r *db.ReconRecord) string {
	if !r.MediaId.Valid {
		return reconcileGlobalBucket
	}
	m, err := database.GetMediaByID(r.MediaId.Int64)
	if err != nil || m == nil || m.CurrentFilePath == "" {
		return reconcileGlobalBucket
	}
	return deriveScanDir(m.CurrentFilePath)
}

// deriveScanDir walks up until it finds the parent of `.movie-output`,
// otherwise returns the file's parent directory.
func deriveScanDir(path string) string {
	dir := filepath.Dir(path)
	cursor := dir
	for cursor != "/" && cursor != "." {
		if _, err := filepath.Glob(filepath.Join(cursor, ".movie-output")); err == nil {
			return cursor
		}
		parent := filepath.Dir(cursor)
		if parent == cursor {
			break
		}
		cursor = parent
	}
	return dir
}

func sortedBuckets(agg map[string]*reconcileBucket) []reconcileBucket {
	out := make([]reconcileBucket, 0, len(agg))
	for _, b := range agg {
		out = append(out, *b)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ScanDir != out[j].ScanDir {
			return out[i].ScanDir < out[j].ScanDir
		}
		return out[i].Action < out[j].Action
	})
	return out
}

func printReconcileBuckets(buckets []reconcileBucket, totalRows int) {
	fmt.Printf("🔄 SmartRescan audit — %d rows scanned, %d buckets\n\n", totalRows, len(buckets))
	fmt.Printf("  %-50s  %-18s  %5s  %s\n", "SCAN DIR", "ACTION", "COUNT", "LATEST")
	fmt.Printf("  %s\n", strings.Repeat("-", 100))
	for i := range buckets {
		fmt.Printf("  %-50s  %-18s  %5d  %s\n",
			truncateLeft(buckets[i].ScanDir, 50),
			buckets[i].Action,
			buckets[i].Count,
			buckets[i].Latest)
	}
}

func truncateLeft(s string, width int) string {
	if len(s) <= width {
		return s
	}
	return "…" + s[len(s)-width+1:]
}
