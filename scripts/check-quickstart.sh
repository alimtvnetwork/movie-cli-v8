#!/usr/bin/env bash
# check-quickstart.sh — Verify QUICKSTART.md exists, is linked from README,
# and contains command blocks for Linux, macOS, and Windows.
#
# Exit 0 on success, 1 on any failure. Emits GitHub annotations when run in CI.

set -uo pipefail

ROOT="${1:-.}"
if [ -f "$ROOT/readme.md" ]; then
    README="$ROOT/readme.md"
else
    README="$ROOT/README.md"
fi
QS="$ROOT/QUICKSTART.md"

fail=0

annotate() {
    local file="$1" msg="$2"
    echo "::error file=${file}::${msg}"
    echo "  [FAIL] ${msg}"
    fail=1
}

# 1. Files exist
if [ ! -f "$QS" ]; then
    annotate "QUICKSTART.md" "QUICKSTART.md is missing from repo root"
    exit 1
fi
if [ ! -f "$README" ]; then
    annotate "README.md" "README.md is missing from repo root"
    exit 1
fi
echo "  [PASS] README.md and QUICKSTART.md present"

# 2. README links to QUICKSTART.md
if grep -qE '\]\(QUICKSTART\.md\)|\]\(\./QUICKSTART\.md\)' "$README"; then
    echo "  [PASS] README.md links to QUICKSTART.md"
else
    annotate "README.md" "README.md does not link to QUICKSTART.md (expected a (QUICKSTART.md) markdown link)"
fi

# 3. Required platform sections (case-insensitive heading match)
for platform in Linux macOS Windows; do
    if grep -qiE "^#{1,6}.*\\b${platform}\\b" "$QS"; then
        echo "  [PASS] QUICKSTART.md has a ${platform} section"
    else
        annotate "QUICKSTART.md" "QUICKSTART.md is missing a heading for ${platform}"
    fi
done

# 4. At least one fenced command block per platform.
#    Strategy: walk the file, track the most recent platform heading, count
#    fenced code blocks (``` …) appearing before the next platform heading.
awk '
    BEGIN { current=""; in_code=0 }
    /^#{1,6}.*[Ll]inux/   { current="Linux";   next }
    /^#{1,6}.*[Mm]ac.*OS|^#{1,6}.*[Mm]acOS/ { current="macOS"; next }
    /^#{1,6}.*[Ww]indows/ { current="Windows"; next }
    /^#{1,6} / && !/[Ll]inux|[Mm]ac|[Ww]indows/ { current=""; next }
    /^```/ {
        if (in_code == 0 && current != "") { count[current]++ }
        in_code = 1 - in_code
        next
    }
    END {
        for (p in count) print p "=" count[p]
    }
' "$QS" > /tmp/qs-blocks.txt

for platform in Linux macOS Windows; do
    n=$(grep "^${platform}=" /tmp/qs-blocks.txt | cut -d= -f2)
    n="${n:-0}"
    if [ "$n" -ge 1 ]; then
        echo "  [PASS] QUICKSTART.md has ${n} command block(s) under ${platform}"
    else
        annotate "QUICKSTART.md" "QUICKSTART.md ${platform} section has no fenced command block"
    fi
done

if [ "$fail" -ne 0 ]; then
    echo ""
    echo "QUICKSTART check failed."
    exit 1
fi
echo ""
echo "✅ QUICKSTART.md is wired up correctly."
exit 0
