#!/usr/bin/env bash
# qa-context-menu-macos.sh — Guided harness for
# spec/08-app/11-context-menu-qa-checklist.md sections A, B, E, F.
#
# Automates everything that does NOT require a human eye on Finder.app.
# For each visual/manual step the script prints the action, waits for
# the user to type y / n / s (skip), and writes a timestamped report to
# /tmp/mahin-contextmenu-qa-<date>.md.
#
# macOS only. Run from any folder.
set -uo pipefail

[[ "$(uname)" == "Darwin" ]] || { echo "macOS only"; exit 1; }
command -v mahin >/dev/null || { echo "mahin not on PATH"; exit 1; }

REPORT="/tmp/mahin-contextmenu-qa-$(date +%Y%m%d-%H%M%S).md"
PASS=0; FAIL=0; SKIP=0
SVC_DIR="$HOME/Library/Services"
DB_PATH="${MOVIE_DB_PATH:-$HOME/.movie/movie.db}"
TEST_DIR="${1:-$HOME/Movies}"

log() { echo "$*" | tee -a "$REPORT"; }
section() { log ""; log "## $*"; log ""; }

ask() {
  # ask "description" "auto-result-or-empty"
  local desc="$1" auto="${2:-}"
  if [[ -n "$auto" ]]; then
    log "- [x] $desc — **auto: $auto**"
    PASS=$((PASS+1)); return
  fi
  read -r -p "  $desc  [y/n/s]: " ans </dev/tty
  case "$ans" in
    y|Y) log "- [x] $desc — PASS"; PASS=$((PASS+1));;
    n|N) read -r -p "    reason: " reason </dev/tty
         log "- [ ] $desc — **FAIL**: $reason"; FAIL=$((FAIL+1));;
    *)   log "- [~] $desc — SKIP"; SKIP=$((SKIP+1));;
  esac
}

cat > "$REPORT" <<EOF
# Context-Menu QA Report
- Date: $(date)
- mahin: $(mahin version 2>&1 | head -1)
- macOS: $(sw_vers -productVersion)
- Test folder: $TEST_DIR
EOF

section "A. Common preconditions"
ver="$(mahin version 2>&1 | head -1)"
ask "mahin on PATH and version printed: $ver" "$ver"
status_pre="$(mahin contextmenu-status 2>&1 || true)"
log "  status: \`$status_pre\`"
if echo "$status_pre" | grep -qi "not installed"; then
  ask "contextmenu-status reports Not installed" "yes"
else
  log "  ⚠️  Already installed — uninstalling for clean run"
  mahin remove-contextmenu >/dev/null 2>&1 || true
  ask "contextmenu-status reports Not installed (after cleanup)" "yes"
fi
leftovers="$(ls "$SVC_DIR"/Movie\ -\ *.workflow 2>/dev/null | wc -l | tr -d ' ')"
ask "No leftover Movie workflows in ~/Library/Services" \
  "$([[ $leftovers -eq 0 ]] && echo yes || echo "FAIL: $leftovers leftover")"

section "B1. Install"
install_out="$(mahin add-contextmenu 2>&1)"; rc=$?
log "\`\`\`"; log "$install_out"; log "\`\`\`"
ask "add-contextmenu exits 0" "$([[ $rc -eq 0 ]] && echo yes || echo "FAIL: rc=$rc")"
ask "Stdout mentions System Settings → Keyboard → Services" \
  "$(echo "$install_out" | grep -qi "System Settings.*Keyboard.*Services" && echo yes || echo "FAIL: hint missing")"
ask "Stdout mentions typed 'y' confirmation" \
  "$(echo "$install_out" | grep -qi "typed 'y'" && echo yes || echo "FAIL: hint missing")"
wf_count="$(ls "$SVC_DIR"/Movie\ -\ *.workflow 2>/dev/null | wc -l | tr -d ' ')"
ask "4 Movie - *.workflow bundles in ~/Library/Services" \
  "$([[ $wf_count -eq 4 ]] && echo yes || echo "FAIL: $wf_count bundles")"
