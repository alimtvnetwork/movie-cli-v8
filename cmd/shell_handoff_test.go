// shell_handoff_test.go — Unit tests for shell handoff writing and wrapper status.
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteHandoffPath(t *testing.T) {
	tempDir := t.TempDir()
	handoffFile := filepath.Join(tempDir, "handoff.txt")

	t.Setenv(envMovieHandoffFile, handoffFile)

	targetPath := `C:\Media\Movies\Inception`
	writeHandoffPath(targetPath)

	content, err := os.ReadFile(handoffFile)

	if err != nil {
		t.Fatalf("expected handoff file to exist: %v", err)
	}

	if string(content) != targetPath {
		t.Fatalf("expected handoff content %q, got %q", targetPath, string(content))
	}
}

func TestWriteHandoffPathUnset(t *testing.T) {
	_ = os.Unsetenv(envMovieHandoffFile)

	// Should not panic or error when env var is unset
	writeHandoffPath(`C:\Media\Movies`)
}

func TestIsWrapperActive(t *testing.T) {
	_ = os.Unsetenv(envMovieCommandWrapper)

	if isWrapperActive() {
		t.Fatalf("expected isWrapperActive to be false when unset")
	}

	t.Setenv(envMovieCommandWrapper, "1")

	if !isWrapperActive() {
		t.Fatalf("expected isWrapperActive to be true when set to 1")
	}
}
