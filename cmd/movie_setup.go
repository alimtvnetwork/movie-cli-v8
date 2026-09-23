// movie_setup.go — Shell wrapper setup and auto-repair for movie CLI.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	wrapperMarkerStart = "# movie-cli command wrapper v1"
	wrapperMarkerEnd   = "# movie-cli command wrapper v1 end"
)

var (
	setupDryRunFlag bool
	setupShellFlag  string
)

var movieSetupCmd = &cobra.Command{
	Use:     "setup",
	Aliases: []string{"init", "shell-setup"},
	Short:   "Install shell wrappers for instant terminal navigation and directory jumps",
	Long: `Configures shell integration so that 'movie cd' and 'mcd' physically change
your terminal shell directory, and 'movie open' / 'movie start' opens the target
folder in your operating system's file explorer.

Supported Shells:
  • PowerShell (Windows PowerShell 5.1 & PowerShell Core 7+)
  • Bash (~/.bashrc)
  • Zsh (~/.zshrc)

Examples:
  movie setup                    Detect shell and install wrapper
  movie setup --dry-run          Preview profile changes without writing
  movie setup --shell pwsh       Install specifically for PowerShell`,
	Run: runMovieSetup,
}

func init() {
	movieSetupCmd.Flags().BoolVar(&setupDryRunFlag, "dry-run", false, "display wrapper code without modifying files")
	movieSetupCmd.Flags().StringVar(&setupShellFlag, "shell", "", "force target shell (powershell, bash, zsh)")
}

func runMovieSetup(cmd *cobra.Command, args []string) {
	targetShell := detectTargetShell(setupShellFlag)
	profilePaths := resolveShellProfilePaths(targetShell)

	if len(profilePaths) == 0 {
		fmt.Fprintln(os.Stderr, "⚠️ No shell profile paths could be detected.")

		return
	}

	snippet := renderWrapperSnippet(targetShell)

	if setupDryRunFlag {
		fmt.Printf("Previewing wrapper for shell: %s\n", targetShell)
		fmt.Println("Target profile paths:")

		for _, p := range profilePaths {
			fmt.Printf("  • %s\n", p)
		}

		fmt.Println("\nWrapper snippet:")
		fmt.Println(snippet)

		return
	}

	var updatedCount int

	for _, profilePath := range profilePaths {
		hasUpdated := installWrapperToProfile(profilePath, snippet)

		if hasUpdated {
			updatedCount++
			fmt.Printf("  ✓ Shell wrapper installed in: %s\n", profilePath)
		}
	}

	if updatedCount > 0 {
		printSetupSuccess(targetShell)
	} else {
		fmt.Println("  ✓ Shell wrapper is already up to date in all detected profiles.")
	}
}

func printSetupSuccess(targetShell string) {
	fmt.Println()
	fmt.Println("🎉 Movie CLI shell integration successfully installed!")
	fmt.Println("   To activate immediately in your current terminal:")

	if targetShell == "powershell" {
		fmt.Println("     . $PROFILE")
	}

	if targetShell == "bash" {
		fmt.Println("     source ~/.bashrc")
	}

	if targetShell == "zsh" {
		fmt.Println("     source ~/.zshrc")
	}

	fmt.Println("   Or simply open a new terminal window.")
}

func runAutoSetupProfiles() {
	targetShell := detectTargetShell("")
	profilePaths := resolveShellProfilePaths(targetShell)
	snippet := renderWrapperSnippet(targetShell)

	for _, profilePath := range profilePaths {
		_ = installWrapperToProfile(profilePath, snippet)
	}
}

func detectTargetShell(override string) string {
	if override != "" {
		cleaned := strings.ToLower(strings.TrimSpace(override))

		if strings.Contains(cleaned, "power") || strings.Contains(cleaned, "pwsh") {
			return "powershell"
		}

		if strings.Contains(cleaned, "bash") {
			return "bash"
		}

		if strings.Contains(cleaned, "zsh") {
			return "zsh"
		}

		return cleaned
	}

	if runtime.GOOS == "windows" {
		return "powershell"
	}

	shellEnv := os.Getenv("SHELL")

	if strings.Contains(shellEnv, "zsh") {
		return "zsh"
	}

	return "bash"
}

func resolveShellProfilePaths(targetShell string) []string {
	if targetShell == "powershell" {
		return resolvePowerShellProfilePaths()
	}

	home, _ := os.UserHomeDir()

	if targetShell == "zsh" {
		return []string{filepath.Join(home, ".zshrc")}
	}

	return []string{filepath.Join(home, ".bashrc")}
}

func resolvePowerShellProfilePaths() []string {
	var paths []string

	if envProfile := strings.TrimSpace(os.Getenv("PROFILE")); len(envProfile) > 0 {
		paths = append(paths, envProfile)
	}

	probed := probePowerShellProfilePaths()
	paths = append(paths, probed...)

	home, _ := os.UserHomeDir()

	if runtime.GOOS == "windows" {
		docs := filepath.Join(home, "Documents")
		paths = append(paths,
			filepath.Join(docs, "PowerShell", "Microsoft.PowerShell_profile.ps1"),
			filepath.Join(docs, "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"),
		)
	} else {
		paths = append(paths,
			filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1"),
		)
	}

	return deduplicatePaths(paths)
}

