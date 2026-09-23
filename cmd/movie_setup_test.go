// movie_setup_test.go — Unit tests for shell wrapper reconciliation and rendering.
package cmd

import (
	"strings"
	"testing"
)

func TestReconcileWrapperContentNew(t *testing.T) {
	initial := "# User custom profile\nWrite-Host 'Hello'\n"
	snippet := renderPowerShellSnippet()

	result := reconcileWrapperContent(initial, snippet)

	if !strings.Contains(result, wrapperMarkerStart) {
		t.Fatalf("expected result to contain marker start")
	}

	if !strings.Contains(result, wrapperMarkerEnd) {
		t.Fatalf("expected result to contain marker end")
	}

	if !strings.HasPrefix(result, initial) {
		t.Fatalf("expected result to preserve initial profile content")
	}
}

func TestReconcileWrapperContentReplace(t *testing.T) {
	oldSnippet := "# movie-cli command wrapper v1\nfunction old { }\n# movie-cli command wrapper v1 end"
	initial := "# Before\n" + oldSnippet + "\n# After\n"
	newSnippet := renderPowerShellSnippet()

	result := reconcileWrapperContent(initial, newSnippet)

	if strings.Contains(result, "function old") {
		t.Fatalf("expected old wrapper function to be replaced")
	}

	if !strings.Contains(result, "function mcd") {
		t.Fatalf("expected new wrapper function mcd to be present")
	}

	if !strings.Contains(result, "# After") {
		t.Fatalf("expected content after wrapper to be preserved")
	}
}

func TestRenderWrapperSnippet(t *testing.T) {
	pwsh := renderWrapperSnippet("powershell")

	if !strings.Contains(pwsh, "function movie") {
		t.Fatalf("expected powershell snippet to define 'function movie'")
	}

	if !strings.Contains(pwsh, "function mcd") {
		t.Fatalf("expected powershell snippet to define 'function mcd'")
	}

	if !strings.Contains(pwsh, "Set-Location -LiteralPath") {
		t.Fatalf("expected powershell snippet to use 'Set-Location -LiteralPath'")
	}

	unix := renderWrapperSnippet("bash")

	if !strings.Contains(unix, "movie()") {
		t.Fatalf("expected bash snippet to define 'movie()'")
	}

	if !strings.Contains(unix, "mcd()") {
		t.Fatalf("expected bash snippet to define 'mcd()'")
	}
}

func TestDeduplicatePaths(t *testing.T) {
	input := []string{
		`C:\Users\Admin\Documents\PowerShell\profile.ps1`,
		`c:\users\admin\documents\powershell\profile.ps1`,
		`C:\Users\Admin\Documents\WindowsPowerShell\profile.ps1`,
		"",
	}

	deduped := deduplicatePaths(input)

	if len(deduped) != 2 {
		t.Fatalf("expected 2 unique paths, got %d: %v", len(deduped), deduped)
	}
}

func TestDetectTargetShell(t *testing.T) {
	tests := []struct {
		override string
		expected string
	}{
		{"pwsh", "powershell"},
		{"powershell", "powershell"},
		{"bash", "bash"},
		{"zsh", "zsh"},
	}

	for _, tc := range tests {
		actual := detectTargetShell(tc.override)

		if actual != tc.expected {
			t.Errorf("detectTargetShell(%q) = %q; expected %q", tc.override, actual, tc.expected)
		}
	}
}
