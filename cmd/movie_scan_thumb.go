// movie_scan_thumb.go — cached thumbnail reuse and synchronization for scan outputs.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ensureThumbnailInOutputDir ensures a media thumbnail is present in outputDir/thumbnails.
func ensureThumbnailInOutputDir(outputDir, basePath, thumbPath string) {
	fileName := extractThumbFileName(thumbPath)
	if fileName == "" || outputDir == "" {
		return
	}

	destDir := filepath.Join(outputDir, "thumbnails")
	destPath := filepath.Join(destDir, fileName)
	if hasValidThumbFile(destPath) {
		return
	}

	_ = os.MkdirAll(destDir, 0755)
	copyCachedThumbToDest(basePath, thumbPath, fileName, destPath)
}

func extractThumbFileName(path string) string {
	if path == "" {
		return ""
	}

	clean := strings.ReplaceAll(path, "\\", "/")
	idx := strings.LastIndex(clean, "/")
	if idx != -1 {
		clean = clean[idx+1:]
	}

	if clean == "." {
		return ""
	}

	return clean
}

func hasValidThumbFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Size() > 0
}

func copyCachedThumbToDest(basePath, thumbPath, fileName, destPath string) {
	if basePath != "" {
		srcDb := filepath.Join(basePath, "thumbnails", fileName)
		if data, err := os.ReadFile(srcDb); err == nil && len(data) > 0 {
			_ = os.WriteFile(destPath, data, 0644)

			return
		}
	}

	if data, err := os.ReadFile(thumbPath); err == nil && len(data) > 0 {
		_ = os.WriteFile(destPath, data, 0644)
	}
}

func tryReuseExistingThumbnail(input ThumbnailInput, destPath, fileName string) bool {
	if input.Database == nil {
		return false
	}

	dbThumbPath := filepath.Join(input.Database.BasePath, "thumbnails", fileName)
	if !hasValidThumbFile(dbThumbPath) {
		return false
	}

	data, err := os.ReadFile(dbThumbPath)
	if err != nil || len(data) == 0 {
		return false
	}

	_ = os.WriteFile(destPath, data, 0644)
	input.Media.ThumbnailPath = "thumbnails/" + fileName
	fmt.Println("     🖼️  Thumbnail reused from cache")

	return true
}
