#!/usr/bin/env bash
# audit-legacy-paths.sh — Scan the entire repo for legacy module paths and
# print a categorized report.
#
# This is a READ-ONLY audit tool. It never modifies files.
#
# Variants detected (all flagged unless they resolve to the current
# canonical path `github.com/alimtvnetwork/movie-cli-v8`):
#
#   VERSIONED       movie-cli-v[1-6]                  — old numbered modules
#   UNVERSIONED     github.com/alimtvnetwork/movie-cli (no -vN suffix)
#   TAGGED          movie-cli-vN@vX.Y.Z   or   movie-cli@vX.Y.Z
#                                                     — pinned go-module tags
#   REPLACE         go.mod `replace` directive pointing at any legacy path
#
# Categories:
#   HISTORICAL (allowed)  — references in CHANGELOG.md, spec/, .lovable/
#                           memory, this script, and CI guard patterns.
#   ACTIVE    (must fix)  — anywhere else (README, scripts, Go source,
#                           go.mod, install scripts, configs, etc.).
#
# Usage:
#   bash scripts/audit-legacy-paths.sh
#   bash scripts/audit-legacy-paths.sh --strict   # exit 1 if any ACTIVE found
#   bash scripts/audit-legacy-paths.sh --json     # machine-readable output

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CANONICAL='github.com/alimtvnetwork/movie-cli-v8'

# Combined pattern, ERE. Each match is later classified by variant.
#   1) movie-cli-v8..v6 (with optional @vX tag)
#   2) movie-cli (no -vN) optionally followed by @vX
#   3) go.mod `replace` directive on its own line referencing movie-cli*
PATTERN_VERSIONED='movie-cli-v[123456]\b(@v[0-9][^[:space:]"'"'"']*)?'
PATTERN_UNVERSIONED='github\.com/alimtvnetwork/movie-cli\b(@v[0-9][^[:space:]"'"'"']*)?(/|$|[^v0-9-])'
PATTERN_REPLACE='^[[:space:]]*replace[[:space:]].*movie-cli'
PATTERN="(${PATTERN_VERSIONED})|(${PATTERN_UNVERSIONED})|(${PATTERN_REPLACE})"

STRICT=0
JSON=0
for arg in "$@"; do
    case "$arg" in
        --strict) STRICT=1 ;;
        --json)   JSON=1 ;;
        -h|--help)
            sed -n '2,28p' "$0"; exit 0 ;;
        *) echo "Unknown arg: $arg" >&2; exit 2 ;;
    esac
done

cd "$ROOT"

RAW="$(grep -rnE "$PATTERN" \
    --exclude-dir=.git \
    --exclude-dir=node_modules \
    --exclude-dir=.release \
    --exclude-dir=.gitmap \
    --exclude-dir=audit-reports \
    --exclude-dir=.lovable \
    --exclude-dir=dist \
    --exclude-dir=build \
    . 2>/dev/null || true)"

# Drop matches that are exactly the canonical path with no tag and no
# legacy suffix — those are clean references to the current module.
filter_canonical() {
    while IFS= read -r line; do
        [ -z "$line" ] && continue
        content=$(printf '%s' "$line" | cut -d: -f3-)
        # Skip if the only movie-cli mention is the canonical path (no @vX tag).
        if echo "$content" | grep -qE 'movie-cli-v[123456]\b'; then
            printf '%s\n' "$line"; continue
        fi
        if echo "$content" | grep -qE '^[[:space:]]*replace[[:space:]].*movie-cli'; then
            printf '%s\n' "$line"; continue
        fi
        if echo "$content" | grep -qE 'github\.com/alimtvnetwork/movie-cli(@v[0-9]|[^v0-9-]|$)' \
           && ! echo "$content" | grep -qE 'github\.com/alimtvnetwork/movie-cli-v8\b' ; then
            printf '%s\n' "$line"; continue
        fi
        # Tagged canonical (e.g. movie-cli-v8@v1.2.3) — also flag, tags shouldn't be pinned in code.
        if echo "$content" | grep -qE 'movie-cli-v8@v[0-9]'; then
            printf '%s\n' "$line"; continue
        fi
    done
}

RAW="$(printf '%s\n' "$RAW" | filter_canonical)"

if [ -z "$RAW" ]; then
    if [ "$JSON" -eq 1 ]; then
        echo '{"historical":[],"active":[],"summary":{"historical":0,"active":0}}'
    else
        echo "✅ No legacy module path references found anywhere."
        echo "   (Variants checked: VERSIONED, UNVERSIONED, TAGGED, REPLACE)"
    fi
    exit 0
fi

