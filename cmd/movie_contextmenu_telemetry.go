// movie_contextmenu_telemetry.go — records context-menu clicks to
// ActionHistory so we can answer "how often does the user invoke Movie
// from the right-click menu?".
//
// Detection: when the OS-installed shortcut runs `movie scan`, it sets the
// env var MOVIE_TRIGGER=contextmenu (and MOVIE_CONTEXTMENU_ENTRY=<key>).
// recordContextMenuClick is called from the relevant command's entry
// point and writes a single ActionHistory row reusing existing FileAction
// enums (no schema migration required).
package cmd

import (
	"fmt"
	"os"

	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
)

// EnvContextMenuTrigger is set to "contextmenu" by the installed shortcut.
const EnvContextMenuTrigger = "MOVIE_TRIGGER"

// EnvContextMenuEntry tells us which submenu item was clicked.
const EnvContextMenuEntry = "MOVIE_CONTEXTMENU_ENTRY"

// triggerValueContextMenu is the single accepted trigger value.
const triggerValueContextMenu = "contextmenu"

// IsContextMenuInvocation returns true when the current process was started
// from the OS right-click menu shortcut.
func IsContextMenuInvocation() bool {
	return os.Getenv(EnvContextMenuTrigger) == triggerValueContextMenu
}

// RecordContextMenuClick logs the click into ActionHistory, if any.
// It silently no-ops when the env var is absent, so it is safe to call
// unconditionally from command entry points.
func RecordContextMenuClick(database *db.DB, workDir string) {
	if !IsContextMenuInvocation() {
		return
	}
	entry := os.Getenv(EnvContextMenuEntry)
	detail := fmt.Sprintf("trigger=contextmenu;entry=%s;cwd=%s", entry, workDir)
	_, err := database.InsertActionSimple(db.ActionSimpleInput{
		FileAction: db.FileActionScanAdd,
		MediaID:    nullInt64Zero(),
		Detail:     detail,
		BatchID:    "",
	})
	if err != nil {
		errlog.Warn("Could not log context-menu click: %v", err)
	}
}

func nullInt64Zero() sql.NullInt64 {
	return sql.NullInt64{Valid: false}
}
