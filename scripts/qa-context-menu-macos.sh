#!/usr/bin/env bash
# qa-context-menu-macos.sh — Guided harness for
# spec/08-app/11-context-menu-qa-checklist.md sections A, B, E, F.
#
# Captures a full transcript at /tmp/movie-contextmenu-qa-<ts>.log
# AND a markdown report at /tmp/movie-contextmenu-qa-<ts>.md.
# Every FAIL records: section, item, user-supplied reason, the relevant
# DB rows (ActionHistoryId / ReconciliationHistoryId) and on-disk paths.
set -uo pipefail

[[ "$(uname)" == "Darwin" ]] || { echo "macOS only"; exit 1; }
command -v movie >/dev/null || { echo "movie not on PATH"; exit 1; }

TS="$(date +%Y%m%d-%H%M%S)"
LOG="/tmp/movie-contextmenu-qa-${TS}.log"
REPORT="/tmp/movie-contextmenu-qa-${TS}.md"
exec > >(tee -a "$LOG") 2>&1
echo "🗒  Transcript: $LOG"
echo "🗒  Markdown:   $REPORT"

PASS=0; FAIL=0; SKIP=0
SVC_DIR="$HOME/Library/Services"
DB_PATH="${MOVIE_DB_PATH:-$HOME/.movie/movie.db}"
TEST_DIR="${1:-$HOME/Movies}"
CURRENT_SECTION=""

md() { echo "$*" >> "$REPORT"; }   # write-only-to-markdown
both() { echo "$*"; md "$*"; }      # console + markdown
section() { CURRENT_SECTION="$1"; both ""; both "## $1"; both ""; }

# Snapshot the latest ActionHistory + ReconciliationHistory IDs so we
# can attribute new rows that appear during a click to that step.
last_action_id() {
  sqlite3 "$DB_PATH" "SELECT COALESCE(MAX(ActionHistoryId),0) FROM ActionHistory;" 2>/dev/null || echo 0
}
last_recon_id() {
  sqlite3 "$DB_PATH" "SELECT COALESCE(MAX(ReconciliationHistoryId),0) FROM ReconciliationHistory;" 2>/dev/null || echo 0
}

dump_new_action_rows() {
  local since="$1"
  [[ ! -f "$DB_PATH" ]] && return
  echo "        --- new ActionHistory rows since #$since ---"
  sqlite3 -header -column "$DB_PATH" \
    "SELECT ActionHistoryId, OccurredAt, Detail
     FROM ActionHistory
     WHERE ActionHistoryId > $since
     ORDER BY ActionHistoryId DESC LIMIT 5;" || true
}

note_fail() {
  # $1 = item label   $2 = reason   $3 (optional) = since-action-id
  local item="$1" reason="$2" since="${3:-}"
  FAIL=$((FAIL+1))
  echo "❌ FAIL [$CURRENT_SECTION] $item"
  echo "        reason: $reason"
  md "- [ ] $item — **FAIL**: $reason"
  [[ -n "$since" ]] && dump_new_action_rows "$since"
}

note_pass() {
  PASS=$((PASS+1))
  echo "✅ PASS [$CURRENT_SECTION] $1"
  md "- [x] $1 — PASS"
}

note_skip() {
  SKIP=$((SKIP+1))
  echo "⏭  SKIP [$CURRENT_SECTION] $1"
  md "- [~] $1 — SKIP"
}

# ask "label" "auto-result-or-empty" [since-action-id]
ask() {
  local label="$1" auto="${2:-}" since="${3:-}"
  if [[ "$auto" == "yes" ]]; then note_pass "$label"; return; fi
  if [[ "$auto" == FAIL:* ]]; then note_fail "$label" "${auto#FAIL:}" "$since"; return; fi
  read -r -p "  $label  [y/n/s]: " ans </dev/tty
  case "$ans" in
    y|Y) note_pass "$label" ;;
    n|N) read -r -p "    reason: " reason </dev/tty
         note_fail "$label" "$reason" "$since" ;;
    *)   note_skip "$label" ;;
  esac
}

# ---------------- header ----------------
cat > "$REPORT" <<EOF
# Context-Menu QA Report
- Date: $(date)
- movie: $(movie version 2>&1 | head -1)
- macOS: $(sw_vers -productVersion)
- Test folder: $TEST_DIR
- Transcript: $LOG
EOF

# ---------------- A. preconditions ----------------
section "A. Common preconditions"
ver="$(movie version 2>&1 | head -1)"
ask "movie on PATH and version printed: $ver" "yes"
status_pre="$(movie contextmenu-status 2>&1 || true)"
echo "  status: $status_pre"
if echo "$status_pre" | grep -qi "not installed"; then
  ask "contextmenu-status reports Not installed" "yes"
else
  echo "  ⚠️  Already installed — uninstalling for clean run"
  movie remove-contextmenu >/dev/null 2>&1 || true
  ask "contextmenu-status reports Not installed (after cleanup)" "yes"
fi
leftovers="$(ls "$SVC_DIR"/Movie\ -\ *.workflow 2>/dev/null | wc -l | tr -d ' ')"
if [[ $leftovers -eq 0 ]]; then
  ask "No leftover Movie workflows in ~/Library/Services" "yes"
else
  ask "No leftover Movie workflows in ~/Library/Services" "FAIL:$leftovers leftover bundle(s) at $SVC_DIR"
fi

# ---------------- B1. Install ----------------
section "B1. Install"
install_out="$(movie add-contextmenu 2>&1)"; rc=$?
echo "$install_out"
md '```'; md "$install_out"; md '```'
if [[ $rc -eq 0 ]]; then ask "add-contextmenu exits 0" "yes"
else ask "add-contextmenu exits 0" "FAIL:exit code $rc"; fi
if echo "$install_out" | grep -qi "System Settings.*Keyboard.*Services"; then
  ask "Stdout mentions System Settings → Keyboard → Services" "yes"
