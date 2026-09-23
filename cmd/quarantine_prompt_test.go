// quarantine_prompt_test.go — Unit tests for quarantine execution, undo restoration, and purge.
package cmd

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func TestExecuteQuarantineRemovalAndRestore(t *testing.T) {
	tempBase := t.TempDir()
	d, err := db.OpenAtPathForTest(tempBase)

	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	defer d.Close()

	scanRoot := filepath.Join(tempBase, "library")
	_ = os.MkdirAll(scanRoot, 0o755)
	_, _ = d.UpsertScanFolder(scanRoot)

	// Create a movie file and sidecar directly inside scanRoot
	videoPath := filepath.Join(scanRoot, "TestMovie.2024.mkv")
	sidecarPath := filepath.Join(scanRoot, "TestMovie.2024.nfo")
	_ = os.WriteFile(videoPath, []byte("fake video data"), 0o644)
	_ = os.WriteFile(sidecarPath, []byte("<movie>Test</movie>"), 0o644)

	mediaID, insErr := d.InsertMedia(&db.Media{
		Title:           "Test Movie",
		CurrentFilePath: videoPath,
		Type:            "movie",
	})

	if insErr != nil {
		t.Fatalf("insert media: %v", insErr)
	}

	rec := &db.StagedActionRecord{
		ActionType: db.StagedDelete,
		MediaId:    sql.NullInt64{Int64: mediaID, Valid: true},
		SourcePath: videoPath,
	}

	qErr := executeQuarantineRemoval(d, rec, "test_batch_1")

	if qErr != nil {
		t.Fatalf("executeQuarantineRemoval failed: %v", qErr)
	}

	// Verify original file and sidecar were moved from scanRoot
	if _, statErr := os.Stat(videoPath); !os.IsNotExist(statErr) {
		t.Errorf("expected original video file to be moved, but it still exists at %s", videoPath)
	}

	if _, statErr := os.Stat(sidecarPath); !os.IsNotExist(statErr) {
		t.Errorf("expected sidecar nfo to be moved, but it still exists at %s", sidecarPath)
	}

	// Verify media is soft-deleted
	var isDeleted int
	_ = d.QueryRow("SELECT IsDeleted FROM Media WHERE MediaId = ?", mediaID).Scan(&isDeleted)

	if isDeleted == 0 {
		t.Errorf("expected media to be soft-deleted (IsDeleted=1), got %d", isDeleted)
	}

	// Verify task record in DB
	task, taskErr := d.GetQuarantinedTaskByMediaID(mediaID)

	if taskErr != nil || task == nil {
		t.Fatalf("expected quarantined task in DB, got: %v", taskErr)
	}

	if task.Status != db.TaskQuarantined {
		t.Errorf("expected task status %q, got %q", db.TaskQuarantined, task.Status)
	}

	// Verify files exist in temp-remove
	if _, statErr := os.Stat(task.QuarantinePath); os.IsNotExist(statErr) {
		t.Fatalf("expected quarantine folder to exist at %s", task.QuarantinePath)
	}

	// Test undo restoration
	undoErr := restoreQuarantinedTask(d, task)

	if undoErr != nil {
		t.Fatalf("restoreQuarantinedTask failed: %v", undoErr)
	}

	// Verify file is back at original videoPath
	if _, statErr := os.Stat(videoPath); os.IsNotExist(statErr) {
		t.Errorf("expected video file to be restored to %s", videoPath)
	}

	// Verify media is active again
	var isDeletedRestored int
	_ = d.QueryRow("SELECT IsDeleted FROM Media WHERE MediaId = ?", mediaID).Scan(&isDeletedRestored)

	if isDeletedRestored != 0 {
		t.Errorf("expected restored media to have IsDeleted=0, got %d", isDeletedRestored)
	}

	// Verify task status is undone
	updatedTask, _ := d.GetTaskByID(task.TaskId)

	if updatedTask.Status != db.TaskUndone {
		t.Errorf("expected updated task status %q, got %q", db.TaskUndone, updatedTask.Status)
	}
}

func TestPurgeQuarantineTasks(t *testing.T) {
	tempBase := t.TempDir()
	d, err := db.OpenAtPathForTest(tempBase)

	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	defer d.Close()

	scanRoot := filepath.Join(tempBase, "library")
	_ = os.MkdirAll(scanRoot, 0o755)
	_, _ = d.UpsertScanFolder(scanRoot)

	videoPath := filepath.Join(scanRoot, "MovieToPurge.mkv")
	_ = os.WriteFile(videoPath, []byte("data"), 0o644)

	mediaID, _ := d.InsertMedia(&db.Media{
		Title:           "Movie To Purge",
		CurrentFilePath: videoPath,
		Type:            "movie",
	})

	rec := &db.StagedActionRecord{
		ActionType: db.StagedDelete,
		MediaId:    sql.NullInt64{Int64: mediaID, Valid: true},
		SourcePath: videoPath,
	}

	_ = executeQuarantineRemoval(d, rec, "purge_batch")

	tasks, _ := d.ListPendingQuarantineTasks()

	if len(tasks) != 1 {
		t.Fatalf("expected 1 quarantined task, got %d", len(tasks))
	}

	quarantinePath := tasks[0].QuarantinePath

	purgedCount := purgeQuarantineTasks(d, tasks)

	if purgedCount != 1 {
		t.Errorf("expected 1 purged task, got %d", purgedCount)
	}

	// Verify directory in temp-remove was deleted
	if _, statErr := os.Stat(quarantinePath); !os.IsNotExist(statErr) {
		t.Errorf("expected quarantine directory to be removed from disk, but it still exists at %s", quarantinePath)
	}

	// Verify task status updated to purged
	updatedTask, _ := d.GetTaskByID(tasks[0].TaskId)

	if updatedTask.Status != db.TaskPurged {
		t.Errorf("expected task status %q, got %q", db.TaskPurged, updatedTask.Status)
	}
}

func TestCalculateQuarantineSizeBytesAndFormatting(t *testing.T) {
	tempBase := t.TempDir()
	taskDir1 := filepath.Join(tempBase, "task1")
	taskDir2 := filepath.Join(tempBase, "task2")
	_ = os.MkdirAll(taskDir1, 0o755)
	_ = os.MkdirAll(taskDir2, 0o755)

	_ = os.WriteFile(filepath.Join(taskDir1, "movie1.mkv"), make([]byte, 1024), 0o644)
	_ = os.WriteFile(filepath.Join(taskDir2, "movie2.mkv"), make([]byte, 2048), 0o644)

	tasks := []db.TaskRecord{
		{QuarantinePath: taskDir1},
		{QuarantinePath: taskDir2},
	}

	size := calculateQuarantineSizeBytes(tasks)

	if size != 3072 {
		t.Errorf("expected 3072 bytes, got %d", size)
	}

	formatted := formatBytes(size)

	if formatted != "3.0 KB" {
		t.Errorf("expected '3.0 KB', got %q", formatted)
	}

	formattedLarge := formatBytes(35 * 1024 * 1024 * 1024)

	if formattedLarge != "35.0 GB" {
		t.Errorf("expected '35.0 GB', got %q", formattedLarge)
	}
}
