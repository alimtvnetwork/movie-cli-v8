package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
	"github.com/alimtvnetwork/movie-cli-v8/version"
)

const (
	remoteInstallerPwsh = "https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install.ps1"
	remoteInstallerBash = "https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install.sh"
)

// ResolveCurrentInstallDir returns the directory containing the currently running binary,
// or falls back to the active binary on PATH.
func ResolveCurrentInstallDir() string {
	selfPath, err := os.Executable()
	if err == nil {
		if realPath, errEval := filepath.EvalSymlinks(selfPath); errEval == nil {
			selfPath = realPath
		}

		return filepath.Dir(selfPath)
	}

	activePath, errLook := exec.LookPath("movie.exe")
	if errLook == nil {
		if realPath, errEval := filepath.EvalSymlinks(activePath); errEval == nil {
			activePath = realPath
		}

		return filepath.Dir(activePath)
	}

	return ""
}

// ResolveCurrentBinaryPath returns the path of the currently running binary,
// or falls back to the active binary on PATH.
func ResolveCurrentBinaryPath() string {
	selfPath, err := os.Executable()
	if err == nil {
		if realPath, errEval := filepath.EvalSymlinks(selfPath); errEval == nil {
			selfPath = realPath
		}

		return selfPath
	}

	binName := "movie.exe"
	if runtime.GOOS != "windows" {
		binName = "movie"
	}

	activePath, errLook := exec.LookPath(binName)
	if errLook == nil {
		if realPath, errEval := filepath.EvalSymlinks(activePath); errEval == nil {
			activePath = realPath
		}

		return activePath
	}

	return ""
}

// HasNetworkConnectivity tests if GitHub is reachable.
func HasNetworkConnectivity() bool {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Head("https://raw.githubusercontent.com")
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode < 500
}

// RunRemoteUpdate downloads and runs the canonical installer.
func RunRemoteUpdate(installDir string) error {
	hasNet := HasNetworkConnectivity()
	if !hasNet {
		return appfault.New("offline: cannot reach GitHub for remote update")
	}

	url := remoteInstallerPwsh
	ext := ".ps1"
	if runtime.GOOS != "windows" {
		url = remoteInstallerBash
		ext = ".sh"
	}

	if installDir == "" {
		installDir = ResolveCurrentInstallDir()
	}

	activeBin := ResolveCurrentBinaryPath()
	currentVer := version.Version

	fmt.Println()
	fmt.Println("  +=============================================+")
	fmt.Println("  |  movie-cli updater                          |")
	fmt.Println("  +=============================================+")
	if activeBin != "" {
		fmt.Printf("  • Active binary : %s\n", activeBin)
	}
	fmt.Printf("  • Current ver   : %s\n", currentVer)
	fmt.Println("  • Update method : remote release installer (canonical)")
	fmt.Println()

	fmt.Printf("  ■ Fetching remote installer: %s\n", url)

	scriptPath, err := downloadRemoteInstaller(url, ext)
	if err != nil {
		return appfault.Wrap("cannot download remote installer", err)
	}

	defer os.Remove(scriptPath)

	fmt.Printf("  ■ Running installer for directory: %s\n", installDir)

	if errRun := executeRemoteInstaller(scriptPath, installDir); errRun != nil {
		return appfault.Wrap("remote installer execution failed", errRun)
	}

	newVer := resolveNewVersion(installDir)

	fmt.Printf("\n  ✓ Successfully updated from %s to %s\n  → Source: %s\n\n",
		formatVersion(currentVer),
		formatVersion(newVer),
		url,
	)
	_ = printPostUpdateIdentity(installDir)

	return nil
}

func formatVersion(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}

	return v
}

func resolveNewVersion(installDir string) string {
	binName := "movie.exe"
	if runtime.GOOS != "windows" {
		binName = "movie"
	}

	binPath := filepath.Join(installDir, binName)
	if _, err := os.Stat(binPath); err == nil {
		cmd := exec.Command(binPath, "version")
		out, errRun := cmd.Output()
		if errRun == nil {
			verStr := strings.TrimSpace(string(out))
			if verStr != "" {
				fields := strings.Fields(verStr)
				if len(fields) > 0 {
					return fields[0]
				}

				return verStr
			}
		}
	}

	return version.Version
}

func printPostUpdateIdentity(installDir string) error {
	binName := "movie.exe"
	if runtime.GOOS != "windows" {
		binName = "movie"
	}

	binPath := filepath.Join(installDir, binName)
	if _, err := os.Stat(binPath); err != nil {
		return err
	}

	cmd := exec.Command(binPath, "binary")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func downloadRemoteInstaller(url, ext string) (string, error) {
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "movie-update-*"+ext)
	if err != nil {
		return "", err
	}

	defer tmpFile.Close()

	if ext == ".ps1" {
		bom := []byte{0xEF, 0xBB, 0xBF}
		if _, err := tmpFile.Write(bom); err != nil {
			return "", err
		}
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", err
	}

	if ext == ".sh" {
		_ = os.Chmod(tmpFile.Name(), 0755)
	}

	return tmpFile.Name(), nil
}

func executeRemoteInstaller(scriptPath, installDir string) error {
	if runtime.GOOS == "windows" {
		args := []string{
			"-ExecutionPolicy", "Bypass",
			"-NoProfile", "-NoLogo",
			"-File", scriptPath,
		}

		if installDir != "" {
			args = append(args, "-InstallDir", installDir)
		}

		cmd := exec.Command("powershell", args...)
		cmd.Env = append(os.Environ(), "MOVIE_UPDATING=1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Dir = filepath.Dir(scriptPath)

		return cmd.Run()
	}

	shell := "bash"
	if _, err := exec.LookPath(shell); err != nil {
		shell = "sh"
	}

	args := []string{scriptPath}
	if installDir != "" {
		args = append(args, "--dir", installDir)
	}

	cmd := exec.Command(shell, args...)
	cmd.Env = append(os.Environ(), "MOVIE_UPDATING=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = filepath.Dir(scriptPath)

	return cmd.Run()
}
