// help_metadata.go — renders GitMap binary and repository identity footer blocks for help and binary commands.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/version"
)

const footerRule = "────────────────────────────────────────────────────────────"

const defaultRepoURL = "https://github.com/alimtvnetwork/movie-cli-v8"

// renderBinaryCard satisfies compatibility with existing callers and forwards to renderUsageFooter.
func renderBinaryCard(b *strings.Builder, isColor bool) {
	renderUsageFooter(b, isColor)
}

// renderUsageFooter renders the two GitMap-style identity blocks:
//  1. movie binary identity (which build is running) — magenta header
//  2. current repo identity (where you are right now)  — cyan header (if in git repo)
func renderUsageFooter(w io.Writer, isColor bool) {
	renderMovieIdentityBlock(w, isColor)

	cwd, err := os.Getwd()

	if err != nil {
		return
	}

	if isFooterGitRepo(cwd) {
		renderCurrentRepoIdentityBlock(w, isColor, cwd)
	}
}

func renderMovieIdentityBlock(w io.Writer, isColor bool) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  "+colorText(footerRule, ansiMagenta, isColor))
	fmt.Fprintln(w, "  "+colorText("movie binary", ansiMagenta, isColor))

	name := "movie"
	gitURL := resolveMovieRepoURL()
	ver := resolveMovieVersion()
	sha := resolveMovieCommitSHA()
	dbPath := resolveDatabasePath()
	installedPath := resolveInstalledBinaryPath()
	builtDate := resolveMovieBuildDate()

	bullet := colorText("●", ansiCyan, isColor)
	fmt.Fprintf(w, "  %s %s           %s\n",
		bullet, colorText("Name:", ansiCyan, isColor), colorText(name, ansiWhite, isColor))

	if len(gitURL) > 0 {
		fmt.Fprintf(w, "  %s %s        %s\n",
			bullet, colorText("Git URL:", ansiCyan, isColor), colorText(gitURL, ansiCyan, isColor))
	}

	fmt.Fprintf(w, "  %s %s        %s\n",
		bullet, colorText("Version:", ansiCyan, isColor), colorText(ver, ansiWhite, isColor))

	if len(sha) > 0 {
		fmt.Fprintf(w, "  %s %s     %s\n",
			bullet, colorText("Commit SHA:", ansiCyan, isColor), colorText(sha, ansiYellow, isColor))
	}

	if len(dbPath) > 0 {
		fmt.Fprintf(w, "  %s %s       %s\n",
			bullet, colorText("Database:", ansiCyan, isColor), colorText(dbPath, ansiWhite, isColor))
	}

	if len(installedPath) > 0 {
		fmt.Fprintf(w, "  %s %s %s\n",
			bullet, colorText("Installed path:", ansiCyan, isColor), colorText(installedPath, ansiWhite, isColor))
	}

	if len(builtDate) > 0 {
		fmt.Fprintf(w, "  %s %s          %s\n",
			bullet, colorText("Built:", ansiCyan, isColor), colorText(builtDate, ansiDim, isColor))
	}

	fmt.Fprintln(w)
}

func renderCurrentRepoIdentityBlock(w io.Writer, isColor bool, cwd string) {
	fmt.Fprintln(w, "  "+colorText(footerRule, ansiCyan, isColor))
	fmt.Fprintln(w, "  "+colorText("current repo", ansiCyan, isColor))

	repoName := resolveLocalRepoName(cwd)
	bullet := colorText("●", ansiCyan, isColor)

	if len(repoName) > 0 {
		fmt.Fprintf(w, "  %s %s                %s\n",
			bullet, colorText("Repo:", ansiCyan, isColor), colorText(repoName, ansiWhite, isColor))
	}

	gitURL := captureGit(cwd, "config", "--get", "remote.origin.url")

	if len(gitURL) > 0 {
		fmt.Fprintf(w, "  %s %s             %s\n",
			bullet, colorText("Git URL:", ansiCyan, isColor), colorText(gitURL, ansiCyan, isColor))
	}

	branch := captureGit(cwd, "rev-parse", "--abbrev-ref", "HEAD")

	if len(branch) > 0 {
		fmt.Fprintf(w, "  %s %s              %s\n",
			bullet, colorText("Branch:", ansiCyan, isColor), colorText(branch, ansiGreen, isColor))
	}

	latestBranch := resolveLatestBranch(cwd)

	if len(latestBranch) > 0 {
		fmt.Fprintf(w, "  %s %s       %s\n",
			bullet, colorText("Latest branch:", ansiCyan, isColor), colorText(latestBranch, ansiWhite, isColor))
	}

	prCount := resolveOpenPRCount(cwd)
	fmt.Fprintf(w, "  %s %s     %s\n",
		bullet, colorText("PR count (open):", ansiCyan, isColor), colorText(prCount, ansiYellow, isColor))

	branchInfo := resolveCurrentBranchInfo(cwd)

	if len(branchInfo) > 0 {
		fmt.Fprintf(w, "  %s %s %s\n",
			bullet, colorText("Current branch info:", ansiCyan, isColor), colorText(branchInfo, ansiWhite, isColor))
	}

	commit := captureGit(cwd, "log", "-1", "--format=%h · %s · %cr")

	if len(commit) > 0 {
		fmt.Fprintf(w, "  %s %s         %s\n",
			bullet, colorText("Last commit:", ansiCyan, isColor), colorText(commit, ansiYellow, isColor))
	}

	sha := captureGit(cwd, "rev-parse", "HEAD")

	if len(sha) > 0 {
		fmt.Fprintf(w, "  %s %s          %s\n",
			bullet, colorText("Commit SHA:", ansiCyan, isColor), colorText(sha, ansiYellow, isColor))
	}

	fmt.Fprintln(w)
}