# Classify variant for a single line of content.
classify_variant() {
    local content="$1"
    if echo "$content" | grep -qE '^[[:space:]]*replace[[:space:]].*movie-cli'; then
        echo "REPLACE"; return
    fi
    if echo "$content" | grep -qE 'movie-cli(-v[1-7])?@v[0-9]'; then
        echo "TAGGED"; return
    fi
    if echo "$content" | grep -qE 'movie-cli-v[123456]\b'; then
        echo "VERSIONED"; return
    fi
    if echo "$content" | grep -qE 'github\.com/alimtvnetwork/movie-cli(@v[0-9]|[^v0-9-]|$)'; then
        echo "UNVERSIONED"; return
    fi
    echo "OTHER"
}

# HISTORICAL paths (allowed).
HISTORICAL_FILTER='^\./(CHANGELOG\.md|spec/|\.lovable/|scripts/audit-legacy-paths\.sh|\.github/workflows/ci\.yml)'

historical=""
active=""
while IFS= read -r line; do
    [ -z "$line" ] && continue
    content=$(printf '%s' "$line" | cut -d: -f3-)
    variant=$(classify_variant "$content")
    tagged_line="[${variant}] ${line}"
    if echo "$line" | grep -qE "$HISTORICAL_FILTER"; then
        historical="${historical}${tagged_line}"$'\n'
    else
        active="${active}${tagged_line}"$'\n'
    fi
done <<< "$RAW"

h_count=$(printf '%s' "$historical" | grep -c . || true)
a_count=$(printf '%s' "$active" | grep -c . || true)

if [ "$JSON" -eq 1 ]; then
    json_array() {
        local input="$1"
        local first=1
        printf '['
        while IFS= read -r line; do
            [ -z "$line" ] && continue
            variant=$(printf '%s' "$line" | sed -nE 's/^\[([A-Z]+)\] .*/\1/p')
            rest=$(printf '%s' "$line" | sed -E 's/^\[[A-Z]+\] //')
            file=$(printf '%s' "$rest" | cut -d: -f1)
            lineno=$(printf '%s' "$rest" | cut -d: -f2)
            content=$(printf '%s' "$rest" | cut -d: -f3- | sed 's/\\/\\\\/g; s/"/\\"/g')
            [ "$first" -eq 0 ] && printf ','
            printf '{"variant":"%s","file":"%s","line":%s,"content":"%s"}' \
                "$variant" "$file" "$lineno" "$content"
            first=0
        done <<< "$input"
        printf ']'
    }
    printf '{"historical":'
    json_array "$historical"
    printf ',"active":'
    json_array "$active"
    printf ',"summary":{"historical":%s,"active":%s,"canonical":"%s"}}\n' \
        "$h_count" "$a_count" "$CANONICAL"
    [ "$STRICT" -eq 1 ] && [ "$a_count" -gt 0 ] && exit 1
    exit 0
fi

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "  Legacy module-path audit"
echo "  Variants : VERSIONED | UNVERSIONED | TAGGED | REPLACE"
echo "  Canonical: ${CANONICAL}"
echo "  Root     : ${ROOT}"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo "── HISTORICAL (allowed) ─────────────────  ${h_count} match(es)"
echo "  Files: CHANGELOG.md, spec/**, .lovable/**, this script, ci.yml guards."
echo ""
if [ "$h_count" -gt 0 ]; then
    printf '%s' "$historical" | sed 's/^/  /'
    echo ""
fi

echo "── ACTIVE (must fix) ─────────────────────  ${a_count} match(es)"
echo "  Anywhere outside the historical zone."
echo ""
if [ "$a_count" -gt 0 ]; then
    printf '%s' "$active" | sed 's/^/  /'
    echo ""
    echo "❌ ${a_count} active legacy reference(s) found."
    echo "   Fix guidance by variant:"
    echo "     VERSIONED   → replace path with ${CANONICAL}"
    echo "     UNVERSIONED → add the -v7 suffix: ${CANONICAL}"
    echo "     TAGGED      → drop the @vX.Y.Z pin; depend on the module head"
    echo "     REPLACE     → remove or repoint the go.mod replace directive"
else
    echo "  ✅ none"
fi
echo ""
echo "── Summary ──────────────────────────────────────────────────────"
echo "  Historical : ${h_count}  (informational, OK to keep)"
echo "  Active     : ${a_count}  $( [ "$a_count" -gt 0 ] && echo '(action required)' || echo '(clean)')"
echo ""

if [ "$STRICT" -eq 1 ] && [ "$a_count" -gt 0 ]; then
    exit 1
fi
exit 0
