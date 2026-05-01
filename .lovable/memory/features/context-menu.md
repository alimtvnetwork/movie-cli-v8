---
name: Context menu integration
description: `movie add-contextmenu` installs Windows/Linux/macOS shell submenu (Scan/Rescan/Report/Stats); clicks logged to ActionHistory via MOVIE_TRIGGER env var
type: feature
---

# Context Menu Integration

## Commands
- `movie add-contextmenu` — idempotent install
- `movie remove-contextmenu` — uninstall
- `movie contextmenu-status` — report installed state

## Submenu Entries
Defined in `cmd/movie_contextmenu.go` as `contextMenuEntries`:
1. Scan with Movie    → `movie scan .`
2. Rescan with Movie  → `movie rescan .`
3. Open Movie Report  → `movie scan .` (regenerates + auto-opens report.html)
4. Show Movie Stats   → `movie stats .`

## OS Implementations (build-tagged sibling files)
- **Windows** (`movie_contextmenu_windows.go`): `HKCU\Software\Classes\Directory\shell\Movie` + `Directory\Background\shell\Movie`. `cmd.exe /c "set MOVIE_TRIGGER=contextmenu&& cd /d %V && movie scan . & if errorlevel 1 pause"`. No admin required.
- **Linux** (`movie_contextmenu_linux.go`): `~/.local/share/file-manager/actions/movie-cli.desktop` (XDG-respecting). One `[X-Action-Profile]` per entry, `MimeTypes=inode/directory`.
- **macOS** (`movie_contextmenu_darwin.go`): one `~/Library/Services/Movie - <key>.workflow` Automator Quick Action per entry. **One-time manual step required:** System Settings → Keyboard → Services → Files & Folders → enable each entry. Cannot be auto-enabled without code-signing (TCC).
- **Other OS**: returns "not supported" error.

## Telemetry
- Installed shortcuts set `MOVIE_TRIGGER=contextmenu` and `MOVIE_CONTEXTMENU_ENTRY=<key>`.
- `RecordContextMenuClick(db, cwd)` in `movie_contextmenu_telemetry.go` writes one ActionHistory row per click.
- Reuses existing `FileActionScanAdd` enum (no migration). Detail format: `trigger=contextmenu;entry=<key>;cwd=<path>`.
- Currently wired into `movie scan` (covers Scan + Report). Wire into `movie rescan` and `movie stats` if those need separate counts.

## Terminal Behavior
- Windows: closes on success, `pause` on non-zero exit.
- Linux: closes on success, `read` prompt on failure.
- macOS: opens `Terminal.app` via `osascript`; user closes manually.
