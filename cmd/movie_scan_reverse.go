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
	Drift             *driftLog
	SidecarsRewritten int
	OrphansRemoved    int
	DeletedPurged     int
	MissingDetected   int
}

// reverseSyncCtx bundles the per-run state passed to each step. Keeps
// step signatures within the 3-parameter project rule.
type reverseSyncCtx struct {
	Database   *db.DB
	DiskSet    map[string]struct{}
	SidecarSet map[string]struct{}
	Drift      *driftLog
	JsonRoot   string
}

// runReverseSync is the public entry point invoked from movie scan.
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
	res := &reverseSyncResult{Drift: &driftLog{}}
	ctx := &reverseSyncCtx{
		Database:   database,
		JsonRoot:   jsonRoot,
		DiskSet:    buildDiskSet(scanDir),
		SidecarSet: indexSidecars(jsonRoot),
		Drift:      res.Drift,
	}
	executeReverseSyncSteps(ctx, rows, res)
	printReverseSyncSummary(database, since, res)
	return res
}

func executeReverseSyncSteps(ctx *reverseSyncCtx, rows []db.ReverseSyncRow, res *reverseSyncResult) {
	res.SidecarsRewritten = reverseStepRewrite(ctx, rows)
	res.DeletedPurged = reverseStepPurgeDeleted(ctx, rows)
	res.MissingDetected = reverseStepDetectMissing(ctx, rows)
	res.OrphansRemoved = reverseStepRemoveOrphans(ctx, rows)
}

func indexSidecars(jsonRoot string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, p := range listJsonSidecars(jsonRoot) {
		set[p] = struct{}{}
	}
	return set
}

// R4: rewrite sidecars whose mtime is older than the DB UpdatedAt.
func reverseStepRewrite(ctx *reverseSyncCtx, rows []db.ReverseSyncRow) int {
	count := 0
	for i := range rows {
		r := &rows[i]
		if r.IsDeleted || r.CurrentFilePath == "" {
			continue
		}
		if _, ok := ctx.DiskSet[r.CurrentFilePath]; !ok {
			continue
		}
		sidecarPath := sidecarPathFor(ctx.Database, ctx.JsonRoot, r)
		reason, needsWrite := classifySidecar(sidecarPath, r.UpdatedAt)
		if !needsWrite {
			continue
		}
		if err := writeSidecarFromDB(ctx.Database, ctx.JsonRoot, r); err != nil {
			errlog.Warn("reverse-sync: rewrite #%d: %v", r.ID, err)
			continue
		}
		ctx.Drift.add(reason, r.ID, sidecarPath)
		_, _ = ctx.Database.InsertReconciliation(db.ReconInput{
			MediaID:    sql.NullInt64{Int64: r.ID, Valid: true},
			ActionType: db.ReconActionReverseSyncedSidecar,
			Details:    sidecarPath,
		})
		ctx.SidecarSet[sidecarPath] = struct{}{}
		count++
	}
	return count
}

// classifySidecar returns the drift reason and whether a rewrite is needed.
func classifySidecar(sidecarPath, dbUpdatedAt string) (driftReason, bool) {
	info, err := os.Stat(sidecarPath)
	if err != nil {
		return driftMissingSidecar, true
	}
	dbTime, parseErr := parseDBTime(dbUpdatedAt)
	if parseErr != nil {
		return driftStaleMtime, true
	}
	if info.ModTime().Before(dbTime) {
		return driftStaleMtime, true
	}
	return driftStaleMtime, false
}

// R6: purge sidecars belonging to soft-deleted DB rows.
func reverseStepPurgeDeleted(ctx *reverseSyncCtx, rows []db.ReverseSyncRow) int {
	count := 0
	for i := range rows {
		r := &rows[i]
		if !r.IsDeleted {
			continue
		}
		sidecarPath := sidecarPathFor(ctx.Database, ctx.JsonRoot, r)
		if _, ok := ctx.SidecarSet[sidecarPath]; !ok {
			continue
		}
		if !removeIfDryRunOff(sidecarPath) {
			continue
		}
		delete(ctx.SidecarSet, sidecarPath)
		ctx.Drift.add(driftSoftDeletedRow, r.ID, sidecarPath)
		_, _ = ctx.Database.InsertReconciliation(db.ReconInput{
			MediaID:    sql.NullInt64{Int64: r.ID, Valid: true},
			ActionType: db.ReconActionRemovedDeletedSidecar,
			Details:    sidecarPath,
		})
		count++
	}
	return count
}

// R7: mark active rows whose disk file disappeared as missing + delete sidecar.
func reverseStepDetectMissing(ctx *reverseSyncCtx, rows []db.ReverseSyncRow) int {
	count := 0
	for i := range rows {
		r := &rows[i]
		if r.IsDeleted || r.CurrentFilePath == "" {
			continue
		}
		if _, ok := ctx.DiskSet[r.CurrentFilePath]; ok {
			continue
		}
		ctx.Drift.add(driftMissingDiskFile, r.ID, r.CurrentFilePath)
		if scanDryRun {
			fmt.Printf("   [dry-run] would mark missing: #%d %s\n", r.ID, r.CurrentFilePath)
			count++
			continue
		}
		if err := ctx.Database.MarkMediaMissing(r.ID); err != nil {
			errlog.Warn("reverse-sync: mark missing #%d: %v", r.ID, err)
			continue
		}
		_ = os.Remove(sidecarPathFor(ctx.Database, ctx.JsonRoot, r))
		_, _ = ctx.Database.InsertReconciliation(db.ReconInput{
			MediaID:    sql.NullInt64{Int64: r.ID, Valid: true},
			ActionType: db.ReconActionReverseDetectedMissing,
			Details:    r.CurrentFilePath,
		})
		count++
	}
	return count
}

// R5: remove orphan sidecars (no DB row, file gone from disk).
func reverseStepRemoveOrphans(ctx *reverseSyncCtx, rows []db.ReverseSyncRow) int {
	known := indexKnownSidecars(ctx.Database, ctx.JsonRoot, rows)
	count := 0
	for sidecar := range ctx.SidecarSet {
		if _, ok := known[sidecar]; ok {
			continue
		}
		if hasMatchingDiskFile(sidecar, ctx.DiskSet) {
			continue
		}
		if !removeIfDryRunOff(sidecar) {
			continue
		}
		ctx.Drift.add(driftDiskOrphanFile, 0, sidecar)
		_, _ = ctx.Database.InsertReconciliation(db.ReconInput{
			ActionType: db.ReconActionRemovedOrphanSidecar,
			Details:    sidecar,
		})
		count++
	}
	return count
}

func indexKnownSidecars(database *db.DB, jsonRoot string, rows []db.ReverseSyncRow) map[string]struct{} {
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
	printDriftSummary(res.Drift)
	counts, err := database.CountReconciliationByType(since)
	if err == nil && len(counts) > 0 {
		fmt.Printf("   audit: %v\n", counts)
	}
}
