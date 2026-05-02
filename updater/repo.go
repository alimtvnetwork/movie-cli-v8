package updater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

const (
	repoDirName    = "movie-cli-v8"
	repoModulePath = "github.com/alimtvnetwork/movie-cli-v8"
)

// findRepoPath locates the git repository root by checking (in order):
//  1. --repo-path flag
//  2. Saved RepoPath in the local DB
//  3. The directory containing the running binary
//  4. A sibling movie-cli-v8/ clone next to the binary
//  5. The current working directory
//  6. Bootstrap clone (fresh clone next to the binary)
func findRepoPath(flagPath string) (string, bool, error) {
	if strings.TrimSpace(flagPath) != "" {
		return resolveFlagRepoPath(flagPath)
	}
	if savedPath, err := loadSavedRepoPath(); err == nil {
		if repoPath, ok := resolveCandidateRepo(savedPath); ok {
			return repoPath, false, nil
		}
	}

	exe, exeErr := os.Executable()
	if exeErr == nil {
		exe, _ = filepath.EvalSymlinks(exe)
		exeDir := filepath.Dir(exe)

		// 1. Binary's own directory
		if repoPath, ok := resolveCandidateRepo(exeDir); ok {
			return repoPath, false, nil
		}

		// 2. Sibling clone
		sibling := filepath.Join(exeDir, repoDirName)
		if repoPath, ok := resolveCandidateRepo(sibling); ok {
			return repoPath, false, nil
		}
	}

	// 3. CWD
	if cwd, cwdErr := os.Getwd(); cwdErr == nil {
		if repoPath, ok := resolveCandidateRepo(cwd); ok {
			return repoPath, false, nil
		}
	}

	// 4. Bootstrap clone
	if exeErr == nil {
		exeDir := filepath.Dir(exe)
		cloneDir := filepath.Join(exeDir, repoDirName)
		fmt.Printf("📥 No local repo found. Cloning to: %s\n", cloneDir)
		if _, cloneErr := gitOutput(exeDir, "clone", "--depth", "1", repoURL); cloneErr != nil {
			return "", false, apperror.Wrap("cannot clone repository", cloneErr)
		}
		return cloneDir, true, nil
	}

	return "", false, apperror.New("cannot locate the movie-cli repository")
}

func resolveFlagRepoPath(flagPath string) (string, bool, error) {
	repoPath, err := normalizeRepoPath(flagPath)
	if err != nil {
		return "", false, err
	}
	if !isValidRepo(repoPath) {
		return "", false, apperror.New("invalid --repo-path: %s", repoPath)
	}
	return repoRoot(repoPath), false, nil
}

func resolveCandidateRepo(path string) (string, bool) {
	repoPath, err := normalizeRepoPath(path)
	if err != nil || !isValidRepo(repoPath) {
		return "", false
	}
	return repoRoot(repoPath), true
}

// isValidRepo checks if a directory is a valid movie-cli-v8 repo
// by verifying both .git and go.mod exist and the module path matches.
func isValidRepo(dir string) bool {
	if strings.TrimSpace(dir) == "" {
		return false
	}
	gitDir := filepath.Join(dir, ".git")
	goMod := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(gitDir); err != nil {
		return false
	}
	if _, err := os.Stat(goMod); err != nil {
		return false
	}
	return hasExpectedModule(goMod)
}

func hasExpectedModule(goModPath string) bool {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}
	expected := "module " + repoModulePath
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == expected {
			return true
		}
	}
	return false
}

func normalizeRepoPath(raw string) (string, error) {
	// Order matters: trim outer whitespace first, then strip surrounding
	// quotes (e.g. when a Windows path was pasted with quotes), then trim
	// any whitespace that was inside the quotes.
	path := strings.TrimSpace(raw)
	path = strings.Trim(path, `"'`)
	path = strings.TrimSpace(path)
	if path == "" {
		return "", apperror.New("repository path is empty")
	}
	expanded, err := expandHomePath(path)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", apperror.Wrap("cannot resolve repository path", err)
	}
	return filepath.Clean(absPath), nil
}

func expandHomePath(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, "~\\") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", apperror.Wrap("cannot resolve home directory", err)
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}

// repoRoot uses git to resolve the actual repo root from a subdirectory.
func repoRoot(dir string) string {
	p, err := gitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return dir
	}
	return p
}

// gitOutput runs a git command in the given directory and returns trimmed stdout.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		if text == "" {
			return "", err
		}
		return "", apperror.New("%s", text)
	}
	return text, nil
}
