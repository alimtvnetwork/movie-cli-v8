# Context Menu — Real-OS QA Checklist

Manual verification steps for `movie add-contextmenu` /
`remove-contextmenu` / `contextmenu-status` on the three supported
platforms. Run each section on a clean machine (no prior install) and
record pass/fail next to every item.

Applies to: `mahin` binary, version ≥ `v2.296.0`.

## A. Common preconditions

- [ ] Binary `mahin` is on `PATH` and `mahin version` prints the expected
      semver.
- [ ] `mahin contextmenu-status` reports **❌ Not installed** before the
      test starts.
- [ ] No leftover entries from a previous install
      (`~/Library/Services/Movie - *.workflow` on macOS,
      `~/.local/share/file-manager/actions/movie-*.desktop` on Linux,
      `HKCU\Software\Classes\Directory\shell\Movie*` on Windows).

## B. macOS (Finder Quick Actions)

### B1. Install
- [ ] Run `mahin add-contextmenu` — exits 0.
- [ ] Stdout shows the post-install hint mentioning **System Settings →
      Keyboard → Keyboard Shortcuts → Services**.
- [ ] Stdout also mentions the **typed 'y' confirmation** for
      destructive actions.
- [ ] `~/Library/Services/` now contains 4 `Movie - *.workflow`
      bundles, each with `Contents/Info.plist` and
      `Contents/document.wflow`.
- [ ] `mahin contextmenu-status` reports **✅ Installed** with
      `all 4 workflows present`.

### B2. Enable in System Settings
- [ ] Open **System Settings → Keyboard → Keyboard Shortcuts →
      Services → Files & Folders**.
- [ ] All four `Movie - …` entries are visible and toggleable.
- [ ] Enable all four.

### B3. Non-destructive actions (Scan / Report / Stats)
For each of `Movie - Scan with Movie`, `Movie - Open Movie Report`,
`Movie - Show Movie Stats`:
- [ ] Right-click a real folder in Finder → submenu shows the entry.
- [ ] Click → a `Terminal.app` window opens cd'd into the clicked
      folder.
- [ ] The corresponding `mahin` command runs **without** any extra
      prompt.
- [ ] `ActionHistory` table contains a new row with
      `Detail` containing `trigger=contextmenu;entry=<key>;cwd=<path>`.

### B4. Destructive action (Rescan) — confirmation gate
- [ ] Right-click a folder → choose `Movie - Rescan with Movie`.
- [ ] A `Terminal.app` window opens.
- [ ] The window displays the banner
      `⚠️  About to run: movie rescan in <folder>`.
- [ ] Prompt `Type y then Enter to continue (anything else cancels):`
      is shown and the process **blocks**.
- [ ] Type `n` + Enter → output prints `cancelled`, exit 0, no DB
      mutation occurred (verify via `mahin history reconcile` row count
      unchanged).
- [ ] Repeat, type `y` + Enter → `mahin rescan .` runs to completion.
- [ ] A new `ReconciliationHistory` row exists for that scanDir.

### B5. Uninstall
- [ ] Run `mahin remove-contextmenu` — exits 0.
- [ ] All four `Movie - *.workflow` bundles are removed from
      `~/Library/Services/`.
- [ ] `mahin contextmenu-status` reports **❌ Not installed**.
- [ ] Finder right-click no longer shows the `Movie` submenu (may
      require Finder relaunch: `killall Finder`).

## C. Linux (GNOME / KDE `.desktop` actions)

- [ ] `mahin add-contextmenu` → exits 0, creates 4 `.desktop` files
      under `~/.local/share/file-manager/actions/`.
- [ ] Nautilus / Dolphin right-click on a folder shows the four
      `Movie …` entries.
- [ ] Clicking a non-destructive entry opens the system terminal and
      runs the command.
- [ ] Clicking `Rescan` shows no native confirm dialog (Linux flow
      does not use the macOS prompt — destructive guard is OS-only).
- [ ] `mahin remove-contextmenu` removes all four `.desktop` files.
- [ ] `mahin contextmenu-status` reflects each state correctly.

## D. Windows (registry)

- [ ] Run `mahin add-contextmenu` from an **elevated** prompt — exits 0.
- [ ] `HKCU\Software\Classes\Directory\shell\Movie\shell\` contains
      4 subkeys.
- [ ] Right-click a folder in Explorer → `Movie ▸` submenu shows 4
      entries.
- [ ] Each entry opens a new `cmd.exe` window in the clicked folder
      and runs the matching `mahin` command.
- [ ] `mahin remove-contextmenu` purges all `Movie*` keys.
- [ ] `mahin contextmenu-status` reflects each state correctly.

## E. Telemetry sanity check (all OSes)

- [ ] Open SQLite DB and run
      `SELECT COUNT(*) FROM ActionHistory WHERE Detail LIKE 'trigger=contextmenu%';`.
- [ ] Count increases by exactly 1 per click.
- [ ] `entry=` value matches the clicked submenu key.

## F. Sign-off

- Tester: ____________________
- Date (UTC+8): ____________________
- mahin version: ____________________
- macOS / Linux distro / Windows build: ____________________
- All sections pass: ☐ yes ☐ no  (if no, file an issue per failure)