func probePowerShellProfilePaths() []string {
	var results []string
	engines := []string{"pwsh", "powershell"}

	for _, engine := range engines {
		cmd := exec.Command(engine, "-NoProfile", "-Command", "$PROFILE.CurrentUserAllHosts; $PROFILE.CurrentUserCurrentHost")
		out, err := cmd.Output()

		if err != nil {
			continue
		}

		lines := strings.Split(string(out), "\n")

		for _, line := range lines {
			cleaned := strings.TrimSpace(strings.TrimSuffix(line, "\r"))

			if len(cleaned) > 0 {
				results = append(results, cleaned)
			}
		}
	}

	return results
}

func deduplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	var deduped []string

	for _, p := range paths {
		trimmed := strings.TrimSpace(p)

		if len(trimmed) == 0 {
			continue
		}

		norm := filepath.Clean(trimmed)

		lookup := strings.ToLower(norm)

		if seen[lookup] {
			continue
		}

		seen[lookup] = true
		deduped = append(deduped, norm)
	}

	return deduped
}

func installWrapperToProfile(profilePath, snippet string) bool {
	dir := filepath.Dir(profilePath)
	_ = os.MkdirAll(dir, 0o755)

	contentBytes, err := os.ReadFile(profilePath)
	content := string(contentBytes)

	if err != nil && !os.IsNotExist(err) {
		return false
	}

	newContent := reconcileWrapperContent(content, snippet)

	if newContent == content {
		return false
	}

	writeErr := os.WriteFile(profilePath, []byte(newContent), 0o644)

	return writeErr == nil
}

func reconcileWrapperContent(existing, snippet string) string {
	startIndex := strings.Index(existing, wrapperMarkerStart)

	if startIndex < 0 {
		trimmed := strings.TrimRight(existing, "\r\n")

		if len(trimmed) == 0 {
			return snippet + "\n"
		}

		return trimmed + "\n\n" + snippet + "\n"
	}

	endRel := strings.Index(existing[startIndex:], wrapperMarkerEnd)

	if endRel < 0 {
		return existing + "\n\n" + snippet + "\n"
	}

	blockEnd := startIndex + endRel + len(wrapperMarkerEnd)
	replaced := existing[:startIndex] + snippet + existing[blockEnd:]

	return replaced
}

func renderWrapperSnippet(targetShell string) string {
	if targetShell == "powershell" {
		return renderPowerShellSnippet()
	}

	return renderUnixSnippet()
}

func renderPowerShellSnippet() string {
	return `# movie-cli command wrapper v1
function Get-MovieCommand {
  $cmd = Get-Command movie.exe -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($cmd) {
    return $cmd.Source
  }

  $cmd = Get-Command movie -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
  if ($cmd) {
    return $cmd.Source
  }

  return $null
}

function movie {
  $real = Get-MovieCommand
  if (-not $real) {
    Write-Error "movie executable not found"

    return
  }

  $handoff = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "movie-handoff-$([System.Guid]::NewGuid().ToString('N')).txt")
  try {
    $env:MOVIE_HANDOFF_FILE = $handoff
    $env:MOVIE_WRAPPER = "1"
    $env:MOVIE_COMMAND_WRAPPER = "1"
    & $real @args
    if ((Test-Path -LiteralPath $handoff) -and ((Get-Item -LiteralPath $handoff).Length -gt 0)) {
      $target = [string](Get-Content -LiteralPath $handoff -Raw)
      $target = $target.Trim()
      if ($target -and (Test-Path -LiteralPath ([string]$target))) {
        Set-Location -LiteralPath ([string]$target)
      }
    }
  }
  finally {
    Remove-Item -LiteralPath $handoff -ErrorAction SilentlyContinue
    Remove-Item Env:\MOVIE_HANDOFF_FILE -ErrorAction SilentlyContinue
  }
}

function mcd { movie cd @args }
function mov { movie @args }
# movie-cli command wrapper v1 end`
}

func renderUnixSnippet() string {
	return `# movie-cli command wrapper v1
mcd() {
  local dest status
  dest="$(MOVIE_COMMAND_WRAPPER=1 MOVIE_WRAPPER=1 command movie cd "$@")"
  status=$?
  if [ $status -ne 0 ]; then

    return $status
  fi
  if [ -n "$dest" ] && [ -d "$dest" ]; then
    builtin cd "$dest" || return $?
  fi
}

movie() {
  if [ "$1" = "cd" ] || [ "$1" = "go" ]; then
    local dest status
    dest="$(MOVIE_COMMAND_WRAPPER=1 MOVIE_WRAPPER=1 command movie "$@")"
    status=$?
    if [ $status -ne 0 ]; then

      return $status
    fi
    if [ -n "$dest" ] && [ -d "$dest" ]; then
      builtin cd "$dest" || return $?
    fi

    return 0
  fi
  local handoff status
  handoff="$(mktemp -t movie-handoff.XXXXXX 2>/dev/null)" || handoff=""
  if [ -n "$handoff" ]; then
    MOVIE_HANDOFF_FILE="$handoff" MOVIE_COMMAND_WRAPPER=1 MOVIE_WRAPPER=1 command movie "$@"
    status=$?
    if [ -s "$handoff" ]; then
      local target
      target="$(cat "$handoff")"
      if [ -n "$target" ] && [ -d "$target" ]; then
        builtin cd "$target" || true
      fi
    fi
    rm -f "$handoff"

    return $status
  fi
  command movie "$@"
}

mov() { movie "$@"; }
# movie-cli command wrapper v1 end`
}