else
  ask "Stdout mentions System Settings → Keyboard → Services" "FAIL:hint string missing from add-contextmenu stdout"
fi
if echo "$install_out" | grep -qi "typed 'y'"; then
  ask "Stdout mentions typed 'y' confirmation" "yes"
else
  ask "Stdout mentions typed 'y' confirmation" "FAIL:hint string missing from add-contextmenu stdout"
fi
wf_count="$(ls "$SVC_DIR"/Movie\ -\ *.workflow 2>/dev/null | wc -l | tr -d ' ')"
if [[ $wf_count -eq 4 ]]; then
  ask "4 Movie - *.workflow bundles in ~/Library/Services" "yes"
else
  ask "4 Movie - *.workflow bundles in ~/Library/Services" "FAIL:found $wf_count bundles, expected 4 — ls $SVC_DIR/Movie*.workflow"
fi
status_post="$(movie contextmenu-status 2>&1 || true)"
echo "  status: $status_post"
if echo "$status_post" | grep -qi "all 4 workflows present"; then
  ask "contextmenu-status reports Installed (4 workflows)" "yes"
else
  ask "contextmenu-status reports Installed (4 workflows)" "FAIL:status said: $status_post"
fi

# ---------------- B2 (manual) ----------------
section "B2. Enable in System Settings (manual)"
both "👉 System Settings → Keyboard → Keyboard Shortcuts → Services → Files & Folders"
open "x-apple.systempreferences:com.apple.preference.keyboard?Services" 2>/dev/null || true
ask "All four 'Movie - …' entries are visible and toggleable"
ask "All four are enabled"

# ---------------- B3. Non-destructive ----------------
section "B3. Non-destructive actions"
mkdir -p "$TEST_DIR"
both "👉 In Finder, right-click \`$TEST_DIR\` and trigger each action below."
for entry in "Movie - Scan with Movie" "Movie - Open Movie Report" "Movie - Show Movie Stats"; do
  since=$(last_action_id)
  read -r -p "  Trigger '$entry' now, then press Enter " </dev/tty
  ask "$entry → Terminal opens cd'd into folder, runs without prompt" "" "$since"
done

# ---------------- B4. Destructive (Rescan) ----------------
section "B4. Destructive action (Rescan) — confirmation gate"
since=$(last_recon_id)
ask "Right-click → 'Movie - Rescan with Movie' opens Terminal" "" "$(last_action_id)"
ask "Banner '⚠️  About to run: movie rescan in <folder>' is shown"
ask "Prompt 'Type y then Enter to continue ...' appears and BLOCKS"
ask "Typing 'n' + Enter prints 'cancelled' and exits 0"
ask "Typing 'y' + Enter runs movie rescan to completion"
new_recon=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ReconciliationHistory WHERE ReconciliationHistoryId > $since;" 2>/dev/null || echo 0)
both "  ReconciliationHistory rows added during B4: $new_recon"

# ---------------- B5. Uninstall ----------------
section "B5. Uninstall"
uninst="$(movie remove-contextmenu 2>&1)"; rc=$?
echo "$uninst"; md '```'; md "$uninst"; md '```'
if [[ $rc -eq 0 ]]; then ask "remove-contextmenu exits 0" "yes"
else ask "remove-contextmenu exits 0" "FAIL:exit code $rc"; fi
remain="$(ls "$SVC_DIR"/Movie\ -\ *.workflow 2>/dev/null | wc -l | tr -d ' ')"
if [[ $remain -eq 0 ]]; then ask "All 4 workflow bundles removed" "yes"
else ask "All 4 workflow bundles removed" "FAIL:$remain bundle(s) remain — ls $SVC_DIR/Movie*.workflow"; fi
status_final="$(movie contextmenu-status 2>&1 || true)"
if echo "$status_final" | grep -qi "not installed"; then
  ask "contextmenu-status reports Not installed" "yes"
else
  ask "contextmenu-status reports Not installed" "FAIL:status said: $status_final"
fi
killall Finder 2>/dev/null || true
ask "Finder no longer shows Movie submenu (after relaunch)"

# ---------------- E. Telemetry ----------------
section "E. Telemetry sanity check"
if [[ -f "$DB_PATH" ]]; then
  cnt="$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ActionHistory WHERE Detail LIKE 'trigger=contextmenu%';")"
  both "  ActionHistory contextmenu rows: $cnt"
  ask "Count increased by 1 per click during B3/B4"
  both "  Last 5 contextmenu telemetry entries:"
  sqlite3 -header -column "$DB_PATH" \
    "SELECT ActionHistoryId, OccurredAt, Detail
     FROM ActionHistory
     WHERE Detail LIKE 'trigger=contextmenu%'
     ORDER BY ActionHistoryId DESC LIMIT 5;" | tee -a "$REPORT"
else
  both "  ⚠️  DB not found at $DB_PATH — telemetry check skipped"
  SKIP=$((SKIP+1))
fi

# ---------------- F. Sign-off ----------------
section "F. Sign-off"
read -r -p "Tester name: " tester </dev/tty
both "- Tester: $tester"
both "- Date (UTC+8): $(TZ=Asia/Kuala_Lumpur date '+%Y-%m-%d %H:%M:%S %Z')"
both "- movie version: $(movie version 2>&1 | head -1)"
both "- macOS: $(sw_vers -productVersion)"

both ""
both "**Summary**: PASS=$PASS  FAIL=$FAIL  SKIP=$SKIP"

echo ""
echo "📄 Markdown report: $REPORT"
echo "📄 Full transcript: $LOG"
[[ $FAIL -eq 0 ]] && exit 0 || exit 1
