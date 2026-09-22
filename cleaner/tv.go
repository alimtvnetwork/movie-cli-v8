// Package cleaner provides filename cleaning and metadata extraction.
package cleaner

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// TVEpisodeInfo contains structured series, season, and episode details.
type TVEpisodeInfo struct {
	ShowTitle     string
	CleanTitle    string
	SeasonNumber  int
	EpisodeNumber int
	IsTV          bool
}

var (
	// seasonEpRegex matches S01E02, S1E5, S01.E02, 1x02, Season 1 Episode 2
	seasonEpRegex = regexp.MustCompile(`(?i)(?:s|season\s*)(\d{1,2})[.\s-]*(?:e|episode\s*|x)(\d{1,3})`)
	// seasonFolderRegex matches Season 1, Season 02, S01, S2 in directory names
	seasonFolderRegex = regexp.MustCompile(`(?i)(?:^|[.\s_-])(?:season|s)\s*(\d{1,2})(?:[.\s_-]|$)`)
)

// ParseTVEpisode analyzes a file path to extract series title, season number, and episode number.
func ParseTVEpisode(filePath string) TVEpisodeInfo {
	base := filepath.Base(filePath)
	parentDir := filepath.Base(filepath.Dir(filePath))

	info := TVEpisodeInfo{}

	matches := seasonEpRegex.FindStringSubmatch(base)
	if len(matches) == 3 {
		sNum, _ := strconv.Atoi(matches[1])
		eNum, _ := strconv.Atoi(matches[2])
		info.SeasonNumber = sNum
		info.EpisodeNumber = eNum
		info.IsTV = true

		loc := seasonEpRegex.FindStringIndex(base)
		if len(loc) == 2 && loc[0] > 0 {
			rawShow := base[:loc[0]]
			cleaned := Clean(rawShow)
			info.ShowTitle = cleaned.CleanTitle
		}
	}

	if info.ShowTitle == "" || len(strings.TrimSpace(info.ShowTitle)) < 2 {
		cleanedDir := Clean(parentDir)
		info.ShowTitle = cleanedDir.CleanTitle

		if info.SeasonNumber == 0 {
			folderMatches := seasonFolderRegex.FindStringSubmatch(parentDir)
			if len(folderMatches) == 2 {
				sNum, _ := strconv.Atoi(folderMatches[1])
				info.SeasonNumber = sNum
				info.IsTV = true
			}
		}
	}

	if info.IsTV {
		if info.SeasonNumber == 0 {
			info.SeasonNumber = 1
		}

		info.CleanTitle = info.ShowTitle
	}

	return info
}
