// quarantine_prompt.go — Quarantine management into temp-remove and terminal exit purge confirmation.
package cmd

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

// moveFileOrDir moves a file or directory safely across paths.
func moveFileOrDir(source, destination string) error {
	destDir := filepath.Dir(destination)

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return appfault.Wrapf(err, "create destination directory %s", destDir)
	}

	renErr := os.Rename(source, destination)

	if renErr == nil {
		return nil
	}

	// Fallback for cross-device moves
	if moveErr := MoveFile(source, destination); moveErr != nil {
		return appfault.Wrapf(moveErr, "move %s to %s", source, destination)
	}

	return nil
}

// sanitizeQuarantineName sanitizes a string for use in folder names.
func sanitizeQuarantineName(name string) string {
	var clean strings.Builder

	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			clean.WriteRune(r)
		} else if r == ' ' || r == '.' {
			clean.WriteRune('_')
		}
	}

	result := clean.String()

	if len(result) == 0 {
		return "item"
	}

	if len(result) > 40 {
		return result[:40]
	}

	return result
}

// executeQuarantineRemoval safely moves target files/folders to temp-remove instead of immediate deletion.
func executeQuarantineRemoval(database *db.DB, rec *db.StagedActionRecord, batchID string) error {
	if valErr := validateRemovalTarget(rec.SourcePath, database); valErr != nil {
		return valErr
	}

	if _, statErr := os.Stat(rec.SourcePath); statErr != nil {
		if os.IsNotExist(statErr) {
			if rec.MediaId.Valid {
				_ = database.SoftDeleteMedia(rec.MediaId.Int64)
			}

			return nil
		}

		return appfault.Wrapf(statErr, "stat source %s", rec.SourcePath)
	}

	scanRoot := findContainingScanRoot(rec.SourcePath, database)

	if len(scanRoot) == 0 {
		scanRoot = filepath.Dir(rec.SourcePath)
	}

	quarantineDir := filepath.Join(scanRoot, "temp-remove")
	_ = os.MkdirAll(quarantineDir, 0o755)

	itemTitle := "item"

	if rec.MediaId.Valid {
		if m, err := database.GetMediaByID(rec.MediaId.Int64); err == nil && m != nil {
			itemTitle = m.Title
			removeRmSidecar(m)
		}
	}

	folderTag := fmt.Sprintf("%s_%d_%s", time.Now().Format("20060102_150405"), rec.MediaId.Int64, sanitizeQuarantineName(itemTitle))
	itemQuarantine := filepath.Join(quarantineDir, folderTag)
	_ = os.MkdirAll(itemQuarantine, 0o755)

	targetBase := filepath.Base(rec.SourcePath)
	destination := filepath.Join(itemQuarantine, targetBase)

	if moveErr := moveFileOrDir(rec.SourcePath, destination); moveErr != nil {
		return appfault.Wrapf(moveErr, "move to quarantine %s", itemQuarantine)
	}

	// Move sidecars if source was a single file
	sidecars := findSidecarFiles(rec.SourcePath)

	for _, sc := range sidecars {
		scBase := filepath.Base(sc)
		_ = moveFileOrDir(sc, filepath.Join(itemQuarantine, scBase))
	}

	if rec.MediaId.Valid {
		if softErr := database.SoftDeleteMedia(rec.MediaId.Int64); softErr != nil {
			return appfault.Wrapf(softErr, "soft delete media #%d", rec.MediaId.Int64)
		}

		_, _ = database.InsertActionSimple(db.ActionSimpleInput{
			FileAction: db.FileActionDelete,
			MediaID:    rec.MediaId.Int64,
			Snapshot:   rec.MediaSnapshot,
			Detail:     fmt.Sprintf("quarantined to %s", itemQuarantine),
			BatchID:    batchID,
		})
	}

	task := &db.TaskRecord{
		TaskType:       db.TaskTypeQuarantineMove,
		MediaId:        rec.MediaId,
		SourcePath:     rec.SourcePath,
		QuarantinePath: itemQuarantine,
		TargetPath:     destination,
		Status:         db.TaskQuarantined,
		BatchId:        batchID,
		ItemTitle:      itemTitle,
		Payload:        rec.MediaSnapshot,
	}

	_, _ = database.InsertTask(task)

	return nil
}

