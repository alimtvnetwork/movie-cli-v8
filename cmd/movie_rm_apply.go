// movie_rm_apply.go — preview, confirm, and apply soft-delete actions.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
)

const rmConfirmThreshold = 5

func previewRmTargets(database *db.DB, ids []int64) {
	fmt.Printf("🗑  %d media will be soft-deleted:\n\n", len(ids))
	limit := len(ids)
	if limit > 20 {
		limit = 20
	}
	for i := 0; i < limit; i++ {
		printRmRow(database, ids[i])
	}
	if len(ids) > limit {
		fmt.Printf("  … and %d more\n", len(ids)-limit)
	}
}

func printRmRow(database *db.DB, id int64) {
	m, err := database.GetMediaByID(id)
	if err != nil {
		fmt.Printf("  - #%d (load error)\n", id)
		return
	}
	fmt.Printf("  - #%d  %s (%d)\n", m.ID, m.Title, m.Year)
}

func confirmRm(count int) bool {
	if rmAssumeYes {
		return true
	}
	if count < rmConfirmThreshold && !rmPurge {
		return true
	}
	verb := "soft-delete"
	if rmPurge {
		verb = "PURGE (delete files from disk)"
	}
	fmt.Printf("\nProceed with %s of %d items? [y/N]: ", verb, count)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("❌ Canceled.")
		return false
	}
	return true
}

func applyRm(database *db.DB, ids []int64) {
	batchID := generateBatchID()
	success := 0
	for _, id := range ids {
		if applySingleRm(database, id, batchID) {
			success++
		}
	}
	fmt.Printf("\n✅ Soft-deleted %d/%d media (batch %s).\n", success, len(ids), batchID[:8])
	if success > 0 {
		regenerateReports(database)
	}
}

func applySingleRm(database *db.DB, id int64, batchID string) bool {
	m, err := database.GetMediaByID(id)
	if err != nil {
		errlog.Warn("rm: load #%d: %v", id, err)
		return false
	}
	if delErr := database.SoftDeleteMedia(id); delErr != nil {
		errlog.Error("rm: soft-delete #%d: %v", id, delErr)
		return false
	}
	removeRmSidecar(m)
	purgeOnDiskFile(m)
	logRmHistory(database, m, batchID)
	return true
}

// purgeOnDiskFile deletes the underlying video file when --purge is set.
// Silently no-ops when the flag is off or the path is empty/missing.
func purgeOnDiskFile(m *db.Media) {
	if !rmPurge {
		return
	}
	if m.CurrentFilePath == "" {
		return
	}
	if err := os.Remove(m.CurrentFilePath); err != nil && !os.IsNotExist(err) {
		errlog.Warn("rm --purge: delete %s: %v", m.CurrentFilePath, err)
		return
	}
	fmt.Printf("  🔥 purged file: %s\n", m.CurrentFilePath)
}

func removeRmSidecar(m *db.Media) {
	if m.CurrentFilePath == "" {
		return
	}
	jsonRoot := filepath.Join(filepath.Dir(m.CurrentFilePath), ".movie-output", "json")
	deleteSidecarFor(jsonRoot, m)
}

func logRmHistory(database *db.DB, m *db.Media, batchID string) {
	snap, snapErr := db.MediaToJSON(m)
	if snapErr != nil {
		errlog.Warn("rm: snapshot #%d: %v", m.ID, snapErr)
		return
	}
	_, histErr := database.InsertActionSimple(db.ActionSimpleInput{
		FileAction: db.FileActionDelete,
		MediaID:    m.ID,
		Snapshot:   snap,
		BatchID:    batchID,
		Detail:     fmt.Sprintf("Soft-deleted: %s (%d)", m.Title, m.Year),
	})
	if histErr != nil {
		errlog.Warn("rm: history #%d: %v", m.ID, histErr)
	}
}
