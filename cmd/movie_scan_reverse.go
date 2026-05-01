// movie_scan_reverse.go — Reverse-sync orchestrator (DB → JSON → disk).
//
// Spec: spec/08-app/10-remove-move-rescan/rescan-reconciliation/02-reverse-sync-spec.md
// Diagram: spec/06-diagrams/17-reverse-sync-flow.mmd
//
// Runs after the forward SmartRescan pass. Treats the SQLite DB as the
// authoritative source and reconciles JSON sidecars + disk awareness to
// match. Zero TMDb calls. Opt-out: --no-reverse-sync.
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

var scanNoReverseSync bool
var scanReverseSyncOnly bool

type reverseSyncResult struct {
	SidecarsRewritten int
	OrphansRemoved    int
	DeletedPurged     int
	MissingDetected   int
}

// runReverseSync is the public entry point invoked from movie scan.
// Returns nil when disabled or no work is possible (no JSON folder).
func runReverseSync(database *db.DB, scanDir string) *reverseSyncResult {
	if scanNoReverseSync && !scanReverseSyncOnly {
		return nil
	}
	jsonRoot := filepath.Join(scanDir, ".movie-output", "json")
	if _, err := os.Stat(jsonRoot); err != nil && os.IsNotExist(err) {
		_ = os.MkdirAll(jsonRoot, 0755)
	}
	rows, err := database.ListReverseSyncRows(scanDir)
	if err != nil {
		errlog.Warn("reverse-sync: list rows: %v", err)
		return nil
	}
	since := time.Now().UTC().Format(time.RFC3339)
	res := &reverseSyncResult{}
	executeReverseSyncSteps(database, scanDir, jsonRoot, rows, res)
	printReverseSyncSummary(database, since, res)
	return res
}

func executeReverseSyncSteps(database *db.DB, scanDir, jsonRoot string,
	rows []db.ReverseSyncRow, res *reverseSyncResult) {
	diskSet := buildDiskSet(scanDir)
	sidecarSet := indexSidecars(jsonRoot)

	res.SidecarsRewritten = reverseStepRewrite(database, jsonRoot, rows, diskSet, sidecarSet)
	res.DeletedPurged = reverseStepPurgeDeleted(database, jsonRoot, rows, sidecarSet)
	res.MissingDetected = reverseStepDetectMissing(database, jsonRoot, rows, diskSet)
	res.OrphansRemoved = reverseStepRemoveOrphans(database, jsonRoot, sidecarSet, diskSet, rows)
}

// indexSidecars returns absolute sidecar path → struct{}.
func indexSidecars(jsonRoot string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, p := range listJsonSidecars(jsonRoot) {
		set[p] = struct{}{}
	}
	return set
}

// R4: rewrite sidecars whose mtime is older than the DB UpdatedAt.
func reverseStepRewrite(database *db.DB, jsonRoot string,
	rows []db.ReverseSyncRow, diskSet, sidecarSet map[string]struct{}) int {
	count := 0
	for i := range rows {
		r := &rows[i]
		if r.IsDeleted || r.CurrentFilePath == "" {
			continue
		}
		if _, ok := diskSet[r.CurrentFilePath]; !ok {
			continue
		}
		sidecarPath := sidecarPathFor(database, jsonRoot, r)
		if !shouldRewriteSidecar(sidecarPath, r.UpdatedAt) {
			continue
		}
		if err := writeSidecarFromDB(database, jsonRoot, r); err != nil {
			errlog.Warn("reverse-sync: rewrite #%d: %v", r.ID, err)
			continue
		}
		_, _ = database.InsertReconciliation(db.ReconInput{
			MediaID:    sql.NullInt64{Int64: r.ID, Valid: true},
			ActionType: db.ReconActionReverseSyncedSidecar,
			Details:    sidecarPath,
		})
		sidecarSet[sidecarPath] = struct{}{}
		count++
	}
	return count
}

