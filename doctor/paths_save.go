// paths_save.go — persists deployPath configuration and platform defaults.
package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

// SaveDeployPath updates the deployPath field in powershell.json.
func SaveDeployPath(deployPath string) error {
	repo, err := findRepoRoot()
	if err != nil {
		return err
	}

	configPath := filepath.Join(repo, powershellConfigName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return appfault.Wrap("cannot read powershell.json", err)
	}

	var raw map[string]interface{}
	if parseErr := json.Unmarshal(data, &raw); parseErr != nil {
		return appfault.Wrap("cannot parse powershell.json", parseErr)
	}

	if pathsEqual(deployPath, defaultDeployDir()) {
		deployPath = ""
	}

	raw["deployPath"] = deployPath

	updated, serErr := json.MarshalIndent(raw, "", "  ")
	if serErr != nil {
		return appfault.Wrap("cannot serialize powershell.json", serErr)
	}

	updated = append(updated, '\n')

	if err := os.WriteFile(configPath, updated, 0644); err != nil {
		return appfault.Wrap("cannot write powershell.json", err)
	}

	return nil
}

func hasSourceBinary(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)

	return err == nil
}

func defaultDeployDir() string {
	if runtime.GOOS == "windows" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "movie-cli")
		}

		return filepath.Join("C:", "movie-cli")
	}

	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".local", "bin")
	}

	return "/usr/local/bin"
}

func expandScanPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	trimmed := strings.TrimPrefix(path, "~")
	trimmed = strings.TrimPrefix(trimmed, "/")
	trimmed = strings.TrimPrefix(trimmed, "\\")

	return filepath.Join(home, trimmed)
}