// promptPendingQuarantinePurge inspects quarantined tasks and prompts the user in the terminal upon UI exit.
func promptPendingQuarantinePurge(database *db.DB) {
	if database == nil {
		return
	}

	tasks, err := database.ListPendingQuarantineTasks()

	if err != nil {
		return
	}

	if len(tasks) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("  ╭────────────────────────────────────────────────────────────────────────────╮")
	fmt.Println("  │ ⚠️  PENDING PERMANENT REMOVAL (temp-remove quarantine)                     │")
	fmt.Println("  ╰────────────────────────────────────────────────────────────────────────────╯")
	fmt.Println("  The following movies were approved for removal and moved to quarantine:")

	var sampleTitle string

	if len(tasks) > 0 {
		randIdx := rand.Intn(len(tasks))
		sampleTitle = tasks[randIdx].ItemTitle
	}

	for i := range tasks {
		t := &tasks[i]
		fmt.Printf("    %2d. %-35s  • Quarantine: %s\n", i+1, t.ItemTitle, t.QuarantinePath)
	}

	fmt.Println()
	fmt.Println("  ⚠️  Permanent deletion is NON-REVERSIBLE.")
	fmt.Printf("  To permanently delete all items in temp-remove, type %q (or 'CONFIRM')\n", sampleTitle)
	fmt.Print("  [or press Enter to keep them safely in quarantine]: ")

	scanner := bufio.NewScanner(os.Stdin)
	hasScanned := scanner.Scan()

	if !hasScanned {
		fmt.Println("\n  🛡️  Preserved all items safely in temp-remove.")

		return
	}

	input := strings.TrimSpace(scanner.Text())
	isConfirmed := strings.EqualFold(input, sampleTitle) || strings.EqualFold(input, "CONFIRM")

	if isConfirmed {
		purgedCount := purgeQuarantineTasks(database, tasks)
		fmt.Printf("\n  🗑️  Permanently removed %d items from temp-remove.\n\n", purgedCount)
	} else {
		fmt.Printf("\n  🛡️  Preserved %d items safely in temp-remove. No files were permanently deleted.\n", len(tasks))
		fmt.Println("     You can review or restore them anytime using: movie undo")
		fmt.Println()
	}
}

// purgeQuarantineTasks permanently removes items from temp-remove and updates task records.
func purgeQuarantineTasks(database *db.DB, tasks []db.TaskRecord) int {
	var count int

	for i := range tasks {
		t := &tasks[i]

		if len(t.QuarantinePath) > 0 {
			if _, statErr := os.Stat(t.QuarantinePath); statErr == nil {
				_ = os.RemoveAll(t.QuarantinePath)
			}
		}

		_ = database.UpdateTaskStatus(t.TaskId, db.TaskPurged, "")
		count++
	}

	return count
}

// restoreQuarantinedTask reverses quarantine and moves files back to their original path.
func restoreQuarantinedTask(database *db.DB, task *db.TaskRecord) error {
	if task == nil {
		return fmt.Errorf("task record cannot be nil")
	}

	if len(task.QuarantinePath) == 0 || len(task.SourcePath) == 0 {
		return fmt.Errorf("task #%d is missing path details for undo", task.TaskId)
	}

	quarantineItemPath := task.TargetPath

	if len(quarantineItemPath) == 0 {
		quarantineItemPath = filepath.Join(task.QuarantinePath, filepath.Base(task.SourcePath))
	}

	if _, statErr := os.Stat(quarantineItemPath); statErr == nil {
		if moveErr := moveFileOrDir(quarantineItemPath, task.SourcePath); moveErr != nil {
			return appfault.Wrapf(moveErr, "restore %s -> %s", quarantineItemPath, task.SourcePath)
		}
	}

	// Clean up empty holding folder
	_ = os.Remove(task.QuarantinePath)

	if task.MediaId.Valid {
		if restoreErr := database.RestoreMedia(task.MediaId.Int64); restoreErr != nil {
			errlog.Warn("Could not un-soft-delete media #%d: %v", task.MediaId.Int64, restoreErr)
		}
	}

	_ = database.UpdateTaskStatus(task.TaskId, db.TaskUndone, "")

	return nil
}
