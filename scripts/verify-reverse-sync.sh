#!/usr/bin/env bash
# verify-reverse-sync.sh — Real-OS verification of v2.310.0+ reverse-sync.
#
# Captures a complete transcript at /tmp/mahin-reverse-sync-<ts>.log
# (every command, every output, every FAIL with the relevant MediaId and
# DB row context). Restores the touched row to its original state.
#
# Usage: bash scripts/verify-reverse-sync.sh /path/to/scanned/folder
#
# Exit codes:
#   0  all checks PASS         2  Test 1 (sidecar rewrite) FAIL
#   1  pre-flight failure      3  Test 2 (sidecar removal) FAIL
set -uo pipefail

TS="$(date +%Y%m%d-%H%M%S)"
LOG="/tmp/mahin-reverse-sync-${TS}.log"
exec > >(tee -a "$LOG") 2>&1   # full transcript: stdout + stderr → log
echo "🗒  Transcript: $LOG"
echo "🕒 Started: $(date)"

PASS_LIST=()  # collected for the final summary
FAIL_LIST=()

note_pass() { PASS_LIST+=("$1"); echo "✅ PASS: $1"; }
note_fail() {
  # $1 = check label   $2 = reason   $3 (optional) = MediaId
  local label="$1" reason="$2" mid="${3:-}"
  FAIL_LIST+=("$label | reason=$reason | MediaId=${mid:-N/A}")
  echo "❌ FAIL: $label"
  echo "        reason: $reason"
  [[ -n "$mid" ]] && {
    echo "        MediaId: $mid"
    echo "        --- DB row snapshot ---"
    sqlite3 -header -column "$DB_PATH" \
      "SELECT MediaId, Title, IsDeleted, UpdatedAt, CurrentFilePath
       FROM Media WHERE MediaId=$mid;" || true
    echo "        --- recent ReconciliationHistory for this row ---"
    sqlite3 -header -column "$DB_PATH" \
      "SELECT rh.OccurredAt, rat.Name, COALESCE(rh.Details,'')
       FROM ReconciliationHistory rh
       JOIN ReconciliationActionType rat
         ON rat.ReconciliationActionTypeId = rh.ReconciliationActionTypeId
       WHERE rh.MediaId=$mid
       ORDER BY rh.ReconciliationHistoryId DESC LIMIT 5;" || true
  }
}

# ---------------- Pre-flight ----------------
FOLDER="${1:-}"
[[ -z "$FOLDER" ]] && { echo "Usage: $0 <scanned-folder>"; exit 1; }
command -v mahin >/dev/null 2>&1 || { note_fail "mahin on PATH" "binary not found"; exit 1; }
[[ -d "$FOLDER" ]] || { note_fail "folder exists" "missing: $FOLDER"; exit 1; }

DB_PATH="${MOVIE_DB_PATH:-$HOME/.movie/movie.db}"
[[ -f "$DB_PATH" ]] || { note_fail "DB exists" "missing: $DB_PATH (set MOVIE_DB_PATH if custom)"; exit 1; }
echo "📦 Folder: $FOLDER"
echo "🗄  DB:     $DB_PATH"
echo "🔧 mahin:  $(mahin version 2>&1 | head -1)"

JSON_DIR="$FOLDER/.movie-output/json"
if [[ ! -d "$JSON_DIR" ]]; then
  echo "ℹ️  No prior scan; seeding..."
  mahin scan "$FOLDER"
fi

mtime() { stat -f %m "$1" 2>/dev/null || stat -c %Y "$1"; }
back_one_minute() { date -v-1M +%Y%m%d%H%M.%S 2>/dev/null || date -d '-1 minute' +%Y%m%d%H%M.%S; }

# Pick the first active row under FOLDER.
ROW="$(sqlite3 "$DB_PATH" "
  SELECT MediaId||'|'||Title||'|'||COALESCE(CurrentFilePath,'')
  FROM Media
  WHERE OriginalFilePath LIKE '${FOLDER%/}/%'
    AND COALESCE(IsDeleted,0)=0
    AND COALESCE(CurrentFilePath,'')!=''
  ORDER BY MediaId LIMIT 1;")"
[[ -z "$ROW" ]] && { note_fail "pick test row" "no active rows under $FOLDER"; exit 1; }
MEDIA_ID="${ROW%%|*}"; REST="${ROW#*|}"
ORIG_TITLE="${REST%%|*}"
MEDIA_FILE="${REST#*|}"
echo "🎯 Target: MediaId=$MEDIA_ID Title='$ORIG_TITLE'"
echo "🎬 File:   $MEDIA_FILE"

# Pre-flight: the on-disk media file MUST exist for reverse-sync to act on it.
if [[ ! -f "$MEDIA_FILE" ]]; then
  note_fail "media file on disk" "CurrentFilePath missing: $MEDIA_FILE" "$MEDIA_ID"
  exit 1
fi

