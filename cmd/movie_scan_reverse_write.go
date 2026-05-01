// movie_scan_reverse_write.go — DB row → JSON sidecar writer used by
// the reverse-sync pass. Builds a Media struct from a ReverseSyncRow
// (which carries only the IDs we need) by re-fetching the full row.
package cmd

import (
	"os"
	"path/filepath"
	"time"

	"github.com/alimtvnetwork/movie-cli-v7/apperror"
	"github.com/alimtvnetwork/movie-cli-v7/db"
)

// sidecarPathFor returns the canonical sidecar path for a Media row,
// using the same slug rules as writeMediaJSON.
func sidecarPathFor(jsonRoot string, r *db.ReverseSyncRow) string {
	subDir := db.JsonSubDir(r.Type)
	stub := &db.Media{ID: r.ID, Type: r.Type}
	media, err := loadMediaForSidecar(stub.ID)
	if err != nil || media == nil {
		// Fallback: derive slug from current file path when DB load fails.
		base := filepath.Base(r.CurrentFilePath)
		base = base[:len(base)-len(filepath.Ext(base))]
		return filepath.Join(jsonRoot, subDir, base+".json")
	}
	return filepath.Join(jsonRoot, subDir, mediaSlug(media)+".json")
}

// shouldRewriteSidecar returns true when the sidecar is missing or its
// mtime is older than the DB UpdatedAt timestamp.
func shouldRewriteSidecar(sidecarPath, dbUpdatedAt string) bool {
	info, err := os.Stat(sidecarPath)
	if err != nil {
		return true // missing → rewrite
	}
	dbTime, parseErr := parseDBTime(dbUpdatedAt)
	if parseErr != nil {
		return true // unparseable → safer to rewrite
	}
	return info.ModTime().Before(dbTime)
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
	media, err := loadMediaForSidecar(r.ID)
	if err != nil {
		return apperror.Wrap("load media", err)
	}
	if media == nil {
		return apperror.New("reverse-sync: media row vanished")
	}
	basePath := filepath.Dir(jsonRoot) // .movie-output
	return writeMediaJSON(basePath, media)
}

// loadMediaForSidecar fetches one Media row by ID. Returns (nil, nil)
// when not found so callers can branch on absence.
func loadMediaForSidecar(id int64) (*db.Media, error) {
	media, err := lookupMediaByID(id)
	if err != nil {
		return nil, err
	}
	return media, nil
}
