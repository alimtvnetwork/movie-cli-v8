// seed.go — seed data for FileAction and default Config.
package db

import (
	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

// seedFileActions inserts the 15 predefined FileAction types.
// Order MUST match the FileActionType enum in action_history.go (1-indexed).
func (d *DB) seedFileActions() error {
	actions := []string{
		"Move", "Rename", "Delete", "Popout", "Restore",
		"ScanAdd", "ScanRemove", "RescanUpdate",
		"TagAdd", "TagRemove",
		"WatchlistAdd", "WatchlistRemove", "WatchlistStatusChange",
		"ConfigChange",
		"Compact", // popout cleanup: subfolder → <root>/.temp/
	}
	for _, name := range actions {
		if _, err := d.Exec("INSERT OR IGNORE INTO FileAction (Name) VALUES (?)", name); err != nil {
			return apperror.Wrapf(err, "seed FileAction %q", name)
		}
	}
	return nil
}

// seedDefaultConfig inserts default config values if not already present.
func (d *DB) seedDefaultConfig() error {
	defaults := [][2]string{
		{"MoviesDir", "~/Movies"},
		{"TvDir", "~/TVShows"},
		{"ArchiveDir", "~/Archive"},
		{"ScanDir", "~/Downloads"},
		{"PageSize", "20"},
	}
	for _, kv := range defaults {
		if _, err := d.Exec("INSERT OR IGNORE INTO Config (ConfigKey, ConfigValue) VALUES (?, ?)", kv[0], kv[1]); err != nil {
			return apperror.Wrapf(err, "seed config %q", kv[0])
		}
	}
	return nil
}
