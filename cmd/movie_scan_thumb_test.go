package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func TestExtractThumbFileName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{".", ""},
		{"thumbnails/sample.jpg", "sample.jpg"},
		{`thumbnails\sample.jpg`, "sample.jpg"},
		{`C:\app\thumbnails\sample.jpg`, "sample.jpg"},
		{"/var/data/thumbnails/sample.jpg", "sample.jpg"},
		{"sample.jpg", "sample.jpg"},
	}

	for _, tc := range tests {
		got := extractThumbFileName(tc.input)
		if got != tc.want {
			t.Errorf("extractThumbFileName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestEnsureThumbnailInOutputDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scan_thumb_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbDir := filepath.Join(tmpDir, "db")
	dbThumbDir := filepath.Join(dbDir, "thumbnails")
	if mkErr := os.MkdirAll(dbThumbDir, 0755); mkErr != nil {
		t.Fatal(mkErr)
	}

	testContent := []byte("image binary payload")
	srcFile := filepath.Join(dbThumbDir, "inception-2010-27205.jpg")
	if wErr := os.WriteFile(srcFile, testContent, 0644); wErr != nil {
		t.Fatal(wErr)
	}

	outDir := filepath.Join(tmpDir, "output")
	ensureThumbnailInOutputDir(outDir, dbDir, "thumbnails/inception-2010-27205.jpg")

	destFile := filepath.Join(outDir, "thumbnails", "inception-2010-27205.jpg")
	data, rErr := os.ReadFile(destFile)
	if rErr != nil {
		t.Fatalf("expected thumbnail copied to output dir, got err: %v", rErr)
	}

	if string(data) != string(testContent) {
		t.Errorf("copied content mismatch: got %q, want %q", string(data), string(testContent))
	}
}

func TestTryReuseExistingThumbnail(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "reuse_thumb_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbDir := filepath.Join(tmpDir, "db")
	dbThumbDir := filepath.Join(dbDir, "thumbnails")
	_ = os.MkdirAll(dbThumbDir, 0755)

	outDir := filepath.Join(tmpDir, "output")
	outThumbDir := filepath.Join(outDir, "thumbnails")
	_ = os.MkdirAll(outThumbDir, 0755)

	fileName := "matrix-1999-603.jpg"
	srcFile := filepath.Join(dbThumbDir, fileName)
	_ = os.WriteFile(srcFile, []byte("matrix poster"), 0644)

	m := &db.Media{CleanTitle: "The Matrix", Year: 1999, TmdbID: 603}
	input := ThumbnailInput{
		Database:  &db.DB{BasePath: dbDir},
		Media:     m,
		OutputDir: outDir,
	}

	destPath := filepath.Join(outThumbDir, fileName)
	reused := tryReuseExistingThumbnail(input, destPath, fileName)
	if !reused {
		t.Errorf("expected tryReuseExistingThumbnail to return true")
	}

	if m.ThumbnailPath != "thumbnails/"+fileName {
		t.Errorf("expected Media.ThumbnailPath to be 'thumbnails/%s', got %q", fileName, m.ThumbnailPath)
	}

	// Missing thumbnail test
	missingInput := ThumbnailInput{
		Database:  &db.DB{BasePath: dbDir},
		Media:     &db.Media{},
		OutputDir: outDir,
	}
	missingReused := tryReuseExistingThumbnail(missingInput, filepath.Join(outThumbDir, "none.jpg"), "none.jpg")
	if missingReused {
		t.Errorf("expected missing thumbnail to return false")
	}
}
