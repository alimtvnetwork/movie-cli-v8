// self_replace.go — atomic rename-first replacement of the active PATH binary
// from a source binary (the deployed copy). This breaks the chicken-and-egg
// problem where a stale active binary keeps spawning a stale handoff worker
// that keeps deploying to the wrong drive.
//
// Why rename-first: on Windows you cannot overwrite a running .exe, but you
// CAN rename it. So the sequence is:
//
//	rename target -> target.old   (releases the name; old PID still runs target.old)
//	copy   source -> target       (new file at the canonical name)
//	(old PID self-deletes target.old on next exit, or the next cleanup pass does it)
//
// On Unix the rename trick is unnecessary (you can unlink a busy file), but the
// same code path works there too.
package updater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

// SelfReplace copies sourcePath over targetPath using rename-first semantics.
// Both paths are resolved to absolute, symlink-evaluated forms before work.
//
// Behavior:
//   - sourcePath defaults to the deployPath from powershell.json + binaryName
//     when empty.
//   - targetPath defaults to the active `movie` on PATH when empty, falling
//     back to the running executable when PATH lookup fails.
//   - If source and target resolve to the same file, it is a no-op success.
//   - The previous target is preserved as `<target>.old` and a best-effort
//     delete is attempted; failure to delete is non-fatal.
func SelfReplace(sourcePath, targetPath string) error {
	resolvedSource, err := resolveSelfReplaceSource(sourcePath)
	if err != nil {
		return err
	}
	resolvedTarget, err := resolveSelfReplaceTarget(targetPath)
	if err != nil {
		return err
	}

	if pathsEqual(resolvedSource, resolvedTarget) {
		fmt.Printf("  Source and target are the same file: %s\n", resolvedTarget)
		fmt.Println("  Nothing to do.")
		return nil
	}

	if _, statErr := os.Stat(resolvedSource); statErr != nil {
		return apperror.Wrap("source binary not found", statErr)
	}

	fmt.Printf("  Source: %s\n", resolvedSource)
	fmt.Printf("  Target: %s\n", resolvedTarget)

	backup := resolvedTarget + ".old"
	_ = os.Remove(backup) // best-effort cleanup of any stale .old

	hadExisting := false
	if _, statErr := os.Stat(resolvedTarget); statErr == nil {
		hadExisting = true
		if renameErr := os.Rename(resolvedTarget, backup); renameErr != nil {
			return apperror.Wrap("cannot rename active binary (close all terminals using it)", renameErr)
		}
		fmt.Printf("  Renamed active binary -> %s\n", filepath.Base(backup))
	}

	if copyErr := copyFile(resolvedSource, resolvedTarget); copyErr != nil {
		// Roll back if we renamed something away.
		if hadExisting {
			_ = os.Rename(backup, resolvedTarget)
			fmt.Println("  Rolled back: restored original binary")
		}
		return apperror.Wrap("cannot copy source over target", copyErr)
	}
	makeExecutable(resolvedTarget)
	fmt.Printf("  Copied new binary -> %s\n", resolvedTarget)

	if hadExisting {
		if removeErr := os.Remove(backup); removeErr != nil {
			fmt.Printf("  [WARN] Could not remove %s yet (likely still in use); will be swept on next cleanup\n", filepath.Base(backup))
		} else {
			fmt.Printf("  Removed backup: %s\n", filepath.Base(backup))
		}
	}

	verifyReplacement(resolvedTarget)
	return nil
}

func resolveSelfReplaceSource(sourcePath string) (string, error) {
	if strings.TrimSpace(sourcePath) != "" {
		return normalizeAbs(sourcePath)
	}
	cfg := loadPowershellConfig()
	if cfg == nil {
		return "", apperror.New("--from is empty and powershell.json could not be loaded; pass --from <path>")
	}
	binaryName := defaultBinaryName()
	candidate := filepath.Join(cfg.DeployPath, binaryName)
	return normalizeAbs(candidate)
}

func resolveSelfReplaceTarget(targetPath string) (string, error) {
	if strings.TrimSpace(targetPath) != "" {
		return normalizeAbs(targetPath)
	}
	if active, err := lookupActiveMovie(); err == nil && active != "" {
		return normalizeAbs(active)
	}
	self, err := os.Executable()
	if err != nil {
		return "", apperror.Wrap("cannot resolve active binary; pass --to <path>", err)
	}
	return normalizeAbs(self)
}

func normalizeAbs(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", apperror.Wrap("cannot resolve absolute path", err)
	}
	if resolved, evalErr := filepath.EvalSymlinks(abs); evalErr == nil {
		return resolved, nil
	}
	return abs, nil
}

func defaultBinaryName() string {
	if runtime.GOOS == "windows" {
		return "movie.exe"
	}
	return "movie"
}

func lookupActiveMovie() (string, error) {
	name := "movie"
	if runtime.GOOS == "windows" {
		name = "movie.exe"
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	return path, nil
}

func verifyReplacement(targetPath string) {
	cmd := exec.Command(targetPath, "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  [WARN] Verify failed: %v\n", err)
		return
	}
	trimmed := strings.TrimSpace(string(out))
	fmt.Printf("  Verified: %s\n", trimmed)
}
