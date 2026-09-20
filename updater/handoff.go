package updater

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

// createHandoffCopy creates a temporary copy of the binary for the handoff worker.
func createHandoffCopy(selfPath string) (string, error) {
	name := handoffName()
	copyPath := filepath.Join(filepath.Dir(selfPath), name)

	if copyFile(selfPath, copyPath) == nil {
		makeExecutable(copyPath)

		return copyPath, nil
	}

	// Fallback to temp directory
	copyPath = filepath.Join(os.TempDir(), name)
	if err := copyFile(selfPath, copyPath); err != nil {
		return "", appfault.Wrap("cannot create handoff copy", err)
	}

	makeExecutable(copyPath)

	return copyPath, nil
}

// launchHandoff executes the handoff copy synchronously in the foreground with cmd.Run().
// Following the GitMap Self-Update Gold Standard, stdout, stderr, and stdin are inherited
// so the terminal stays attached and the user sees all progress output.
func launchHandoff(copyPath, repoPath, targetBinary string) error {
	args := []string{
		"update-runner",
		"--repo-path", repoPath,
		"--target-binary", targetBinary,
	}

	fmt.Printf("==> Handing off update to %s\n", copyPath)

	cmd := exec.Command(copyPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return appfault.Wrap("update worker failed", err)
	}

	// Clean up handoff copy now that worker has finished
	_ = os.Remove(copyPath)

	// Clean up any remaining .old/.bak artifacts
	_, _ = Cleanup("")

	return nil
}

// handoffName returns the temp binary name with PID suffix.
func handoffName() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("movie-update-%d.exe", os.Getpid())
	}

	return fmt.Sprintf("movie-update-%d", os.Getpid())
}

// makeExecutable sets +x permission on Unix systems.
func makeExecutable(path string) {
	if runtime.GOOS == "windows" {
		return
	}

	_ = os.Chmod(path, 0o755)
}

// copyFile copies src to dst.
func copyFile(src, dst string) error {
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
