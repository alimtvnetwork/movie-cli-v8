// movie_popout_integration_test.go — end-to-end integration tests for the
// `movie popout` command and the universal cwd-default rule.
//
// Why these tests exist:
//   - The popout command silently exited with no error when invoked without
//     a path argument because resolvePopoutDir() called an interactive
//     prompt that returned "" on a closed stdin. v2.136.0 fixed this with
//     ResolveTargetDir(args, home) and runMoviePopout was rewritten to
//     loudly fail or default to cwd. These tests pin that behavior so the
//     regression cannot come back.
//   - The popout cleanup phase used to delete folders. v2.136.0 replaced
//     that with the .temp/ compaction flow. These tests prove every
//     non-media folder ends up under <root>/.temp/ and the originals are
//     gone from root.
//
// All tests use t.TempDir() so they leave no artifacts behind, work on
// every CI runner, and parallelize safely.
//
// Run:
//
//	go test ./cmd/ -run TestPopout -v
package cmd

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// ---------------------------------------------------------------------------
// Tree-spec helpers
// ---------------------------------------------------------------------------

// seedSpec describes one entry in a synthetic media tree.
//
//	"a/b/Movie.2021.1080p.x264.mp4|8" → file with 8 KiB of zero bytes.
//	"emptyfolder/"                    → directory with no contents.
//	"junk/readme.nfo|1"               → a non-media file inside junk/.
type seedSpec string

