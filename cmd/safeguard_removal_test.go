// safeguard_removal_test.go — Unit tests for root folder safeguards and removal target resolution.
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func TestValidateRemovalTargetRootsAndEmpty(t *testing.T) {
	d, err := db.OpenAtPathForTest(t.TempDir())

	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	defer d.Close()

	emptyErr := validateRemovalTarget("", d)

	if emptyErr == nil {
		t.Errorf("expected error for empty target path, got nil")
	}

	rootErr := validateRemovalTarget("/", d)

	if rootErr == nil {
		t.Errorf("expected error for root directory '/', got nil")
	}
}

func TestValidateRemovalTargetScanRootAndAncestors(t *testing.T) {
	tempBase := t.TempDir()
	d, err := db.OpenAtPathForTest(tempBase)

	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	defer d.Close()

	scanRoot := filepath.Join(tempBase, "library_root")
	_ = os.MkdirAll(scanRoot, 0o755)

	_, _ = d.UpsertScanFolder(scanRoot)

	// Registered scan root itself must be protected
	rootErr := validateRemovalTarget(scanRoot, d)

	if rootErr == nil {
		t.Errorf("expected safeguard error for registered scan root, got nil")
	}

	// Ancestor of registered scan root must be protected
	ancErr := validateRemovalTarget(tempBase, d)

	if ancErr == nil {
		t.Errorf("expected safeguard error for ancestor of scan root, got nil")
	}

	// Subfolder inside scan root with 0 or 1 movies is safe
	subDir := filepath.Join(scanRoot, "SubMovieFolder")
	_ = os.MkdirAll(subDir, 0o755)

	subErr := validateRemovalTarget(subDir, d)

	if subErr != nil {
		t.Errorf("expected no error for empty subfolder inside scan root, got: %v", subErr)
	}
}

func TestValidateRemovalTargetMultiMovieFolder(t *testing.T) {
	tempBase := t.TempDir()
	d, err := db.OpenAtPathForTest(tempBase)

	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	defer d.Close()

	sharedDir := filepath.Join(tempBase, "shared_movies")
	_ = os.MkdirAll(sharedDir, 0o755)

	fileA := filepath.Join(sharedDir, "MovieA.mkv")
	fileB := filepath.Join(sharedDir, "MovieB.mkv")
	_ = os.WriteFile(fileA, []byte("dataA"), 0o644)
	_ = os.WriteFile(fileB, []byte("dataB"), 0o644)

	_, _ = d.InsertMedia(&db.Media{
		Title:           "Movie A",
		CurrentFilePath: fileA,
		Type:            "movie",
	})
	_, _ = d.InsertMedia(&db.Media{
		Title:           "Movie B",
		CurrentFilePath: fileB,
		Type:            "movie",
	})

	// Shared folder containing 2 active movies must be protected from folder-level removal
	sharedErr := validateRemovalTarget(sharedDir, d)

	if sharedErr == nil {
		t.Errorf("expected safeguard error for shared folder with multiple movies, got nil")
	}

	// Removing single file fileA must be allowed
	fileErr := validateRemovalTarget(fileA, d)

	if fileErr != nil {
		t.Errorf("expected no safeguard error for single file removal, got: %v", fileErr)
	}
}

func TestResolveMediaRemovalTargetScanRootFile(t *testing.T) {
	tempBase := t.TempDir()
	d, err := db.OpenAtPathForTest(tempBase)

	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	defer d.Close()

	scanRoot := filepath.Join(tempBase, "movies")
	_ = os.MkdirAll(scanRoot, 0o755)
	_, _ = d.UpsertScanFolder(scanRoot)

	// Movie directly in scan root (like # Movies/Agent.Jita.mkv)
	movieFile := filepath.Join(scanRoot, "Agent.Jita.2023.mkv")
	_ = os.WriteFile(movieFile, []byte("video"), 0o644)

	sidecarNfo := filepath.Join(scanRoot, "Agent.Jita.2023.nfo")
	_ = os.WriteFile(sidecarNfo, []byte("nfo"), 0o644)

	mediaID, insErr := d.InsertMedia(&db.Media{
		Title:           "Agent Jita",
		CurrentFilePath: movieFile,
		Type:            "movie",
	})

	if insErr != nil {
		t.Fatalf("insert media: %v", insErr)
	}

	media, getErr := d.GetMediaByID(mediaID)

	if getErr != nil {
		t.Fatalf("get media: %v", getErr)
	}

	info, resErr := resolveMediaRemovalTarget(media, d)

	if resErr != nil {
		t.Fatalf("resolveMediaRemovalTarget failed: %v", resErr)
	}

	if info.IsDedicatedFolder {
		t.Errorf("expected IsDedicatedFolder=false for file directly in scan root")
	}

	if info.TargetPath != movieFile {
		t.Errorf("expected TargetPath=%s, got %s", movieFile, info.TargetPath)
	}

	hasNfo := false

	for _, sc := range info.SidecarPaths {
		if filepath.Base(sc) == "Agent.Jita.2023.nfo" {
			hasNfo = true
		}
	}

	if !hasNfo {
		t.Errorf("expected sidecar nfo in SidecarPaths, got %v", info.SidecarPaths)
	}
}

func TestResolveMediaRemovalTargetDedicatedFolder(t *testing.T) {
	tempBase := t.TempDir()
	d, err := db.OpenAtPathForTest(tempBase)

	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	defer d.Close()

	scanRoot := filepath.Join(tempBase, "movies")
	dedicatedDir := filepath.Join(scanRoot, "Inception (2010)")
	_ = os.MkdirAll(dedicatedDir, 0o755)
	_, _ = d.UpsertScanFolder(scanRoot)

	movieFile := filepath.Join(dedicatedDir, "Inception.mkv")
	_ = os.WriteFile(movieFile, []byte("video"), 0o644)

	mediaID, insErr := d.InsertMedia(&db.Media{
		Title:           "Inception",
		CurrentFilePath: movieFile,
		Type:            "movie",
	})

	if insErr != nil {
		t.Fatalf("insert media: %v", insErr)
	}

	media, getErr := d.GetMediaByID(mediaID)

	if getErr != nil {
		t.Fatalf("get media: %v", getErr)
	}

	info, resErr := resolveMediaRemovalTarget(media, d)

	if resErr != nil {
		t.Fatalf("resolveMediaRemovalTarget failed: %v", resErr)
	}

	if !info.IsDedicatedFolder {
		t.Errorf("expected IsDedicatedFolder=true for dedicated folder")
	}

	if info.TargetPath != dedicatedDir {
		t.Errorf("expected TargetPath=%s, got %s", dedicatedDir, info.TargetPath)
	}
}
