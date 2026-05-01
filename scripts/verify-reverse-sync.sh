#!/usr/bin/env bash
# verify-reverse-sync.sh — Real-OS verification of v2.310.0 reverse-sync.
#
# Run on the user's machine where `mahin` is installed and a test folder
# with at least one already-scanned video file exists.
#
# Usage: bash scripts/verify-reverse-sync.sh /path/to/test-folder
#
# Exit codes:
#   0  all checks PASS
#   1  pre-flight failure (binary/folder/sidecar missing)
#   2  Test 1 (sidecar rewrite) FAIL
#   3  Test 2 (sidecar removal on soft-delete) FAIL
set -uo pipefail

FOLDER="${1:-}"
if [[ -z "$FOLDER" ]]; then
  echo "Usage: $0 <scanned-folder>"
  exit 1
fi
if ! command -v mahin >/dev/null 2>&1; then
  echo "❌ mahin binary not on PATH"
  exit 1
fi
if [[ ! -d "$FOLDER" ]]; then
  echo "❌ folder not found: $FOLDER"
  exit 1
fi

JSON_DIR="$FOLDER/.movie-output/json"
if [[ ! -d "$JSON_DIR" ]]; then
  echo "ℹ️  No prior scan found; running initial scan to seed sidecars..."
  mahin scan "$FOLDER" >/dev/null
fi

# Resolve DB path (mahin keeps it under ~/.movie or env override).
DB_PATH="${MOVIE_DB_PATH:-$HOME/.movie/movie.db}"
if [[ ! -f "$DB_PATH" ]]; then
  echo "❌ SQLite DB not found at $DB_PATH (set MOVIE_DB_PATH if custom)"
  exit 1
fi

mtime() { stat -f %m "$1" 2>/dev/null || stat -c %Y "$1"; }

pick_target_row() {
  sqlite3 "$DB_PATH" "
    SELECT MediaId, Title, CurrentFilePath
    FROM Media
    WHERE OriginalFilePath LIKE '${FOLDER%/}/%'
      AND COALESCE(IsDeleted,0)=0
      AND COALESCE(CurrentFilePath,'')!=''
    ORDER BY MediaId LIMIT 1;"
}

ROW="$(pick_target_row)"
if [[ -z "$ROW" ]]; then
  echo "❌ No active Media rows under $FOLDER — cannot test"
  exit 1
fi
MEDIA_ID="$(echo "$ROW" | cut -d'|' -f1)"
ORIG_TITLE="$(echo "$ROW" | cut -d'|' -f2)"
echo "🎯 Target row: MediaId=$MEDIA_ID Title='$ORIG_TITLE'"

# Find the matching sidecar (any *.json under json/movie or json/tv).
find_sidecar_for() {
  local id="$1"
  grep -lR "\"title\": \"$ORIG_TITLE\"" "$JSON_DIR" 2>/dev/null | head -1
}
SIDECAR="$(find_sidecar_for "$MEDIA_ID")"
if [[ -z "$SIDECAR" ]]; then
  echo "❌ No sidecar found for '$ORIG_TITLE'"
  exit 1
fi
echo "📄 Sidecar: $SIDECAR"

# ===== Test 1: edit Media row → rerun scan → sidecar mtime advances =====
echo ""
echo "==== Test 1: sidecar rewrite on DB edit ===="
BEFORE_MTIME="$(mtime "$SIDECAR")"
NEW_TITLE="${ORIG_TITLE} [verify-$(date +%s)]"
sqlite3 "$DB_PATH" "
  UPDATE Media
  SET Title = '$NEW_TITLE', UpdatedAt = datetime('now')
  WHERE MediaId = $MEDIA_ID;"
echo "✏️  Title updated → '$NEW_TITLE'"

# Force mtime backwards 60s so the rewrite condition is unambiguous.
touch -t "$(date -v-1M +%Y%m%d%H%M.%S 2>/dev/null || date -d '-1 minute' +%Y%m%d%H%M.%S)" "$SIDECAR"

mahin scan "$FOLDER" --reverse-sync-only >/tmp/reverse-sync.log 2>&1
cat /tmp/reverse-sync.log | tail -10

AFTER_MTIME="$(mtime "$SIDECAR")"
SIDECAR_TITLE="$(grep -o '"title": *"[^"]*"' "$SIDECAR" | head -1)"

T1_PASS=true
if (( AFTER_MTIME <= BEFORE_MTIME )); then
  echo "❌ Test 1 FAIL: sidecar mtime did not advance ($BEFORE_MTIME → $AFTER_MTIME)"
  T1_PASS=false
fi
if [[ "$SIDECAR_TITLE" != *"$NEW_TITLE"* ]]; then
  echo "❌ Test 1 FAIL: sidecar title did not update; got: $SIDECAR_TITLE"
  T1_PASS=false
fi
$T1_PASS && echo "✅ Test 1 PASS: sidecar rewritten with new title"

# Restore original title so the user's library isn't permanently mutated.
sqlite3 "$DB_PATH" "
  UPDATE Media SET Title='$ORIG_TITLE', UpdatedAt=datetime('now')
  WHERE MediaId=$MEDIA_ID;"

# ===== Test 2: soft-delete row → rerun → sidecar removed =====
echo ""
echo "==== Test 2: sidecar removal on soft-delete ===="
sqlite3 "$DB_PATH" "
  UPDATE Media SET IsDeleted=1, UpdatedAt=datetime('now')
  WHERE MediaId=$MEDIA_ID;"
echo "🗑  Row $MEDIA_ID soft-deleted"

mahin scan "$FOLDER" --reverse-sync-only >/tmp/reverse-sync.log 2>&1
cat /tmp/reverse-sync.log | tail -10

T2_PASS=true
if [[ -f "$SIDECAR" ]]; then
  echo "❌ Test 2 FAIL: sidecar still present at $SIDECAR"
  T2_PASS=false
else
  echo "✅ Test 2 PASS: sidecar removed"
fi

# Restore row so re-scan can recreate sidecar.
sqlite3 "$DB_PATH" "
  UPDATE Media SET IsDeleted=0, UpdatedAt=datetime('now')
  WHERE MediaId=$MEDIA_ID;"
mahin scan "$FOLDER" >/dev/null 2>&1

# ===== Audit table check =====
echo ""
echo "==== Audit log (last 10 reconciliation rows) ===="
sqlite3 "$DB_PATH" "
  SELECT rh.OccurredAt, rat.Name, COALESCE(rh.Details,'')
  FROM ReconciliationHistory rh
  JOIN ReconciliationActionType rat
    ON rat.ReconciliationActionTypeId = rh.ReconciliationActionTypeId
  ORDER BY rh.ReconciliationHistoryId DESC LIMIT 10;"

echo ""
$T1_PASS && $T2_PASS && { echo "🎉 ALL REVERSE-SYNC CHECKS PASS"; exit 0; }
$T1_PASS || exit 2
exit 3
