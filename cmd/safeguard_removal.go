// safeguard_removal.go — Strict safeguards preventing root folder and multi-movie deletions.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// RemovalTargetInfo contains the resolved target and metadata for safe removal.
type RemovalTargetInfo struct {
	Title             string
	TargetPath        string
	ScanRootFolder    string
	SidecarPaths      []string
	MediaID           int64
	IsDedicatedFolder bool
}

// validateRemovalTarget verifies that a target path is strictly safe to remove.
// Prevents root folder deletion, system directory deletion, and multi-movie folder destruction.
func validateRemovalTarget(targetPath string, database *db.DB) error {
	trimmed := strings.TrimSpace(targetPath)

	if len(trimmed) == 0 {
		return fmt.Errorf("safeguard violation: target path cannot be empty")
	}

	cleanTarget := filepath.Clean(trimmed)

	if isFilesystemRoot(cleanTarget) {
		return fmt.Errorf("CRITICAL SAFEGUARD: target %q is a filesystem root and CANNOT be removed", targetPath)
	}

	if isSystemDirectory(cleanTarget) {
		return fmt.Errorf("CRITICAL SAFEGUARD: target %q is a system directory and CANNOT be removed", targetPath)
	}

	if database == nil {
		return nil
	}

	stats, err := database.ListScanFoldersWithStats()

	if err != nil {
		return fmt.Errorf("safeguard check failed: unable to verify registered scan roots: %w", err)
	}

	for _, s := range stats {
		cleanRoot := filepath.Clean(strings.TrimSpace(s.FolderPath))

		if len(cleanRoot) == 0 {
			continue
		}

		if isPathsIdentical(cleanTarget, cleanRoot) {
			return fmt.Errorf("CRITICAL SAFEGUARD: target %q is registered scan root folder %q and CANNOT be removed", targetPath, cleanRoot)
		}

		if isPathAncestorOf(cleanTarget, cleanRoot) {
			return fmt.Errorf("CRITICAL SAFEGUARD: target %q is an ancestor of scan root folder %q and CANNOT be removed", targetPath, cleanRoot)
		}
	}

	fi, statErr := os.Stat(cleanTarget)

	if statErr == nil {
		if fi.IsDir() {
			count, countErr := countActiveMoviesInDir(database, cleanTarget)

			if countErr == nil {
				if count > 1 {
					return fmt.Errorf("SAFEGUARD: folder %q contains %d active movies in library; whole folder cannot be removed (delete individual movies instead)", targetPath, count)
				}
			}
		}
	}

	return nil
}

// resolveMediaRemovalTarget determines whether to remove the dedicated folder or the single movie file.
func resolveMediaRemovalTarget(media *db.Media, database *db.DB) (*RemovalTargetInfo, error) {
	if media == nil {
		return nil, fmt.Errorf("media record cannot be nil")
	}

	filePath := strings.TrimSpace(media.CurrentFilePath)

	if len(filePath) == 0 {
		filePath = strings.TrimSpace(media.OriginalFilePath)
	}

	if len(filePath) == 0 {
		return nil, fmt.Errorf("media #%d (%s) has no file path registered", media.ID, media.Title)
	}

	cleanFile := filepath.Clean(filePath)
	scanRoot := findContainingScanRoot(cleanFile, database)
	parentDir := filepath.Dir(cleanFile)

	targetInfo := &RemovalTargetInfo{
		MediaID:           media.ID,
		Title:             media.Title,
		ScanRootFolder:    scanRoot,
		IsDedicatedFolder: false,
		TargetPath:        cleanFile,
	}

	isParentScanRoot := false

	if len(scanRoot) > 0 {
		if isPathsIdentical(parentDir, scanRoot) {
			isParentScanRoot = true
		}
	}

	if isParentScanRoot {
		targetInfo.IsDedicatedFolder = false
		targetInfo.TargetPath = cleanFile
		targetInfo.SidecarPaths = findSidecarFiles(cleanFile)

		return targetInfo, validateRemovalTarget(targetInfo.TargetPath, database)
	}

	otherMoviesCount, err := countOtherActiveMoviesInDir(database, parentDir, media.ID)

	if err == nil {
		if otherMoviesCount == 0 {
			isDedicatedDir := checkDirectoryIsDedicated(parentDir, cleanFile)

			if isDedicatedDir {
				targetInfo.IsDedicatedFolder = true
				targetInfo.TargetPath = parentDir

				return targetInfo, validateRemovalTarget(targetInfo.TargetPath, database)
			}
		}
	}

	targetInfo.IsDedicatedFolder = false
	targetInfo.TargetPath = cleanFile
	targetInfo.SidecarPaths = findSidecarFiles(cleanFile)

	return targetInfo, validateRemovalTarget(targetInfo.TargetPath, database)
}