// R6: purge sidecars belonging to soft-deleted DB rows.
func reverseStepPurgeDeleted(database *db.DB, jsonRoot string,
	rows []db.ReverseSyncRow, sidecarSet map[string]struct{}) int {
	count := 0
	for i := range rows {
		r := &rows[i]
		if !r.IsDeleted {
			continue
		}
		sidecarPath := sidecarPathFor(database, jsonRoot, r)
		if _, ok := sidecarSet[sidecarPath]; !ok {
			continue
		}
		if removeIfDryRunOff(sidecarPath) {
			delete(sidecarSet, sidecarPath)
			_, _ = database.InsertReconciliation(db.ReconInput{
				MediaID:    sql.NullInt64{Int64: r.ID, Valid: true},
				ActionType: db.ReconActionRemovedDeletedSidecar,
				Details:    sidecarPath,
			})
			count++
		}
	}
	return count
}

// R7: mark active rows whose disk file disappeared as missing + delete sidecar.
func reverseStepDetectMissing(database *db.DB, jsonRoot string,
	rows []db.ReverseSyncRow, diskSet map[string]struct{}) int {
	count := 0
	for i := range rows {
		r := &rows[i]
		if r.IsDeleted || r.CurrentFilePath == "" {
			continue
		}
		if _, ok := diskSet[r.CurrentFilePath]; ok {
			continue
		}
		if scanDryRun {
			fmt.Printf("   [dry-run] would mark missing: #%d %s\n", r.ID, r.CurrentFilePath)
			count++
			continue
		}
		if err := database.MarkMediaMissing(r.ID); err != nil {
			errlog.Warn("reverse-sync: mark missing #%d: %v", r.ID, err)
			continue
		}
		_ = os.Remove(sidecarPathFor(database, jsonRoot, r))
		_, _ = database.InsertReconciliation(db.ReconInput{
			MediaID:    sql.NullInt64{Int64: r.ID, Valid: true},
			ActionType: db.ReconActionReverseDetectedMissing,
			Details:    r.CurrentFilePath,
		})
		count++
	}
	return count
}

// R5: remove orphan sidecars (no DB row, file gone from disk).
func reverseStepRemoveOrphans(database *db.DB, jsonRoot string,
	sidecarSet, diskSet map[string]struct{}, rows []db.ReverseSyncRow) int {
	known := indexKnownSidecars(jsonRoot, rows)
	count := 0
	for sidecar := range sidecarSet {
		if _, ok := known[sidecar]; ok {
			continue
		}
		if hasMatchingDiskFile(sidecar, diskSet) {
			continue
		}
		if !removeIfDryRunOff(sidecar) {
			continue
		}
		_, _ = database.InsertReconciliation(db.ReconInput{
			ActionType: db.ReconActionRemovedOrphanSidecar,
			Details:    sidecar,
		})
		count++
	}
	return count
}

func indexKnownSidecars(jsonRoot string, rows []db.ReverseSyncRow) map[string]struct{} {
	set := make(map[string]struct{}, len(rows))
	for i := range rows {
		set[sidecarPathFor(database, jsonRoot, &rows[i])] = struct{}{}
	}
	return set
}

// hasMatchingDiskFile returns true when the sidecar appears to belong
// to a video file currently on disk (best-effort: matches slug to any
// disk filename, since orphan sidecars often arrive without DB rows).
func hasMatchingDiskFile(sidecarPath string, diskSet map[string]struct{}) bool {
	slug := strings.TrimSuffix(filepath.Base(sidecarPath), ".json")
	for diskPath := range diskSet {
		if strings.Contains(strings.ToLower(filepath.Base(diskPath)), slug) {
			return true
		}
	}
	return false
}

func removeIfDryRunOff(path string) bool {
	if scanDryRun {
		fmt.Printf("   [dry-run] would remove sidecar: %s\n", path)
		return false
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		errlog.Warn("reverse-sync: remove %s: %v", path, err)
		return false
	}
	return true
}

func printReverseSyncSummary(database *db.DB, since string, res *reverseSyncResult) {
	fmt.Printf("↩️  ReverseSync: rewritten=%d orphans=%d deleted=%d missing=%d\n",
		res.SidecarsRewritten, res.OrphansRemoved, res.DeletedPurged, res.MissingDetected)
	counts, err := database.CountReconciliationByType(since)
	if err == nil && len(counts) > 0 {
		fmt.Printf("   audit: %v\n", counts)
	}
}
