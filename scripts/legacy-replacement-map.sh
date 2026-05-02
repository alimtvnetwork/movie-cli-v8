#!/usr/bin/env bash
# legacy-replacement-map.sh — Generate a file-and-line replacement map from
# the strict auditor output so legacy references can be updated quickly.
#
# Reads JSON from `scripts/audit-legacy-paths.sh --json` and emits, for every
# ACTIVE match, a line in the form:
#
#   <file>:<line>  [VARIANT]  '<old>' -> '<new>'
#
# Replacement rules per variant (canonical = github.com/alimtvnetwork/movie-cli-v8):
#
#   VERSIONED     movie-cli-v[1-6]            -> movie-cli-v8   (drop @vX.Y.Z if any)
#   UNVERSIONED   .../movie-cli (no -vN)      -> .../movie-cli-v8
#   TAGGED        movie-cli(-v7)?@vX.Y.Z      -> movie-cli-v8   (drop the tag)
#   REPLACE       go.mod replace ... movie-cli*   -> REMOVE the directive
#
# Usage:
#   bash scripts/legacy-replacement-map.sh                # plain text map
#   bash scripts/legacy-replacement-map.sh --json         # machine-readable
#   bash scripts/legacy-replacement-map.sh --out FILE     # write to file
#
# Exit codes:
#   0  no active matches (nothing to replace)
#   1  active matches found and a map was produced
#   2  bad arguments / missing dependencies

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CANONICAL='github.com/alimtvnetwork/movie-cli-v8'
AUDITOR="${ROOT}/scripts/audit-legacy-paths.sh"

JSON_OUT=0
OUT_FILE=""

while [ $# -gt 0 ]; do
    case "$1" in
        --json) JSON_OUT=1 ;;
        --out)  shift; OUT_FILE="${1:-}" ;;
        -h|--help) sed -n '2,30p' "$0"; exit 0 ;;
        *) echo "Unknown arg: $1" >&2; exit 2 ;;
    esac
    shift
done

command -v python3 >/dev/null 2>&1 || {
    echo "python3 is required" >&2; exit 2;
}
[ -x "$AUDITOR" ] || [ -f "$AUDITOR" ] || {
    echo "Auditor not found at $AUDITOR" >&2; exit 2;
}

AUDIT_JSON="$(bash "$AUDITOR" --json 2>/dev/null || true)"
if [ -z "$AUDIT_JSON" ]; then
    echo "Auditor produced no output" >&2; exit 2
fi

RESULT="$(CANONICAL="$CANONICAL" JSON_OUT="$JSON_OUT" AUDIT_JSON="$AUDIT_JSON" python3 - <<'PY'
import json, os, re, sys

canonical = os.environ["CANONICAL"]
json_out  = os.environ["JSON_OUT"] == "1"
data = json.loads(os.environ["AUDIT_JSON"])

active = data.get("active", [])

# Match the legacy needle inside the source line.
re_versioned   = re.compile(r'github\.com/alimtvnetwork/movie-cli-v[1-6](@v[0-9][^\s"\']*)?')
re_tagged_any  = re.compile(r'github\.com/alimtvnetwork/movie-cli(-v[1-7])?@v[0-9][^\s"\']*')
re_unversioned = re.compile(r'github\.com/alimtvnetwork/movie-cli(?!-v\d)(@v[0-9][^\s"\']*)?')

def derive(variant, content):
    """Return list of (old, new) replacements for one source line."""
    pairs = []
    if variant == "REPLACE":
        # Whole replace directive — instruct removal.
        pairs.append((content.strip(), "<<REMOVE LINE>>"))
        return pairs
    if variant == "TAGGED":
        for m in re_tagged_any.finditer(content):
            pairs.append((m.group(0), canonical))
        if pairs: return pairs
    if variant == "VERSIONED":
        for m in re_versioned.finditer(content):
            pairs.append((m.group(0), canonical))
        if pairs: return pairs
    if variant == "UNVERSIONED":
        for m in re_unversioned.finditer(content):
            old = m.group(0)
            # Skip if it's actually the canonical (shouldn't be, but safe).
            if old.endswith("movie-cli-v8"): continue
            pairs.append((old, canonical))
        if pairs: return pairs
    # Fallback: any movie-cli mention.
    for m in re.finditer(r'github\.com/alimtvnetwork/movie-cli[\w\-@.]*', content):
        old = m.group(0)
        if old == canonical: continue
        pairs.append((old, canonical))
    return pairs

rows = []
for item in active:
    pairs = derive(item["variant"], item["content"])
    for old, new in pairs:
        rows.append({
            "file":    item["file"],
            "line":    item["line"],
            "variant": item["variant"],
            "old":     old,
            "new":     new,
        })

if json_out:
    print(json.dumps({
        "canonical": canonical,
        "count":     len(rows),
        "replacements": rows,
    }, indent=2))
else:
    if not rows:
        print("✅ No active legacy references — nothing to replace.")
    else:
        print(f"Replacement map ({len(rows)} edit(s)) — canonical: {canonical}")
        print("─" * 72)
        for r in rows:
            if r["new"] == "<<REMOVE LINE>>":
                print(f"{r['file']}:{r['line']}  [{r['variant']}]  REMOVE LINE: {r['old']}")
            else:
                print(f"{r['file']}:{r['line']}  [{r['variant']}]  '{r['old']}' -> '{r['new']}'")
        print("─" * 72)
        print("Apply with sed/IDE find-replace; re-run scripts/audit-legacy-paths.sh --strict to verify.")

# Exit code: 1 if any rows produced (work to do), else 0.
sys.exit(1 if rows else 0)
PY
)"
RC=$?

if [ -n "$OUT_FILE" ]; then
    printf '%s\n' "$RESULT" > "$OUT_FILE"
    echo "Wrote replacement map to $OUT_FILE"
else
    printf '%s\n' "$RESULT"
fi

exit $RC
