// movie_scan_reconcile.go — SmartRescan orchestrator (Phase 5).
//
// Spec: spec/08-app/10-remove-move-rescan/rescan-reconciliation/01-spec.md
//
// Triggered automatically before every `movie scan` when
// <folder>/.movie-output/json contains sidecar files. Reconciles three
// sources of truth (disk ⇄ JSON ⇄ DB) with ZERO TMDb calls for the
// converged majority. Opt-out: --no-reconcile.
package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
)

var scanNoReconcile bool

// reconcileResult is the summary returned to the scan footer.
type reconcileResult struct {
	Hydrated  int
	Missing   int
	Converged int
	NewPaths  []string // paths to feed to executeScan; empty = nothing new
}

// runSmartRescan is the public entry point. Returns nil result when the
// pre-scan check fails or --no-reconcile is set; caller falls through.
func runSmartRescan(database *db.DB, scanDir string) *reconcileResult {
	if scanNoReconcile {
		return nil
	}
	jsonRoot := filepath.Join(scanDir, ".movie-output", "json")
	jsonPaths := listJsonSidecars(jsonRoot)
	if len(jsonPaths) == 0 {
		return nil
	}
	since := time.Now().UTC().Format(time.RFC3339)
	res := &reconcileResult{}
	executeReconcileSteps(database, scanDir, jsonPaths, res)
	printReconcileSummary(database, since, res)
	return res
}

func executeReconcileSteps(database *db.DB, scanDir string, jsonPaths []string, res *reconcileResult) {
	jsonItems := loadJsonItems(jsonPaths)
	dbItems, _ := database.GetMediaByScanDir(scanDir)
	diskSet := buildDiskSet(scanDir)

	if len(dbItems) == 0 && len(jsonItems) > 0 {
		res.Hydrated = hydrateAllFromJson(database, jsonItems)
		dbItems, _ = database.GetMediaByScanDir(scanDir)
	}
	res.Missing = markMissingItems(database, dbItems, diskSet, jsonRootOf(scanDir))
	res.NewPaths, res.Converged = classifyDiskPaths(database, dbItems, diskSet)
}

func jsonRootOf(scanDir string) string {
	return filepath.Join(scanDir, ".movie-output", "json")
}

// ---- step 1+2: disk + json discovery ----

func buildDiskSet(scanDir string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, vf := range collectVideoFiles(scanDir, scanRecursive, scanDepth) {
		out[vf.FullPath] = struct{}{}
	}
	return out
}

func listJsonSidecars(jsonRoot string) []string {
	var out []string
	_ = filepath.WalkDir(jsonRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		out = append(out, path)
		return nil
	})
	return out
}

// ---- step 5: missing detection ----

func markMissingItems(database *db.DB, items []db.Media, diskSet map[string]struct{}, jsonRoot string) int {
	count := 0
	for i := range items {
		if items[i].CurrentFilePath == "" {
			continue
		}
		if _, ok := diskSet[items[i].CurrentFilePath]; ok {
			continue
		}
		applyMissingMark(database, &items[i], jsonRoot)
		count++
	}
	return count
}

func applyMissingMark(database *db.DB, m *db.Media, jsonRoot string) {
	if updErr := database.MarkMediaMissing(m.ID); updErr != nil {
		errlog.Warn("recon: mark missing #%d: %v", m.ID, updErr)
	}
	deleteSidecarFor(jsonRoot, m)
	_, _ = database.InsertReconciliation(db.ReconInput{
		MediaID:    sql.NullInt64{Int64: m.ID, Valid: true},
		ActionType: db.ReconActionRemovedMissing,
		Details:    m.CurrentFilePath,
	})
}

func deleteSidecarFor(jsonRoot string, m *db.Media) {
	subDir := db.JsonSubDir(m.Type)
	slug := mediaSlug(m)
	sidecar := filepath.Join(jsonRoot, subDir, slug+".json")
	_ = os.Remove(sidecar)
}

// ---- step 6+7: classify disk paths ----

func classifyDiskPaths(database *db.DB, dbItems []db.Media, diskSet map[string]struct{}) ([]string, int) {
	known := make(map[string]struct{}, len(dbItems))
	for i := range dbItems {
		if dbItems[i].CurrentFilePath != "" {
			known[dbItems[i].CurrentFilePath] = struct{}{}
		}
	}
	var newPaths []string
	converged := 0
	for path := range diskSet {
		if _, ok := known[path]; ok {
			converged++
			continue
		}
		newPaths = append(newPaths, path)
	}
	if converged > 0 {
		_, _ = database.InsertReconciliation(db.ReconInput{
			ActionType: db.ReconActionConverged,
			Details:    fmt.Sprintf("%d files converged", converged),
		})
	}
	return newPaths, converged
}

func printReconcileSummary(database *db.DB, since string, res *reconcileResult) {
	fmt.Printf("🔄 SmartRescan: hydrated=%d missing=%d converged=%d new=%d\n",
		res.Hydrated, res.Missing, res.Converged, len(res.NewPaths))
	counts, err := database.CountReconciliationByType(since)
	if err == nil && len(counts) > 0 {
		fmt.Printf("   audit: %v\n", counts)
	}
}
