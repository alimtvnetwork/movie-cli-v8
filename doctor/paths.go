// paths.go — path resolution helpers shared by checks and report.
package doctor

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

const powershellConfigName = "powershell.json"

type powershellConfig struct {
	DeployPath string `json:"deployPath"`
	BinaryName string `json:"binaryName"`
}

func resolveDeploySource() (string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", err
	}
	name := cfg.BinaryName
	if name == "" {
		name = defaultBinaryName()
	}
	return absPath(filepath.Join(cfg.DeployPath, name))
}

func resolveActiveBinary() (string, error) {
	name := defaultBinaryName()
	path, err := exec.LookPath(name)
	if err != nil {
		return "", apperror.Wrap("active binary not on PATH", err)
	}
	return absPath(path)
}

func loadConfig() (*powershellConfig, error) {
	repo, err := findRepoRoot()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(repo, powershellConfigName))
	if err != nil {
		return nil, apperror.Wrap("cannot read powershell.json", err)
	}
	var cfg powershellConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, apperror.Wrap("cannot parse powershell.json", err)
	}
	return &cfg, nil
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", apperror.Wrap("cannot get working directory", err)
	}
	dir := wd
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, powershellConfigName)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", apperror.New("powershell.json not found in cwd or parents")
}

func absPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", apperror.Wrap("cannot resolve absolute path", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
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

func pathContainsDir(dir string) bool {
	pathEnv := os.Getenv("PATH")
	sep := string(os.PathListSeparator)
	for _, entry := range strings.Split(pathEnv, sep) {
		if entry == "" {
			continue
		}
		if pathsEqual(filepath.Clean(entry), filepath.Clean(dir)) {
			return true
		}
	}
	return false
}
