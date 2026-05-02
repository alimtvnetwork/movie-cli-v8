// undo_redo_e2e_test.go — end-to-end CLI integration tests for `movie undo`
// and `movie redo`.
//
// What this verifies:
//
//  1. A seeded MoveHistory row + on-disk file at ToPath gets reverted by
//     `movie undo --yes --global`: the file moves back to FromPath, the
//     MoveHistory row flips IsReverted=1, and Media.CurrentFilePath is
//     updated to FromPath.
//
//  2. `movie redo --yes --global` re-applies that same move: file lands
//     at ToPath again, IsReverted flips back to 0, and CurrentFilePath
//     follows the file.
//
//  3. Action-history undo by ID: a seeded Delete action with a media
//     snapshot is reverted by `movie undo --id <action_id> --yes`. The
//     correct ActionHistoryId is marked IsReverted=1 and a Restore row
//     is appended (matching executeActionUndo → undoDelete behavior).
//
//  4. Action-history redo by ID: `movie redo --id <action_id> --yes`
//     flips the same row back and removes the restored media (matching
//     executeActionRedo → redoDelete behavior).
//
// Both tests share one CLI binary built into a temp dir so the binary's
// sibling `data/movie.db` is the same file the test seeds via
// db.OpenAtPathForTest. Set E2E=1 to enable (matches the existing
// pipeline test gating).
package cmd

import (
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// undoRedoE2EFlag mirrors the existing pipeline test's gate so CI can opt
// into the heavier integration suite without slowing down `go test ./...`.
const undoRedoE2EFlag = "E2E"

func TestE2EUndoRedoMove(t *testing.T) {
	if os.Getenv(undoRedoE2EFlag) != "1" {
		t.Skipf("skipping undo/redo E2E (set %s=1 to enable)", undoRedoE2EFlag)
	}

	workDir := t.TempDir()
	binary := buildE2EBinary(t, workDir)

	database := openSharedDB(t, workDir)
	defer database.Close()

	mediaID := seedTestMedia(t, database, "Inception", 27205)
	fromPath, toPath := seedMovedFile(t, workDir, mediaID, database)

	// --- UNDO ---
	runUndoRedoCLI(t, binary, workDir, "undo", "--yes", "--global")
	assertFileExists(t, fromPath)
	assertFileMissing(t, toPath)
	assertMoveReverted(t, database, true)
	assertMediaPath(t, database, mediaID, fromPath)

	// --- REDO ---
	runUndoRedoCLI(t, binary, workDir, "redo", "--yes", "--global")
	assertFileExists(t, toPath)
	assertFileMissing(t, fromPath)
	assertMoveReverted(t, database, false)
	assertMediaPath(t, database, mediaID, toPath)
}

func TestE2EUndoRedoActionByID(t *testing.T) {
	if os.Getenv(undoRedoE2EFlag) != "1" {
		t.Skipf("skipping undo/redo E2E (set %s=1 to enable)", undoRedoE2EFlag)
	}

	workDir := t.TempDir()
	binary := buildE2EBinary(t, workDir)

	database := openSharedDB(t, workDir)
	defer database.Close()

	mediaID := seedTestMedia(t, database, "Arrival", 329865)
	actionID := seedDeleteAction(t, database, mediaID)

	// --- UNDO by ID: should mark this exact action_id reverted and
	// re-insert the snapshot as a Restore row. ---
	runUndoRedoCLI(t, binary, workDir, "undo", "--id", itoa(actionID), "--yes")
	assertActionReverted(t, database, actionID, true)
	if got := countActionsByType(t, database, db.FileActionRestore); got != 1 {
		t.Fatalf("expected 1 Restore action after undo, got %d", got)
	}

	// --- REDO by ID: flips the same action_id back to not-reverted. ---
	runUndoRedoCLI(t, binary, workDir, "redo", "--id", itoa(actionID), "--yes")
	assertActionReverted(t, database, actionID, false)
}

// --- helpers ---------------------------------------------------------------

// openSharedDB opens the on-disk database that the CLI binary will also
// use (binary's sibling data/movie.db). Using OpenAtPathForTest guarantees
// schema parity with production Open().
func openSharedDB(t *testing.T, workDir string) *db.DB {
	t.Helper()
	d, err := db.OpenAtPathForTest(workDir)
	if err != nil {
		t.Fatalf("open shared DB: %v", err)
	}
	return d
}

func seedTestMedia(t *testing.T, d *db.DB, title string, tmdbID int) int64 {
	t.Helper()
	id, err := d.InsertMedia(&db.Media{
		Title:            title,
		CleanTitle:       title,
		Year:             2020,
		Type:             string(db.MediaTypeMovie),
		TmdbID:           tmdbID,
		OriginalFileName: title + ".mkv",
		OriginalFilePath: "/seed/" + title + ".mkv",
		CurrentFilePath:  "/seed/" + title + ".mkv",
		FileExtension:    ".mkv",
		FileSizeMb:       1024.0,
	})
	if err != nil {
		t.Fatalf("seed media %q: %v", title, err)
	}
	return id
}

// seedMovedFile creates a real file at toPath, registers a MoveHistory
// row (FromPath → ToPath), and updates Media.CurrentFilePath to ToPath
// — i.e. the exact state the DB would be in right after a successful
// `movie move`. Returns the (fromPath, toPath) pair for assertions.
func seedMovedFile(t *testing.T, workDir string, mediaID int64, d *db.DB) (string, string) {
	t.Helper()
	fromDir := filepath.Join(workDir, "src")
	toDir := filepath.Join(workDir, "dst")
	if err := os.MkdirAll(fromDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.MkdirAll(toDir, 0o755); err != nil {
		t.Fatalf("mkdir dst: %v", err)
	}
	fromPath := filepath.Join(fromDir, "Inception.mkv")
	toPath := filepath.Join(toDir, "Inception.mkv")
	if err := os.WriteFile(toPath, []byte("fake-video"), 0o644); err != nil {
		t.Fatalf("write moved file: %v", err)
	}
	if err := d.InsertMoveHistory(db.MoveInput{
		FromPath: fromPath, ToPath: toPath,
		OrigName: "Inception.mkv", NewName: "Inception.mkv",
		MediaID: mediaID, FileActionID: int(db.FileActionMove),
	}); err != nil {
		t.Fatalf("insert move history: %v", err)
	}
	if err := d.UpdateMediaPath(mediaID, toPath); err != nil {
		t.Fatalf("update media path: %v", err)
	}
	return fromPath, toPath
}

// seedDeleteAction inserts a Delete ActionHistory row carrying a full
// media snapshot, then deletes the media (mirroring what `movie delete`
// would persist). Returns the new action_id so the CLI can target it.
func seedDeleteAction(t *testing.T, d *db.DB, mediaID int64) int64 {
	t.Helper()
	media, err := d.GetMediaByID(mediaID)
	if err != nil {
		t.Fatalf("load media for snapshot: %v", err)
	}
	snap, snapErr := db.MediaToJSON(media)
	if snapErr != nil {
		t.Fatalf("snapshot media: %v", snapErr)
	}
	actionID, insertErr := d.InsertActionSimple(db.ActionSimpleInput{
		FileAction: db.FileActionDelete,
		MediaID:    mediaID,
		Snapshot:   snap,
		Detail:     "Deleted: " + media.Title,
	})
	if insertErr != nil {
		t.Fatalf("insert delete action: %v", insertErr)
	}
	if delErr := d.DeleteMediaByID(mediaID); delErr != nil {
		t.Fatalf("delete media: %v", delErr)
	}
	return actionID
}

// runUndoRedoCLI runs the built binary with HOME pointed inside the
// workDir and TMDb keys cleared. Fails the test on non-zero exit so we
// catch CLI regressions immediately.
func runUndoRedoCLI(t *testing.T, binary, workDir string, args ...string) string {
	t.Helper()
	homeDir := filepath.Join(workDir, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"USERPROFILE="+homeDir,
		"TMDB_API_KEY=",
		"TMDB_TOKEN=",
		"OMDB_API_KEY=",
	)
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		t.Fatalf("movie %s failed: %v\n%s", strings.Join(args, " "), runErr, out)
	}
	return string(out)
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file at %s, got: %v", path, err)
	}
}

func assertFileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected no file at %s, stat err=%v", path, err)
	}
}

func assertMoveReverted(t *testing.T, d *db.DB, want bool) {
	t.Helper()
	moves, err := d.ListMoveHistory(10)
	if err != nil || len(moves) == 0 {
		t.Fatalf("list move history: err=%v len=%d", err, len(moves))
	}
	if moves[0].IsReverted != want {
		t.Fatalf("MoveHistory.IsReverted = %v, want %v", moves[0].IsReverted, want)
	}
}

func assertMediaPath(t *testing.T, d *db.DB, id int64, want string) {
	t.Helper()
	m, err := d.GetMediaByID(id)
	if err != nil {
		t.Fatalf("get media %d: %v", id, err)
	}
	if m.CurrentFilePath != want {
		t.Fatalf("CurrentFilePath = %q, want %q", m.CurrentFilePath, want)
	}
}

func assertActionReverted(t *testing.T, d *db.DB, actionID int64, want bool) {
	t.Helper()
	a, err := d.GetActionByID(actionID)
	if err != nil {
		t.Fatalf("get action %d: %v", actionID, err)
	}
	if a.IsReverted != want {
		t.Fatalf("action %d IsReverted = %v, want %v", actionID, a.IsReverted, want)
	}
}

// countActionsByType counts ActionHistory rows of a given FileActionId,
// used to verify the Restore-side-effect of executeActionUndo.
func countActionsByType(t *testing.T, d *db.DB, kind db.FileActionType) int {
	t.Helper()
	var n int
	row := d.QueryRow("SELECT COUNT(*) FROM ActionHistory WHERE FileActionId = ?", int(kind))
	if err := row.Scan(&n); err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		t.Fatalf("count actions: %v", err)
	}
	return n
}

// itoa keeps the call-site terse without pulling in strconv at the top
// (kept local so the test file is self-contained for reviewers).
func itoa(v int64) string {
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	buf := make([]byte, 0, 20)
	for v > 0 {
		buf = append([]byte{digits[v%10]}, buf...)
		v /= 10
	}
	if neg {
		return "-" + string(buf)
	}
	return string(buf)
}
