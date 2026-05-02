package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidRepoRequiresV5Module(t *testing.T) {
	repoDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	goMod := []byte("module github.com/alimtvnetwork/movie-cli-v8\n\ngo 1.22\n")
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), goMod, 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if !isValidRepo(repoDir) {
		t.Fatal("expected v5 repo to be valid")
	}
}

func TestIsValidRepoRejectsOldModulePath(t *testing.T) {
	repoDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	// Use a clearly-fake legacy path so the CI old-module-path guard
	// (which forbids any v2/v3/v4 reference) does not flag this test.
	goMod := []byte("module github.com/alimtvnetwork/movie-cli-legacy\n\ngo 1.22\n")
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), goMod, 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if isValidRepo(repoDir) {
		t.Fatal("expected non-v5 module path to be rejected")
	}
}

func TestNormalizeRepoPathTrimsQuotes(t *testing.T) {
	repoDir := t.TempDir()
	quoted := "  \"" + repoDir + "\"  "
	got, err := normalizeRepoPath(quoted)
	if err != nil {
		t.Fatalf("normalize repo path: %v", err)
	}
	want := filepath.Clean(repoDir)
	if got != want {
		t.Fatalf("normalize repo path = %q, want %q", got, want)
	}
}
