// checks_test.go — unit tests for doctor diagnostic checks and path helpers.
package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExpandScanPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("skipping test; cannot determine home directory")
	}

	tests := []struct {
		input string
		want  string
	}{
		{
			input: "~/Downloads",
			want:  filepath.Join(home, "Downloads"),
		},
		{
			input: "~",
			want:  home,
		},
		{
			input: filepath.Join("var", "movies"),
			want:  filepath.Join("var", "movies"),
		},
	}

	for _, tt := range tests {
		got := expandScanPath(tt.input)
		if got != tt.want {
			t.Errorf("expandScanPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCheckPathMismatch(t *testing.T) {
	t.Run("no target binary", func(t *testing.T) {
		r := &Report{Source: "dummy", Target: ""}
		f := checkPathMismatch(r)

		if f.Severity != SeverityErr {
			t.Errorf("expected SeverityErr, got %s", f.Severity)
		}
	})

	t.Run("matching paths", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "movie.exe")
		r := &Report{Source: p, Target: p}
		f := checkPathMismatch(r)

		if f.Severity != SeverityOK {
			t.Errorf("expected SeverityOK, got %s", f.Severity)
		}
	})

	t.Run("source does not exist on disk", func(t *testing.T) {
		nonExistent := filepath.Join(t.TempDir(), "non-existent", "movie.exe")
		targetFile := filepath.Join(t.TempDir(), "active", "movie.exe")

		r := &Report{Source: nonExistent, Target: targetFile}
		f := checkPathMismatch(r)

		if f.Severity != SeverityWarn {
			t.Errorf("expected SeverityWarn for missing source, got %s", f.Severity)
		}

		if !f.IsFixable {
			t.Error("expected IsFixable to be true")
		}

		if !strings.Contains(f.FixHint, "syncs deployPath") {
			t.Errorf("expected sync hint, got %s", f.FixHint)
		}
	})
}

func TestCheckVersionDriftMissingSource(t *testing.T) {
	nonExistent := filepath.Join(t.TempDir(), "non-existent", "movie.exe")
	r := &Report{Source: nonExistent, Target: ""}

	f := checkVersionDrift(r)
	if f.Severity != SeverityOK {
		t.Errorf("expected SeverityOK when source does not exist, got %s", f.Severity)
	}
}

func TestHasSourceBinary(t *testing.T) {
	if hasSourceBinary("") {
		t.Error("expected false for empty path")
	}

	tempFile := filepath.Join(t.TempDir(), "test.bin")
	if hasSourceBinary(tempFile) {
		t.Error("expected false before file creation")
	}

	if err := os.WriteFile(tempFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if !hasSourceBinary(tempFile) {
		t.Error("expected true after file creation")
	}
}

func TestDefaultDeployDir(t *testing.T) {
	dir := defaultDeployDir()
	if dir == "" {
		t.Error("expected non-empty defaultDeployDir")
	}

	if runtime.GOOS == "windows" {
		if !strings.Contains(dir, "movie-cli") {
			t.Errorf("expected movie-cli in default deploy dir, got %s", dir)
		}
	}
}
