// movie_popout_cleanup.go — folder compaction phase for popout command.
//
// REPLACES the previous destructive "remove folder" prompt. After media
// files are popped out to root, any subfolder that:
//
//  1. was originally empty, OR
//  2. contains zero media files (only samples/subs/.nfo/.txt/etc.)
//
// is MOVED into <root>/.temp/ instead of deleted. Each move is recorded as
// a FileActionCompact action so `movie undo --batch <id>` can restore the
// folder to its original location.
//
// User-facing surface:
//   - With no flag         → interactive prompt: y / s (select) / n (keep) / l (list)
//   - With --auto-compact  → no prompt; every qualifying folder goes to .temp/
//
// See spec/09-app-issues/08-popout-silent-failure.md and
// mem://features/popout-spec.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/cleaner"
	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

// compactNonMediaFolders is the public entry point for the compaction phase.
// It scans every direct subfolder of rootDir, classifies each as
// "has-media" vs "no-media", and offers (or auto-applies) compaction into
// <root>/.temp/.
//
// folderNames is the pre-discovered list of direct subfolder names that
// existed at the start of the popout run. Anything new (e.g. .temp/ itself)
// is skipped automatically.
func compactNonMediaFolders(cc CleanupContext, rootDir string, folderNames []string, autoCompact bool) int {
	candidates := classifyCompactCandidates(rootDir, folderNames)
	if len(candidates) == 0 {
		return 0
	}

	printCompactSummary(candidates)

	if autoCompact {
		fmt.Println("\n  ⚙️  --auto-compact: moving all candidates into .temp/")
		return applyCompactAll(cc, rootDir, candidates)
	}
	return promptCompactAction(cc, rootDir, candidates)
}

// classifyCompactCandidates returns subfolders that should be candidates
// for compaction (no media inside, OR originally empty).
func classifyCompactCandidates(rootDir string, folderNames []string) []popoutFolderInfo {
	var candidates []popoutFolderInfo
	for _, name := range folderNames {
		if name == popoutTempDir {
			continue
		}
		dirPath := filepath.Join(rootDir, name)
		info, statErr := os.Stat(dirPath)
		if statErr != nil || !info.IsDir() {
			continue
		}
		fi := scanFolderContents(name, dirPath)
		if folderHasMedia(dirPath) {
			continue
		}
		candidates = append(candidates, fi)
	}
	return candidates
}

// folderHasMedia reports whether the folder (recursively) contains any
// video file. Used to decide whether to compact it.
func folderHasMedia(dirPath string) bool {
	hasMedia := false
	_ = filepath.Walk(dirPath, func(p string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil || fi == nil {
			return walkErr
		}
		if fi.IsDir() {
			return nil
		}
		if cleaner.IsVideoFile(fi.Name()) {
			hasMedia = true
			return filepath.SkipDir
		}
		return nil
	})
	return hasMedia
}

func scanFolderContents(name, dirPath string) popoutFolderInfo {
	var files []string
	var totalSize int64
	_ = filepath.Walk(dirPath, func(p string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fi.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dirPath, p)
		files = append(files, fmt.Sprintf("%s (%s)", rel, humanSize(fi.Size())))
		totalSize += fi.Size()
		return nil
	})
	return popoutFolderInfo{name: name, path: dirPath, files: files, totalSize: totalSize}
}

func printCompactSummary(folders []popoutFolderInfo) {
	fmt.Println("  📦 Folders eligible for .temp/ compaction:")
	fmt.Println()
	for i, f := range folders {
		if len(f.files) == 0 {
			fmt.Printf("  %d. %s/   (empty)\n", i+1, f.name)
			continue
		}
		fmt.Printf("  %d. %s/   (%d non-media files, %s)\n",
			i+1, f.name, len(f.files), humanSize(f.totalSize))
	}
}