func isFilesystemRoot(p string) bool {
	clean := filepath.Clean(p)

	if runtime.GOOS == "windows" {
		vol := filepath.VolumeName(clean)

		if clean == vol || clean == vol+`\` || clean == vol+`/` {
			return true
		}

		if len(clean) <= 3 {
			return true
		}
	}

	if clean == "/" || clean == "." {
		return true
	}

	return false
}

func isSystemDirectory(p string) bool {
	clean := strings.ToLower(filepath.Clean(p))

	if runtime.GOOS == "windows" {
		winDir := strings.ToLower(os.Getenv("SystemRoot"))

		if len(winDir) > 0 {
			if strings.HasPrefix(clean, strings.ToLower(winDir)) {
				return true
			}
		}

		progFiles := strings.ToLower(os.Getenv("ProgramFiles"))

		if len(progFiles) > 0 {
			if strings.HasPrefix(clean, progFiles) {
				return true
			}
		}

		progFilesX86 := strings.ToLower(os.Getenv("ProgramFiles(x86)"))

		if len(progFilesX86) > 0 {
			if strings.HasPrefix(clean, progFilesX86) {
				return true
			}
		}
	}

	home, err := os.UserHomeDir()

	if err == nil {
		cleanHome := strings.ToLower(filepath.Clean(home))

		if clean == cleanHome {
			return true
		}
	}

	return false
}

func isPathsIdentical(pathA, pathB string) bool {
	normA := filepath.Clean(strings.TrimSpace(pathA))
	normB := filepath.Clean(strings.TrimSpace(pathB))

	if runtime.GOOS == "windows" {
		return strings.EqualFold(normA, normB)
	}

	return normA == normB
}

func isPathAncestorOf(potentialAncestor, targetPath string) bool {
	anc := filepath.Clean(strings.TrimSpace(potentialAncestor))
	tgt := filepath.Clean(strings.TrimSpace(targetPath))

	if runtime.GOOS == "windows" {
		anc = strings.ToLower(anc)
		tgt = strings.ToLower(tgt)
	}

	if anc == tgt {
		return false
	}

	prefix := anc + string(filepath.Separator)

	return strings.HasPrefix(tgt, prefix)
}

func findContainingScanRoot(targetFile string, database *db.DB) string {
	if database == nil {
		return ""
	}

	stats, err := database.ListScanFoldersWithStats()

	if err != nil {
		return ""
	}

	var bestMatch string
	cleanFile := filepath.Clean(targetFile)

	for _, s := range stats {
		cleanRoot := filepath.Clean(strings.TrimSpace(s.FolderPath))

		if isPathAncestorOf(cleanRoot, cleanFile) {
			if len(cleanRoot) > len(bestMatch) {
				bestMatch = cleanRoot
			}
		}
	}

	return bestMatch
}

func countActiveMoviesInDir(database *db.DB, dir string) (int, error) {
	prefix := filepath.Clean(dir) + string(filepath.Separator) + "%"
	row := database.QueryRow(`
		SELECT COUNT(DISTINCT MediaId)
		FROM Media
		WHERE IsDeleted = 0
		  AND (CurrentFilePath LIKE ? OR CurrentFilePath = ?)`, prefix, dir)

	var count int
	err := row.Scan(&count)

	return count, err
}

func countOtherActiveMoviesInDir(database *db.DB, dir string, currentMediaID int64) (int, error) {
	prefix := filepath.Clean(dir) + string(filepath.Separator) + "%"
	row := database.QueryRow(`
		SELECT COUNT(DISTINCT MediaId)
		FROM Media
		WHERE IsDeleted = 0
		  AND MediaId != ?
		  AND (CurrentFilePath LIKE ? OR CurrentFilePath = ?)`, currentMediaID, prefix, dir)

	var count int
	err := row.Scan(&count)

	return count, err
}

func checkDirectoryIsDedicated(dir, currentVideoFile string) bool {
	entries, err := os.ReadDir(dir)

	if err != nil {
		return false
	}

	currentBase := strings.ToLower(filepath.Base(currentVideoFile))
	videoExts := map[string]bool{
		".mkv": true, ".mp4": true, ".avi": true, ".mov": true,
		".wmv": true, ".flv": true, ".m4v": true, ".webm": true,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.ToLower(entry.Name())

		if name == currentBase {
			continue
		}

		ext := strings.ToLower(filepath.Ext(name))

		if videoExts[ext] {
			return false
		}
	}

	return true
}

func findSidecarFiles(videoFile string) []string {
	dir := filepath.Dir(videoFile)
	baseNoExt := strings.TrimSuffix(filepath.Base(videoFile), filepath.Ext(videoFile))
	entries, err := os.ReadDir(dir)

	if err != nil {
		return nil
	}

	var sidecars []string
	sidecarExts := map[string]bool{
		".nfo": true, ".srt": true, ".vtt": true, ".sub": true,
		".idx": true, ".jpg": true, ".png": true,
	}

	lowerBase := strings.ToLower(baseNoExt)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		entryBase := strings.ToLower(strings.TrimSuffix(name, ext))

		if sidecarExts[ext] {
			if entryBase == lowerBase || strings.HasPrefix(entryBase, lowerBase) {
				sidecars = append(sidecars, filepath.Join(dir, name))
			}
		}
	}

	return sidecars
}