func resolveMovieRepoURL() string {
	cwd, err := os.Getwd()

	if err == nil {
		u := captureGit(cwd, "config", "--get", "remote.origin.url")

		if strings.Contains(u, "movie-cli") {
			return u
		}
	}

	return defaultRepoURL
}

func resolveMovieVersion() string {
	v := version.Short()

	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}

	return v
}

func resolveMovieCommitSHA() string {
	hasCommit := version.Commit != ""

	if hasCommit {
		if version.Commit != "none" {
			return version.Commit
		}
	}

	cwd, err := os.Getwd()

	if err == nil {
		if isFooterGitRepo(cwd) {
			u := captureGit(cwd, "config", "--get", "remote.origin.url")

			if strings.Contains(u, "movie-cli") {
				sha := captureGit(cwd, "rev-parse", "HEAD")

				if sha != "" {
					return sha
				}
			}
		}
	}

	return readVersionJSONCommitSHA()
}

func readVersionJSONCommitSHA() string {
	candidates := []string{
		"version.json",
		filepath.Join(resolveInstallDir(), "version.json"),
		filepath.Join(resolveInstallDir(), "..", "version.json"),
	}

	for _, p := range candidates {
		data, err := os.ReadFile(p)

		if err != nil {
			continue
		}

		var parsed struct {
			LastCommitSha string `json:"LastCommitSha,omitempty"`
			Git           struct {
				Sha string `json:"sha"`
			} `json:"git"`
		}

		errUnmarshal := json.Unmarshal(data, &parsed)

		if errUnmarshal == nil {
			if parsed.Git.Sha != "" {
				return parsed.Git.Sha
			}

			if parsed.LastCommitSha != "" {
				return parsed.LastCommitSha
			}
		}
	}

	return ""
}

func resolveInstalledBinaryPath() string {
	exePath, err := os.Executable()

	if err != nil {
		exePath = "movie"
	}

	resolved, errEval := filepath.EvalSymlinks(exePath)

	if errEval == nil {
		exePath = resolved
	}

	absPath, errAbs := filepath.Abs(exePath)

	if errAbs == nil {
		exePath = absPath
	}

	isTemp := strings.Contains(exePath, "go-build") ||
		strings.Contains(exePath, "Temp") ||
		strings.Contains(exePath, "tmp") ||
		strings.Contains(exePath, "cache")

	if isTemp {
		binName := "movie"

		if runtime.GOOS == "windows" {
			binName = "movie.exe"
		}

		if lookPath, errLook := exec.LookPath(binName); errLook == nil {
			if realLook, errReal := filepath.EvalSymlinks(lookPath); errReal == nil {
				return realLook
			}

			return lookPath
		}

		localAppData := os.Getenv("LOCALAPPDATA")

		if localAppData != "" {
			appDataBin := filepath.Join(localAppData, "movie-cli", binName)

			if _, errStat := os.Stat(appDataBin); errStat == nil {
				return appDataBin
			}
		}

		homeDir, errHome := os.UserHomeDir()

		if errHome == nil {
			homeBin := filepath.Join(homeDir, ".movie", binName)

			if _, errStat := os.Stat(homeBin); errStat == nil {
				return homeBin
			}
		}
	}

	return exePath
}

func resolveInstallDir() string {
	return filepath.Dir(resolveInstalledBinaryPath())
}