func promptCompactAction(cc CleanupContext, rootDir string, folders []popoutFolderInfo) int {
	fmt.Println()
	fmt.Println("  Options:")
	fmt.Println("    [a] Compact ALL listed folders into .temp/")
	fmt.Println("    [s] Select folders one by one")
	fmt.Println("    [l] List files in each folder before deciding")
	fmt.Println("    [n] Keep all folders in place")
	fmt.Print("\n  Choose [a/s/l/N]: ")

	if !cc.Scanner.Scan() {
		fmt.Println("  📁 No folders compacted.")
		return 0
	}
	choice := strings.ToLower(strings.TrimSpace(cc.Scanner.Text()))

	switch choice {
	case "a":
		return applyCompactAll(cc, rootDir, folders)
	case "s":
		return selectiveCompact(cc, rootDir, folders)
	case "l":
		return listThenCompact(cc, rootDir, folders)
	default:
		fmt.Println("  📁 All folders kept.")
	}
	return 0
}

func applyCompactAll(cc CleanupContext, rootDir string, folders []popoutFolderInfo) int {
	count := 0
	for _, f := range folders {
		if compactFolder(cc, rootDir, f) != "" {
			count++
		}
	}
	return count
}

func selectiveCompact(cc CleanupContext, rootDir string, folders []popoutFolderInfo) int {
	count := 0
	for _, f := range folders {
		status := "empty"
		if len(f.files) > 0 {
			status = fmt.Sprintf("%d files (%s)", len(f.files), humanSize(f.totalSize))
		}
		fmt.Printf("\n  %s/ — %s\n", f.name, status)
		fmt.Print("    Compact to .temp/? [y/N]: ")
		if !cc.Scanner.Scan() {
			return count
		}
		ans := strings.ToLower(strings.TrimSpace(cc.Scanner.Text()))
		if ans == "y" || ans == "yes" {
			if compactFolder(cc, rootDir, f) != "" {
				count++
			}
			continue
		}
		fmt.Println("    Kept.")
	}
	return count
}

func listThenCompact(cc CleanupContext, rootDir string, folders []popoutFolderInfo) int {
	count := 0
	for _, f := range folders {
		fmt.Printf("\n  📁 %s/\n", f.name)
		printFolderListing(f)
		fmt.Print("    Compact to .temp/? [y/N]: ")
		if !cc.Scanner.Scan() {
			return count
		}
		ans := strings.ToLower(strings.TrimSpace(cc.Scanner.Text()))
		if ans == "y" || ans == "yes" {
			if compactFolder(cc, rootDir, f) != "" {
				count++
			}
			continue
		}
		fmt.Println("    Kept.")
	}
	return count
}

func printFolderListing(f popoutFolderInfo) {
	if len(f.files) == 0 {
		fmt.Println("    (empty)")
		return
	}
	for _, file := range f.files {
		fmt.Printf("    - %s\n", file)
	}
}

// compactFolder is the actual filesystem move. It is the only place that
// performs the .temp/ destination move and the only place that records the
// FileActionCompact history row.
//
// Returns the destination path (or "" on failure) so tests can assert it.
func compactFolder(cc CleanupContext, rootDir string, f popoutFolderInfo) string {
	tempDir := filepath.Join(rootDir, popoutTempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		errlog.Error("Cannot create .temp/ dir: %v", err)
		return ""
	}

	destPath := filepath.Join(tempDir, f.name)
	if _, statErr := os.Stat(destPath); statErr == nil {
		// Collision — append a numeric suffix instead of overwriting.
		destPath = uniqueTempPath(destPath)
	}

	if err := MoveFile(f.path, destPath); err != nil {
		errlog.Error("Cannot compact %s: %v", f.name, err)
		return ""
	}
	fmt.Printf("    📦 Compacted: %s/  →  .temp/%s\n", f.name, filepath.Base(destPath))

	snapshot := fmt.Sprintf(`{"original_path":%q,"compact_path":%q}`, f.path, destPath)
	detail := fmt.Sprintf("Compacted folder %s/ → .temp/%s", f.name, filepath.Base(destPath))
	_, _ = cc.Database.InsertActionSimple(db.ActionSimpleInput{
		FileAction: db.FileActionCompact,
		Snapshot:   snapshot,
		Detail:     detail,
		BatchID:    cc.BatchID,
	})
	return destPath
}

// uniqueTempPath returns destPath with a -2/-3/... suffix appended until it
// no longer exists. Defensive: avoids overwriting on collisions inside .temp/.
func uniqueTempPath(destPath string) string {
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s-%d", destPath, i)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return destPath
}

// (No re-export needed — db.FileActionCompact is referenced directly above.)
