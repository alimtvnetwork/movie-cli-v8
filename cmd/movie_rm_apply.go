// movie_rm_apply.go — preview, confirm, and apply soft-delete actions.
package cmd

import (
	"bufio"
	"fmt"
	"os"
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
	if count < rmConfirmThreshold {
		return true
	}
	fmt.Printf("\nProceed with soft-delete of %d items? [y/N]: ", count)
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
	success := 0
	for _, id := range ids {
		if applySingleRm(database, id) {
			success++
		}
	}
	fmt.Printf("\n✅ Soft-deleted %d/%d media.\n", success, len(ids))
}

func applySingleRm(database *db.DB, id int64) bool {
	m, err := database.GetMediaByID(id)
	if err != nil {
		errlog.Warn("rm: load #%d: %v", id, err)
		return false
	}
	if delErr := database.SoftDeleteMedia(id); delErr != nil {
		errlog.Error("rm: soft-delete #%d: %v", id, delErr)
		return false
	}
	logRmHistory(database, m)
	return true
}

func logRmHistory(database *db.DB, m *db.Media) {
	histErr := database.InsertMoveHistory(db.MoveInput{
		MediaID:      m.ID,
		FileActionID: int(db.FileActionDelete),
		FromPath:     m.CurrentFilePath,
		ToPath:       "",
		OrigName:     m.OriginalFileName,
		NewName:      "",
	})
	if histErr != nil {
		errlog.Warn("rm: history #%d: %v", m.ID, histErr)
	}
}