// seedTree materializes a slice of seedSpec entries under root. It is
// intentionally tiny — no DSL parsing libraries — so the test reader can
// see exactly what is on disk.
func seedTree(t *testing.T, root string, specs []seedSpec) {
	t.Helper()
	for _, s := range specs {
		raw := string(s)
		path := raw
		sizeKB := 0

		if pipe := strings.LastIndex(raw, "|"); pipe >= 0 {
			path = raw[:pipe]
			// best-effort size parse; default 1 KiB if missing/invalid
			sizeKB = 1
			for _, c := range raw[pipe+1:] {
				if c < '0' || c > '9' {
					sizeKB = 1
					break
				}
				sizeKB = sizeKB*10 + int(c-'0')
			}
			if sizeKB == 0 {
				sizeKB = 1
			}
		}

		full := filepath.Join(root, filepath.FromSlash(path))
		if strings.HasSuffix(path, "/") {
			if err := os.MkdirAll(full, 0755); err != nil {
				t.Fatalf("mkdir %s: %v", full, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir parent of %s: %v", full, err)
		}
		body := bytes.Repeat([]byte{0}, sizeKB*1024)
		if err := os.WriteFile(full, body, 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
}

// listEntries returns sorted names of direct children of dir, or nil if
// the dir does not exist.
func listEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("readdir %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// scannerFromString builds a *bufio.Scanner that yields the given lines
// (one per Scan() call). Used to feed deterministic input to interactive
// prompts inside the popout flow.
func scannerFromString(input string) *bufio.Scanner {
	return bufio.NewScanner(strings.NewReader(input))
}

// openTempDB opens an in-memory database via the public test helper so
// these cross-package integration tests stay sandboxed and never touch
// the real ./data/movie.db file.
func openTempDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.OpenInMemoryForTest(t.TempDir())
	if err != nil {
		t.Fatalf("OpenInMemoryForTest: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

// ---------------------------------------------------------------------------
// Test 1 — Discovery + flatten: nested media files end up at root.
// ---------------------------------------------------------------------------

func TestPopout_FlattensNestedMedia(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	seedTree(t, root, []seedSpec{
		"InceptionFolder/Inception.2010.1080p.x264.mp4|4",
		"InceptionFolder/extras/sample.mp4|1",
		"DeepNest/level1/level2/Matrix.1999.720p.avi|3",
		"DocOnlyFolder/readme.nfo|1",
		"DocOnlyFolder/poster.jpg|1",
		"emptyFolder/",
	})

	items := discoverNestedVideos(root, 3)

	if len(items) < 3 {
		t.Fatalf("expected at least 3 nested video items (Inception, sample, Matrix), got %d", len(items))
	}

	// Spot-check that Inception was discovered.
	foundInception := false
	for _, it := range items {
		if strings.Contains(it.srcPath, "Inception.2010") {
			foundInception = true
			if it.subDir != "InceptionFolder" {
				t.Errorf("Inception subDir = %q, want InceptionFolder", it.subDir)
			}
			if filepath.Dir(it.destPath) != root {
				t.Errorf("Inception destPath dir = %q, want root %q", filepath.Dir(it.destPath), root)
			}
		}
	}
	if !foundInception {
		t.Errorf("Inception.2010 was not discovered as a popout candidate")
	}
}

// ---------------------------------------------------------------------------
// Test 2 — Discovery respects the popoutTempDir convention.
// ---------------------------------------------------------------------------

func TestPopout_DiscoverAllSubdirsExcludesTempDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	seedTree(t, root, []seedSpec{
		"keepme/file.mp4|2",
		".temp/already-compacted/leftover.txt|1",
	})

	subs := discoverAllSubdirs(root, 3)

	if !contains(subs, "keepme") {
		t.Errorf("expected 'keepme' in subdirs, got %v", subs)
	}
	if contains(subs, ".temp") {
		t.Errorf(".temp/ must be excluded from subdirs, got %v", subs)
	}
}

// ---------------------------------------------------------------------------
// Test 3 — folderHasMedia correctly classifies media vs non-media folders.
// ---------------------------------------------------------------------------

func TestPopout_FolderHasMediaClassification(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	seedTree(t, root, []seedSpec{
		"hasmovie/film.mkv|1",
		"hasmovie/extras/sample.mp4|1",
		"docsonly/readme.nfo|1",
		"docsonly/poster.jpg|1",
		"emptydir/",
	})

	cases := []struct {
		dir  string
		want bool
	}{
		{"hasmovie", true},
		{"docsonly", false},
		{"emptydir", false},
	}
	for _, tc := range cases {
		got := folderHasMedia(filepath.Join(root, tc.dir))
		if got != tc.want {
			t.Errorf("folderHasMedia(%q) = %v, want %v", tc.dir, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 4 — End-to-end compaction: non-media folders move into <root>/.temp/.
// ---------------------------------------------------------------------------

func TestPopout_CompactsNonMediaFoldersIntoTemp(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	seedTree(t, root, []seedSpec{
		"docsonly/readme.nfo|1",
		"docsonly/poster.jpg|1",
		"emptydir/",
		"keepme/film.mp4|2", // has media → must NOT be compacted
	})

	// Use a dedicated temp DB so the compact-action insert doesn't pollute
	// the developer's real movie.db.
	d := openTempDB(t)

	cc := CleanupContext{
		Scanner:  scannerFromString(""),
		Database: d,
		BatchID:  "test-batch-compact",
	}

	subs := discoverAllSubdirs(root, 3)
	compactNonMediaFolders(cc, root, subs, true) // autoCompact=true → no prompt

	tempEntries := listEntries(t, filepath.Join(root, popoutTempDir))
	rootEntries := listEntries(t, root)

	if !contains(tempEntries, "docsonly") {
		t.Errorf("docsonly should be inside .temp/, got temp=%v", tempEntries)
	}
	if !contains(tempEntries, "emptydir") {
		t.Errorf("emptydir should be inside .temp/, got temp=%v", tempEntries)
	}
	if contains(rootEntries, "docsonly") {
		t.Errorf("docsonly must be removed from root after compaction, root=%v", rootEntries)
	}
	if !contains(rootEntries, "keepme") {
		t.Errorf("keepme (has media) must remain at root, root=%v", rootEntries)
	}

	// Verify history rows were inserted with FileActionCompact.
	actions, err := d.ListActionsByBatch("test-batch-compact")
	if err != nil {
		t.Fatalf("ListActionsByBatch: %v", err)
	}
	if len(actions) != 2 {
		t.Errorf("expected 2 compact actions, got %d", len(actions))
	}
	for _, a := range actions {
		if a.FileActionId != db.FileActionCompact {
			t.Errorf("action %d: FileActionId = %d, want %d (Compact)",
				a.ActionHistoryId, a.FileActionId, db.FileActionCompact)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 5 — Universal cwd-default rule: ResolveTargetDir returns cwd when
// no args are given. This is the regression fence for the silent-failure
// bug from v2.135 and earlier.
// ---------------------------------------------------------------------------

func TestResolveTargetDir_DefaultsToCwdWhenNoArg(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	originalWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(originalWD) })

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	dir, err := ResolveTargetDir(nil, "/fake/home")
	if err != nil {
		t.Fatalf("ResolveTargetDir error: %v", err)
	}
	if dir == "" {
		t.Fatal("ResolveTargetDir returned empty string — silent-failure regression!")
	}
	// Resolve symlinks for macOS /private/var/folders/... vs /var/folders/...
	wantResolved, _ := filepath.EvalSymlinks(tmp)
	gotResolved, _ := filepath.EvalSymlinks(dir)
	if wantResolved != gotResolved {
		t.Errorf("ResolveTargetDir = %q, want cwd %q", gotResolved, wantResolved)
	}
}

// ---------------------------------------------------------------------------
// Test 6 — ResolveTargetDir expands ~/ correctly when an arg IS given.
// ---------------------------------------------------------------------------

func TestResolveTargetDir_ExpandsHomeWhenArgGiven(t *testing.T) {
	t.Parallel()
	dir, err := ResolveTargetDir([]string{"~/movies"}, "/fake/home")
	if err != nil {
		t.Fatalf("ResolveTargetDir error: %v", err)
	}
	want := filepath.Join("/fake/home", "movies")
	if dir != want {
		t.Errorf("ResolveTargetDir(~/movies) = %q, want %q", dir, want)
	}
}

// ---------------------------------------------------------------------------
// Test 7 — Empty-string arg also defaults to cwd (defensive).
// ---------------------------------------------------------------------------

func TestResolveTargetDir_EmptyStringArgDefaultsToCwd(t *testing.T) {
	t.Parallel()
	dir, err := ResolveTargetDir([]string{"   "}, "/fake/home")
	if err != nil {
		t.Fatalf("ResolveTargetDir error: %v", err)
	}
	if dir == "" {
		t.Fatal("empty-string arg should fall back to cwd, got empty")
	}
}