func resolveDatabasePath() string {
	instDir := resolveInstallDir()

	isTemp := strings.Contains(instDir, "go-build") ||
		strings.Contains(instDir, "Temp") ||
		strings.Contains(instDir, "tmp") ||
		strings.Contains(instDir, "cache")

	if !isTemp {
		return filepath.Join(instDir, "data", "movie.db")
	}

	homeDir, errHome := os.UserHomeDir()

	if errHome == nil {
		userDB := filepath.Join(homeDir, ".movie", "data", "movie.db")

		if _, errStat := os.Stat(userDB); errStat == nil {
			return userDB
		}
	}

	localAppData := os.Getenv("LOCALAPPDATA")

	if localAppData != "" {
		appDataDB := filepath.Join(localAppData, "movie-cli", "data", "movie.db")

		if _, errStat := os.Stat(appDataDB); errStat == nil {
			return appDataDB
		}
	}

	cwd, errCwd := os.Getwd()

	if errCwd == nil {
		localDB := filepath.Join(cwd, "data", "movie.db")

		if _, errStat := os.Stat(localDB); errStat == nil {
			return localDB
		}
	}

	if homeDir != "" {
		return filepath.Join(homeDir, ".movie", "data", "movie.db")
	}

	return filepath.Join(instDir, "data", "movie.db")
}

func resolveMovieBuildDate() string {
	hasBuildDate := version.BuildDate != ""

	if hasBuildDate {
		if version.BuildDate != "unknown" {
			return version.BuildDate
		}
	}

	candidates := []string{"version.json", filepath.Join(resolveInstallDir(), "version.json")}

	for _, p := range candidates {
		data, err := os.ReadFile(p)

		if err != nil {
			continue
		}

		var parsed struct {
			ReleaseDate string `json:"releaseDate"`
		}

		errUnmarshal := json.Unmarshal(data, &parsed)

		if errUnmarshal == nil {
			if parsed.ReleaseDate != "" {
				return parsed.ReleaseDate
			}
		}
	}

	return ""
}

func isFooterGitRepo(dir string) bool {
	out := captureGit(dir, "rev-parse", "--is-inside-work-tree")

	return out == "true"
}

func captureGit(dir string, args ...string) string {
	if len(dir) == 0 {
		return ""
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()

	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func resolveLocalRepoName(dir string) string {
	top := captureGit(dir, "rev-parse", "--show-toplevel")

	if len(top) > 0 {
		return filepath.Base(top)
	}

	return filepath.Base(dir)
}

func resolveLatestBranch(dir string) string {
	latest := captureGit(dir, "for-each-ref", "--sort=-committerdate", "refs/heads/", "--format=%(refname:short) (%(committerdate:relative))", "--count=1")

	if len(latest) > 0 {
		return latest
	}

	return captureGit(dir, "rev-parse", "--abbrev-ref", "HEAD")
}

func resolveOpenPRCount(dir string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", "pr", "list", "--state", "open", "--limit", "100", "--json", "number", "--jq", "length")
	cmd.Dir = dir
	out, err := cmd.Output()

	if err != nil {
		return "0"
	}

	count := strings.TrimSpace(string(out))

	if len(count) == 0 {
		return "0"
	}

	return count
}

func resolveBranchDirtyStatus(dir string) string {
	out := captureGit(dir, "status", "--porcelain")

	if len(out) == 0 {
		return "clean"
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")

	return fmt.Sprintf("dirty (%d changed)", len(lines))
}

func formatSyncCounts(ahead, behind string) string {
	isBothZero := ahead == "0" && behind == "0"

	if isBothZero {
		return "up to date"
	}

	isAheadOnly := ahead != "0" && behind == "0"

	if isAheadOnly {
		return fmt.Sprintf("ahead %s", ahead)
	}

	isBehindOnly := ahead == "0" && behind != "0"

	if isBehindOnly {
		return fmt.Sprintf("behind %s", behind)
	}

	return fmt.Sprintf("ahead %s, behind %s", ahead, behind)
}

func resolveBranchSyncStatus(dir string) string {
	counts := captureGit(dir, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")

	if len(counts) == 0 {
		return ""
	}

	parts := strings.Fields(counts)

	if len(parts) != 2 {
		return ""
	}

	return formatSyncCounts(parts[0], parts[1])
}

func resolveCurrentBranchInfo(dir string) string {
	dirtyStatus := resolveBranchDirtyStatus(dir)
	syncStatus := resolveBranchSyncStatus(dir)

	if len(syncStatus) > 0 {
		return dirtyStatus + " · " + syncStatus
	}

	return dirtyStatus
}
