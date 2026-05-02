// movie_scan_reverse_write.go — DB row → JSON sidecar writer used by
// the reverse-sync pass.
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v7/apperror"
	"github.com/alimtvnetwork/movie-cli-v7/db"
)

// sidecarPathFor returns the canonical sidecar path for a Media row,
// using the same slug rules as writeMediaJSON. Falls back to a
// filename-derived slug when the row cannot be loaded.
func sidecarPathFor(database *db.DB, jsonRoot string, r *db.ReverseSyncRow) string {
	subDir := db.JsonSubDir(r.Type)
	media, err := database.GetMediaByID(r.ID)
	if err != nil || media == nil {
		base := filepath.Base(r.CurrentFilePath)
		base = strings.TrimSuffix(base, filepath.Ext(base))
		return filepath.Join(jsonRoot, subDir, base+".json")
	}
	return filepath.Join(jsonRoot, subDir, mediaSlug(media)+".json")
}

// parseDBTime accepts both `datetime('now')` (UTC, no TZ) and RFC3339.
func parseDBTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02 15:04:05", s)
}

// writeSidecarFromDB rewrites the sidecar for the given row using the
// existing writeMediaJSON helper (single source of truth for layout).
func writeSidecarFromDB(database *db.DB, jsonRoot string, r *db.ReverseSyncRow) error {
	if scanDryRun {
		return nil
	}
	media, err := database.GetMediaByID(r.ID)
	if err != nil {
		return apperror.Wrap("load media", err)
	}
	if media == nil {
		return apperror.New("reverse-sync: media row vanished")
	}
	basePath := filepath.Dir(jsonRoot) // .movie-output
	return writeMediaJSON(basePath, media)
}
