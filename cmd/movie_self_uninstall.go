// movie_self_uninstall.go — uninstalls movie CLI binary, data, and PATH entries.
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

var (
	selfUninstallYes      bool
	selfUninstallKeepData bool
	selfUninstallDryRun   bool
)

var movieSelfUninstallCmd = &cobra.Command{
	Use:     "self-uninstall",
	Aliases: []string{"uninstall"},
	Short:   "Uninstall movie CLI binary, data, and PATH entries",
	Long: `self-uninstall removes the movie CLI binary, installation directory,
and optionally user data at ~/.movie.

On Windows, running binaries are locked by the OS; self-uninstall spawns a
temporary handoff process to safely delete the binary after exit.`,
	Run: runMovieSelfUninstall,
}

func init() {
	movieSelfUninstallCmd.Flags().BoolVarP(&selfUninstallYes, "yes", "y", false, "Confirm uninstallation without prompting")
	movieSelfUninstallCmd.Flags().BoolVar(&selfUninstallKeepData, "keep-data", false, "Preserve ~/.movie user configuration and database")
	movieSelfUninstallCmd.Flags().BoolVar(&selfUninstallDryRun, "dry-run", false, "Show removal actions without executing them")
}

func runMovieSelfUninstall(cmd *cobra.Command, args []string) {
	selfPath, err := os.Executable()
	if err != nil {
		errlog.Error("Could not resolve movie binary path: %v", err)
		return
	}

	installDir := filepath.Dir(selfPath)
	userDataDir := resolveUserDataDir()

	if selfUninstallDryRun {
		printSelfUninstallDryRun(selfPath, installDir, userDataDir)
		return
	}

	if !selfUninstallYes {
		if !promptSelfUninstallConfirm(selfPath, userDataDir) {
			fmt.Println("Uninstallation aborted.")
			return
		}
	}

	executeSelfUninstall(selfPath, installDir, userDataDir)
}

func resolveUserDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, ".movie")
}

func printSelfUninstallDryRun(selfPath, installDir, userDataDir string) {
	fmt.Println("movie CLI self-uninstall dry-run preview:")
	fmt.Printf("  Binary      : %s\n", selfPath)
	fmt.Printf("  Install Dir : %s\n", installDir)
	if !selfUninstallKeepData {
		fmt.Printf("  User Data   : %s\n", userDataDir)
	}

	fmt.Println("\nDry-run complete. No files were removed.")
}

func promptSelfUninstallConfirm(selfPath, userDataDir string) bool {
	fmt.Println("This will remove the movie CLI installation:")
	fmt.Printf("  Binary: %s\n", selfPath)
	if !selfUninstallKeepData {
		fmt.Printf("  Data  : %s\n", userDataDir)
	}

	fmt.Print("\nProceed with uninstallation? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	cleaned := strings.TrimSpace(strings.ToLower(answer))

	return cleaned == "y" || cleaned == "yes"
}

func executeSelfUninstall(selfPath, installDir, userDataDir string) {
	fmt.Println("Uninstalling movie CLI...")

	cleanUserData(userDataDir)
	cleanPathEntries(installDir)

	if runtime.GOOS == "windows" {
		executeWindowsHandoff(selfPath, installDir)
		return
	}

	executeUnixUninstall(selfPath, installDir)
}

func cleanUserData(userDataDir string) {
	if selfUninstallKeepData {
		fmt.Printf("  Preserved user data at %s\n", userDataDir)
		return
	}

	if userDataDir != "" {
		_ = os.RemoveAll(userDataDir)
		fmt.Printf("  Removed user data at %s\n", userDataDir)
	}
}

func cleanPathEntries(installDir string) {
	if runtime.GOOS == "windows" {
		cleanWindowsPath(installDir)
		return
	}

	cleanUnixPath(installDir)
}

func cleanWindowsPath(installDir string) {
	psScript := fmt.Sprintf(
		`$p = [Environment]::GetEnvironmentVariable('PATH', 'User'); `+
			`if ($p) { $n = ($p -split ';' | Where-Object { $_.Trim() -and ($_.Trim() -ine '%s') }) -join ';'; `+
			`[Environment]::SetEnvironmentVariable('PATH', $n, 'User') }`,
		installDir,
	)
	_ = exec.Command("powershell", "-NoProfile", "-Command", psScript).Run()
}

func cleanUnixPath(installDir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	candidates := []string{
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".profile"),
	}

	for _, path := range candidates {
		stripLineFromFile(path, installDir)
	}
}

func stripLineFromFile(filePath, needle string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	var kept []string
	for _, line := range lines {
		if !strings.Contains(line, needle) {
			kept = append(kept, line)
		}
	}

	_ = os.WriteFile(filePath, []byte(strings.Join(kept, "\n")), 0o644)
}

func executeUnixUninstall(selfPath, installDir string) {
	if err := os.Remove(selfPath); err == nil {
		fmt.Printf("  Removed binary %s\n", selfPath)
	}

	_ = os.Remove(installDir)
	fmt.Println("Uninstallation complete.")
}

func executeWindowsHandoff(selfPath, installDir string) {
	handoffPath := filepath.Join(os.TempDir(), fmt.Sprintf("movie-handoff-%d.exe", os.Getpid()))
	if err := copyBinary(selfPath, handoffPath); err != nil {
		errlog.Error("Could not create handoff copy: %v", err)
		return
	}

	cmdStr := fmt.Sprintf("ping 127.0.0.1 -n 2 > nul & del /F /Q \"%s\" & rmdir \"%s\" 2> nul & del /F /Q \"%s\"",
		selfPath, installDir, handoffPath)

	cmd := exec.Command("cmd.exe", "/C", cmdStr)
	if err := cmd.Start(); err != nil {
		errlog.Error("Failed to schedule self-deletion: %v", err)
		return
	}

	fmt.Println("  Scheduled binary removal upon exit.")
	fmt.Println("Uninstallation complete.")
}

func copyBinary(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, in)

	return err
}
