// shell_handoff.go — Shell navigation handoff file and wrapper status for movie CLI.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	envMovieHandoffFile     = "MOVIE_HANDOFF_FILE"
	envMovieCommandWrapper  = "MOVIE_COMMAND_WRAPPER"
	envMovieWrapperVal      = "1"
	errShellHandoffWriteFmt = "  ⚠️ Could not write shell-handoff file %s: %v\n"
)

func writeHandoffPath(targetPath string) {
	handoffFile := os.Getenv(envMovieHandoffFile)

	if len(handoffFile) == 0 {
		return
	}

	if len(targetPath) == 0 {
		return
	}

	if err := os.WriteFile(handoffFile, []byte(targetPath), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, errShellHandoffWriteFmt, handoffFile, err)
	}
}

func isWrapperActive() bool {
	wrapperVal := os.Getenv(envMovieCommandWrapper)

	return wrapperVal == envMovieWrapperVal
}

func warnIfNoWrapper() {
	if isWrapperActive() {
		return
	}

	autoRunSetupForCd()
	printNoWrapperWarning()
}

func autoRunSetupForCd() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "  (auto-setup skipped: %v)\n", r)
		}
	}()

	runAutoSetupProfiles()
}

func printNoWrapperWarning() {
	fmt.Fprintln(os.Stderr, "  ⚠️ Command wrapper not active — path printed to stdout but cannot change shell directory.")
	fmt.Fprintln(os.Stderr, "    Shell profile has been auto-updated; reload it in your terminal:")
	fmt.Fprintln(os.Stderr, "      PowerShell: . $PROFILE")
	fmt.Fprintln(os.Stderr, "      Bash:       source ~/.bashrc")
	fmt.Fprintln(os.Stderr, "      Zsh:        source ~/.zshrc")
}

func openDirectoryInOS(targetPath string) error {
	cleanPath := filepath.Clean(targetPath)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", cleanPath)
	case "darwin":
		cmd = exec.Command("open", cleanPath)
	default:
		cmd = exec.Command("xdg-open", cleanPath)
	}

	return cmd.Start()
}
