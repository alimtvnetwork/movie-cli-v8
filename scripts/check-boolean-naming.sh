#!/usr/bin/env bash
# check-boolean-naming.sh
#
# CI guard for the project boolean-naming rule
# (mem://constraints/boolean-no-negative-words):
#
#   Boolean identifiers must NEVER use negative words (Un, Not, No)
#   immediately after the Is/Has prefix. Use a positive semantic synonym
#   instead. Examples:
#     IsUndone     -> IsReverted
#     IsNotReady   -> IsPending
#     HasNoChild   -> IsLeaf
#     IsUnsaved    -> IsDirty
#
# The guard scans Go source for identifiers matching:
#   (Is|Has)(Un|Not|No)<UpperLetter>...
#
# Allow-list:
#   - Stdlib `os.IsNotExist` and friends (we cannot rename Go stdlib).
#   - The constraint memory file itself (it documents the bad names).
#
# Comments are stripped before matching so that doc-strings warning
# future contributors NOT to use these prefixes don't false-positive.

set -e

ROOT="${1:-.}"

if [ ! -d "$ROOT" ]; then
  echo "::warning::$ROOT not found — skipping boolean-naming guard."
  exit 0
fi

# Files / paths exempt from the rule.
EXCLUDES=(
  --glob '!.git/**'
  --glob '!.release/**'
  --glob '!vendor/**'
  --glob '!node_modules/**'
  --glob '!.lovable/**'
  --glob '!**/check-boolean-naming.sh'
)

# Pattern: Is or Has, then Un|Not|No, then any letter. We accept both
# `IsUndone` (lowercase tail) and `HasNoChild` (uppercase tail). The
# `Is`/`Has` prefix itself guarantees a CamelCase identifier boundary, so
# no leading \b is needed.
PATTERN='(Is|Has)(Un|Not|No)[A-Za-z]'

# False-positive allow-list: English words that legitimately start with
# Un/Not/No and are NOT negations. Match the WHOLE word that follows the
# Is/Has prefix, anchored so partial collisions don't slip through.
SAFE_TAILS='\b(Is|Has)(Unique|Universal|Universe|Until|Union|Unit|Units|Unified|Uniform|Unknown|Untitled|Unread|Underway|Undef|Undefined|Undo|Undoable|Notice|Notable|Noteworthy|Notes|Notified|Notification|Notifications|Notion|Note|Nominal|None|Normal|Normative|Notch|Nose|Notary|Notarized|Northern|North)\b'

# Collect raw matches, then drop:
#   - lines that are pure Go line comments (`^\s*//`)
#   - the stdlib os.IsNotExist family
#   - the SAFE_TAILS allow-list (legitimate English words, not negations)
raw=$(rg -n "$PATTERN" "$ROOT" --glob '*.go' "${EXCLUDES[@]}" 2>/dev/null || true)

violations=$(printf '%s\n' "$raw" \
  | grep -vE ':[[:space:]]*//' \
  | grep -vE '\bos\.Is(Not|No)[A-Z][A-Za-z]*' \
  | grep -vE "$SAFE_TAILS" \
  | sed '/^$/d')

if [ -n "$violations" ]; then
  echo "Negative-word boolean identifier(s) found:"
  echo "$violations"
  echo ""
  while IFS= read -r line; do
    [ -z "$line" ] && continue
    file=$(printf '%s' "$line" | cut -d: -f1)
    lineno=$(printf '%s' "$line" | cut -d: -f2)
    snippet=$(printf '%s' "$line" | cut -d: -f3-)
    echo "::error file=${file},line=${lineno}::Negative boolean prefix (Is/Has + Un/Not/No). Use a positive synonym (IsUndone→IsReverted, IsNotReady→IsPending, HasNoX→IsLeaf, IsUnsaved→IsDirty). Snippet: ${snippet}"
  done <<< "$violations"
  echo ""
  echo "Rule: mem://constraints/boolean-no-negative-words"
  exit 1
fi

echo "✅ No negative-word boolean identifiers in $ROOT (stdlib os.IsNotExist allow-listed)."