status_post="$(mahin contextmenu-status 2>&1 || true)"
log "  status: \`$status_post\`"
ask "contextmenu-status reports Installed (4 workflows)" \
  "$(echo "$status_post" | grep -qi "all 4 workflows present" && echo yes || echo "FAIL")"

section "B2. Enable in System Settings (manual)"
log "👉 Open: System Settings → Keyboard → Keyboard Shortcuts → Services → Files & Folders"
open "x-apple.systempreferences:com.apple.preference.keyboard?Services" 2>/dev/null || true
ask "All four 'Movie - …' entries are visible and toggleable"
ask "All four are enabled"

section "B3. Non-destructive actions"
log "👉 In Finder, right-click \`$TEST_DIR\` and trigger each action below."
mkdir -p "$TEST_DIR"
for entry in "Movie - Scan with Movie" "Movie - Open Movie Report" "Movie - Show Movie Stats"; do
  ask "$entry → Terminal opens cd'd into folder, runs without prompt"
done

section "B4. Destructive action (Rescan) — confirmation gate"
ask "Right-click → 'Movie - Rescan with Movie' opens Terminal"
ask "Banner '⚠️  About to run: mahin rescan in <folder>' is shown"
ask "Prompt 'Type y then Enter to continue ...' appears and BLOCKS"
ask "Typing 'n' + Enter prints 'cancelled' and exits 0"
ask "Typing 'y' + Enter runs mahin rescan to completion"

section "B5. Uninstall"
uninst="$(mahin remove-contextmenu 2>&1)"; rc=$?
log "\`\`\`"; log "$uninst"; log "\`\`\`"
ask "remove-contextmenu exits 0" "$([[ $rc -eq 0 ]] && echo yes || echo "FAIL: rc=$rc")"
remain="$(ls "$SVC_DIR"/Movie\ -\ *.workflow 2>/dev/null | wc -l | tr -d ' ')"
ask "All 4 workflow bundles removed" \
  "$([[ $remain -eq 0 ]] && echo yes || echo "FAIL: $remain remain")"
status_final="$(mahin contextmenu-status 2>&1 || true)"
ask "contextmenu-status reports Not installed" \
  "$(echo "$status_final" | grep -qi "not installed" && echo yes || echo "FAIL")"
killall Finder 2>/dev/null || true
ask "Finder no longer shows Movie submenu (after relaunch)"

section "E. Telemetry sanity check"
if [[ -f "$DB_PATH" ]]; then
  cnt="$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ActionHistory WHERE Detail LIKE 'trigger=contextmenu%';")"
  log "  ActionHistory contextmenu rows: $cnt"
  ask "Count increased by 1 per click during B3/B4"
  log "  Last 5 entries:"
  sqlite3 "$DB_PATH" "SELECT OccurredAt, Detail FROM ActionHistory WHERE Detail LIKE 'trigger=contextmenu%' ORDER BY ActionHistoryId DESC LIMIT 5;" \
    | sed 's/^/    /' | tee -a "$REPORT"
else
  log "  ⚠️  DB not found at $DB_PATH — telemetry check skipped"
  SKIP=$((SKIP+1))
fi

section "F. Sign-off"
read -r -p "Tester name: " tester </dev/tty
log "- Tester: $tester"
log "- Date (UTC+8): $(TZ=Asia/Kuala_Lumpur date '+%Y-%m-%d %H:%M:%S %Z')"
log "- mahin version: $(mahin version 2>&1 | head -1)"
log "- macOS: $(sw_vers -productVersion)"
log ""
log "**Summary**: PASS=$PASS  FAIL=$FAIL  SKIP=$SKIP"

echo ""
echo "📄 Report written to: $REPORT"
[[ $FAIL -eq 0 ]] && exit 0 || exit 1
