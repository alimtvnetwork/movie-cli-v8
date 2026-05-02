package updater

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
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
		return "", apperror.Wrap("cannot create handoff copy", err)
	}
	makeExecutable(copyPath)
	return copyPath, nil
}

// launchHandoff starts the worker DETACHED with its own console and returns
// immediately so the parent process can exit. Exiting the parent releases the
// OS file lock on the original binary, which is the entire reason the
// copy-and-handoff dance exists in the first place.
//
// See spec/13-self-update-app-update/03-copy-and-handoff.md and
// HANDOFF-LESSONS.md before changing this. Do NOT switch back to a blocking
// cmd.Run() — that re-introduces the Windows file-lock bug.
func launchHandoff(copyPath, repoPath, targetBinary string) error {
	args := []string{
		"update-runner",
		"--repo-path", repoPath,
		"--target-binary", targetBinary,
	}

	fmt.Printf("  🚀 Update handed off to %s\n", copyPath)
	fmt.Println("  ↪  Worker is taking over in a new window; this terminal is free.")

	cmd := exec.Command(copyPath, args...)
	configureDetached(cmd)

	if err := cmd.Start(); err != nil {
		return apperror.Wrap("cannot start update worker", err)
	}
	if cmd.Process != nil {
		_ = cmd.Process.Release()
	}
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