SIDECAR="$(grep -lR "\"title\": \"$ORIG_TITLE\"" "$JSON_DIR" 2>/dev/null | head -1)"
if [[ -z "$SIDECAR" ]]; then
  note_fail "find sidecar" "no JSON sidecar for '$ORIG_TITLE' under $JSON_DIR" "$MEDIA_ID"
  exit 1
fi
if [[ ! -f "$SIDECAR" ]]; then
  note_fail "sidecar file on disk" "sidecar path resolved but missing: $SIDECAR" "$MEDIA_ID"
  exit 1
fi
echo "📄 Sidecar: $SIDECAR"

# ---------------- Test 1: rewrite on DB edit ----------------
echo ""; echo "==== Test 1: sidecar rewrite on DB edit ===="
BEFORE_MTIME="$(mtime "$SIDECAR")"
NEW_TITLE="${ORIG_TITLE} [verify-${TS}]"
sqlite3 "$DB_PATH" "UPDATE Media SET Title='$NEW_TITLE', UpdatedAt=datetime('now') WHERE MediaId=$MEDIA_ID;"
echo "✏️  Title → '$NEW_TITLE'"
touch -t "$(back_one_minute)" "$SIDECAR"

mahin scan "$FOLDER" --reverse-sync-only

T1_OK=true
if [[ ! -f "$MEDIA_FILE" ]]; then
  note_fail "T1 media file present" "expected on-disk file vanished: $MEDIA_FILE" "$MEDIA_ID"
  T1_OK=false
fi
if [[ ! -f "$SIDECAR" ]]; then
  note_fail "T1 sidecar present" "expected sidecar missing after rewrite: $SIDECAR" "$MEDIA_ID"
  T1_OK=false
fi

if $T1_OK; then
  AFTER_MTIME="$(mtime "$SIDECAR")"
  SIDECAR_TITLE="$(grep -o '"title": *"[^"]*"' "$SIDECAR" | head -1)"
  if (( AFTER_MTIME <= BEFORE_MTIME )); then
    note_fail "T1 mtime advanced" "before=$BEFORE_MTIME after=$AFTER_MTIME" "$MEDIA_ID"
    T1_OK=false
  fi
  if [[ "$SIDECAR_TITLE" != *"$NEW_TITLE"* ]]; then
    note_fail "T1 sidecar title updated" "expected substring '$NEW_TITLE' got: $SIDECAR_TITLE" "$MEDIA_ID"
    T1_OK=false
  fi
fi
$T1_OK && note_pass "T1 sidecar rewritten with new title (MediaId=$MEDIA_ID)"

# Restore title before Test 2.
sqlite3 "$DB_PATH" "UPDATE Media SET Title='$ORIG_TITLE', UpdatedAt=datetime('now') WHERE MediaId=$MEDIA_ID;"

# ---------------- Test 2: sidecar removal on soft-delete ----------------
echo ""; echo "==== Test 2: sidecar removal on soft-delete ===="
if [[ ! -f "$SIDECAR" ]]; then
  note_fail "T2 pre-check sidecar present" "sidecar missing before soft-delete: $SIDECAR" "$MEDIA_ID"
  exit 3
fi
sqlite3 "$DB_PATH" "UPDATE Media SET IsDeleted=1, UpdatedAt=datetime('now') WHERE MediaId=$MEDIA_ID;"
echo "🗑  Row $MEDIA_ID soft-deleted"

mahin scan "$FOLDER" --reverse-sync-only

T2_OK=true
if [[ -f "$SIDECAR" ]]; then
  note_fail "T2 sidecar removed" "sidecar still present at $SIDECAR" "$MEDIA_ID"
  T2_OK=false
else
  note_pass "T2 sidecar removed (MediaId=$MEDIA_ID)"
fi

# Restore row + recreate sidecar.
sqlite3 "$DB_PATH" "UPDATE Media SET IsDeleted=0, UpdatedAt=datetime('now') WHERE MediaId=$MEDIA_ID;"
mahin scan "$FOLDER" >/dev/null

# ---------------- Audit dump ----------------
echo ""; echo "==== Last 10 ReconciliationHistory rows ===="
sqlite3 -header -column "$DB_PATH" "
  SELECT rh.OccurredAt, rat.Name, rh.MediaId, COALESCE(rh.Details,'')
  FROM ReconciliationHistory rh
  JOIN ReconciliationActionType rat ON rat.ReconciliationActionTypeId=rh.ReconciliationActionTypeId
  ORDER BY rh.ReconciliationHistoryId DESC LIMIT 10;"

# ---------------- Final summary ----------------
echo ""; echo "==== SUMMARY ===="
echo "PASS: ${#PASS_LIST[@]}    FAIL: ${#FAIL_LIST[@]}"
for p in "${PASS_LIST[@]}"; do echo "  ✅ $p"; done
for f in "${FAIL_LIST[@]}"; do echo "  ❌ $f"; done
echo ""
echo "📄 Full transcript: $LOG"

$T1_OK && $T2_OK && exit 0
$T1_OK || exit 2
exit 3
