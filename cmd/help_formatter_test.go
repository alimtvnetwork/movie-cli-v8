package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestColorText(t *testing.T) {
	colored := colorText("hello", ansiCyan, true)
	if !strings.HasPrefix(colored, ansiCyan) {
		t.Fatalf("expected color prefix, got: %s", colored)
	}

	if !strings.HasSuffix(colored, ansiReset) {
		t.Fatalf("expected reset suffix, got: %s", colored)
	}

	plain := colorText("hello", ansiCyan, false)
	if plain != "hello" {
		t.Fatalf("expected plain text hello, got: %s", plain)
	}
}

func TestIsColorEnabled_NoColorEnv(t *testing.T) {
	oldVal := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", oldVal)

	os.Setenv("NO_COLOR", "1")
	if isColorEnabled() {
		t.Fatal("expected isColorEnabled to be false when NO_COLOR is set")
	}
}

func TestIsColorEnabled_DumbTerminal(t *testing.T) {
	oldTerm := os.Getenv("TERM")
	defer os.Setenv("TERM", oldTerm)

	os.Setenv("TERM", "dumb")
	if isColorEnabled() {
		t.Fatal("expected isColorEnabled to be false when TERM=dumb")
	}
}
