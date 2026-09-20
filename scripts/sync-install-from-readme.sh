#!/usr/bin/env bash
# sync-install-from-readme.sh
#
# Extracts the canonical install block from README.md and rewrites the
# install snippets in QUICKSTART.md and spec/03-general/01-install-guide.md
# so wording and headers stay identical to the root README.
#
# Source of truth: README.md (per mem://preferences/readme-structure).
# Sub-docs MUST contain sentinel markers around the install block:
#
#   <!-- INSTALL:BEGIN -->
#   ...generated content...
#   <!-- INSTALL:END -->
#
# Anything outside the markers is preserved verbatim.
#
# Usage:
#   scripts/sync-install-from-readme.sh                       # rewrite files
#   scripts/sync-install-from-readme.sh --check               # exit 1 if drift
#   scripts/sync-install-from-readme.sh --init-markers        # add sentinels if missing
#   scripts/sync-install-from-readme.sh --print               # print extracted block
#   scripts/sync-install-from-readme.sh --json                # print block + metadata as JSON
#   scripts/sync-install-from-readme.sh --list-targets        # show resolved target list
#   scripts/sync-install-from-readme.sh --targets a.md,b.md   # one-shot custom targets
#   scripts/sync-install-from-readme.sh --discover            # also auto-find any *.md
#                                                             # with INSTALL:BEGIN sentinels
#
# Targets are resolved in this order (first wins):
#   1. --targets flag (comma-separated)
#   2. SYNC_INSTALL_TARGETS env var (comma- or newline-separated)
#   3. scripts/sync-install-targets.txt (one path per line, # for comments)
#   4. Built-in defaults: QUICKSTART.md, spec/03-general/01-install-guide.md
# Paths may be absolute or relative to the repo root.
#
# Exit codes:
#   0  success / no drift
#   1  drift detected (--check) or missing markers
#   2  README install block not found / unknown option

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
if [ -f "$ROOT/readme.md" ]; then
  README="$ROOT/readme.md"
else
  README="$ROOT/README.md"
fi
CONFIG_FILE="$ROOT/scripts/sync-install-targets.txt"

# Default targets — used when no config file, env var, or flag is provided.
DEFAULT_TARGETS=(
  "QUICKSTART.md"
  "spec/03-general/01-install-guide.md"
)

BEGIN_MARK="<!-- INSTALL:BEGIN -->"
END_MARK="<!-- INSTALL:END -->"

CHECK_ONLY=0
INIT_MARKERS=0
PRINT_ONLY=0
JSON_ONLY=0
DISCOVER=0
TARGETS_FLAG=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check)         CHECK_ONLY=1 ;;
    --init-markers)  INIT_MARKERS=1 ;;
    --print)         PRINT_ONLY=1 ;;
    --json)          JSON_ONLY=1 ;;
    --discover)      DISCOVER=1 ;;
    --targets)       TARGETS_FLAG="${2:-}"; shift ;;
    --targets=*)     TARGETS_FLAG="${1#--targets=}" ;;
    --list-targets)  LIST_ONLY=1 ;;
    "")              ;;
    *) echo "Unknown option: $1" >&2; exit 2 ;;
  esac
  shift
done
LIST_ONLY="${LIST_ONLY:-0}"

# --- Resolve target list (priority: flag > env > config file > defaults) ----
# Each entry may be absolute or relative to repo root. Lines starting with #
# and blank lines in the config file are ignored.

resolve_targets() {
  local raw=""
  if [[ -n "$TARGETS_FLAG" ]]; then
    raw="${TARGETS_FLAG//,/$'\n'}"
  elif [[ -n "${SYNC_INSTALL_TARGETS:-}" ]]; then
    raw="${SYNC_INSTALL_TARGETS//,/$'\n'}"
  elif [[ -f "$CONFIG_FILE" ]]; then
    raw="$(grep -vE '^[[:space:]]*(#|$)' "$CONFIG_FILE" || true)"
  else
    raw="$(printf '%s\n' "${DEFAULT_TARGETS[@]}")"
  fi

  TARGETS=()
  while IFS= read -r line; do
    line="${line#"${line%%[![:space:]]*}"}"   # ltrim
    line="${line%"${line##*[![:space:]]}"}"   # rtrim
    [[ -z "$line" ]] && continue
    if [[ "$line" = /* ]]; then TARGETS+=("$line"); else TARGETS+=("$ROOT/$line"); fi
  done <<< "$raw"

  # --discover: also append any *.md under ROOT that contains the sentinels
  # but isn't already in the list. Skips node_modules, .git, .release, dist.
  if [[ "$DISCOVER" -eq 1 ]]; then
    local found already t
    while IFS= read -r found; do
      already=0
      for t in "${TARGETS[@]}"; do
        if [[ "$t" == "$found" ]]; then already=1; break; fi
      done
      if [[ "$already" -eq 0 ]]; then TARGETS+=("$found"); fi
    done < <(
      grep -RlF --include='*.md' \
        --exclude-dir=node_modules \
        --exclude-dir=.git \
        --exclude-dir=.release \
        --exclude-dir=dist \
        "$BEGIN_MARK" "$ROOT" 2>/dev/null \
        | grep -vF "$README" || true
    )
  fi
}

resolve_targets

if [[ "$LIST_ONLY" -eq 1 ]]; then
  printf 'Resolved %d target(s):\n' "${#TARGETS[@]}"
  for t in "${TARGETS[@]}"; do printf '  %s\n' "$t"; done
  exit 0
fi


# --- 0. Optionally insert sentinels into targets that lack them -------------
# Appends a fresh install section (with INSTALL:BEGIN / INSTALL:END markers)
# to the end of any target file that is missing the sentinels. Existing
# install content is left untouched — the sync step below will fill the new
# block on the next run.

init_markers() {
  local file="$1"
  if [[ ! -f "$file" ]]; then
    echo "SKIP: $file not found" >&2
    return 0
  fi
  if grep -qF "$BEGIN_MARK" "$file" && grep -qF "$END_MARK" "$file"; then
    echo "OK:    $file already has sentinels"
    return 0
  fi
  {
    printf '\n\n## Install\n\n'
    printf '%s\n\n%s\n' "$BEGIN_MARK" "$END_MARK"
  } >> "$file"
  echo "INIT:  $file sentinels appended"
}

if [[ "$INIT_MARKERS" -eq 1 ]]; then
  rc=0
  for f in "${TARGETS[@]}"; do
    init_markers "$f" || rc=1
  done
  exit $rc
fi

# --- 1. Extract install block from README -----------------------------------
# Strategy (most → least reliable):
#   1. Explicit sentinels in README:
#        <!-- README-INSTALL:BEGIN -->  ...  <!-- README-INSTALL:END -->
#      This is the canonical source — surrounding headings and tables can
#      change freely without breaking the sync.
#   2. Heuristic fallback for older READMEs: capture from the
#      "🚀 Install in 10 seconds" line through the first <sub>…</sub>
#      caption that follows the closing </table>.
# Both strategies trim trailing blank lines so the generated block is stable.

extract_by_sentinels() {
  awk '
    /<!-- README-INSTALL:BEGIN -->/ { capture=1; next }
    /<!-- README-INSTALL:END -->/   { exit }
    capture { print }
  ' "$README"
}

extract_by_heuristic() {
  awk '
    /\*\*🚀 Install in 10 seconds/ { capture=1 }
    capture { print }
    /<\/table>/ { seen_table=1 }
    capture && seen_table && /<\/sub>[[:space:]]*$/ { exit }
  ' "$README"
}

trim_blank_edges() {
  awk '
    { lines[NR]=$0 }
    END {
      first=1; last=NR
      while (first<=last && lines[first] ~ /^[[:space:]]*$/) first++
      while (last>=first && lines[last]  ~ /^[[:space:]]*$/) last--
      for (i=first; i<=last; i++) print lines[i]
    }
  '
}

extract_block() {
  local out
  out="$(extract_by_sentinels | trim_blank_edges)"
  if [[ -n "$out" ]]; then
    echo "$out"
    return 0
  fi
  echo "INFO: README sentinels not found — falling back to heuristic" >&2
  extract_by_heuristic | trim_blank_edges
}

BLOCK="$(extract_block)"
if [[ -z "$BLOCK" ]]; then
  echo "ERROR: install block not found in README.md (no sentinels, no heuristic match)" >&2
  exit 2
fi

GENERATED=$'<!-- Generated from README.md by scripts/sync-install-from-readme.sh — do not edit by hand -->\n\n'"$BLOCK"

# --- 1b. --print: emit the extracted block to stdout and exit ---------------
# Useful for previewing exactly what would be written before touching any
# target file. Header/footer go to stderr so the body on stdout stays
# pipe-friendly (e.g. `... --print | less`, `... --print > preview.md`,
# `diff <(... --print) <(sed -n '/INSTALL:BEGIN/,/INSTALL:END/p' QUICKSTART.md)`).

if [[ "$PRINT_ONLY" -eq 1 ]]; then
  {
    echo "----- README install block (extracted from README.md) -----"
    echo "----- $(printf '%s' "$BLOCK" | wc -l | tr -d ' ') lines, $(printf '%s' "$BLOCK" | wc -c | tr -d ' ') bytes -----"
  } >&2
  printf '%s\n' "$BLOCK"
  echo "----- end of block -----" >&2
  exit 0
fi

# --- 1c. --json: emit metadata + block as a JSON object on stdout -----------
# Schema (stable, additive):
#   {
#     "source":     "README.md",
#     "extracted":  "sentinels" | "heuristic",
#     "lines":      <int>,
#     "bytes":      <int>,
#     "sha256":     "<hex>",
#     "targets":    ["QUICKSTART.md", ...],   // resolved, relative to repo root
#     "block":      "<full install block as a string>"
#   }
# Stdout is pure JSON (no logs) so it can be piped into jq, GitHub Actions
# `$GITHUB_OUTPUT`, etc. Diagnostics still go to stderr.

if [[ "$JSON_ONLY" -eq 1 ]]; then
  json_lines="$(printf '%s' "$BLOCK" | wc -l | tr -d ' ')"
  json_bytes="$(printf '%s' "$BLOCK" | wc -c | tr -d ' ')"
  json_sha="$(printf '%s' "$BLOCK" | sha256sum | awk '{print $1}')"
  json_method="sentinels"
  if ! grep -qF '<!-- README-INSTALL:BEGIN -->' "$README"; then
    json_method="heuristic"
  fi

  # Build "targets" array as repo-root-relative paths.
  rel_targets=()
  for t in "${TARGETS[@]}"; do
    rel_targets+=("${t#"$ROOT/"}")
  done

  # JSON-escape the block via python (handles quotes, backslashes, newlines,
  # unicode emojis safely). Falls back to python3 / python.
  py=""
  for cand in python3 python; do
    if command -v "$cand" >/dev/null 2>&1; then py="$cand"; break; fi
  done
  if [[ -z "$py" ]]; then
    echo "ERROR: --json requires python3 or python on PATH" >&2
    exit 2
  fi

  BLOCK_VAL="$BLOCK" \
  TARGETS_VAL="$(printf '%s\n' "${rel_targets[@]}")" \
  METHOD="$json_method" LINES="$json_lines" BYTES="$json_bytes" SHA="$json_sha" \
  "$py" - <<'PY'
import json, os
targets = [t for t in os.environ["TARGETS_VAL"].splitlines() if t]
out = {
    "source":    "README.md",
    "extracted": os.environ["METHOD"],
    "lines":     int(os.environ["LINES"]),
    "bytes":     int(os.environ["BYTES"]),
    "sha256":    os.environ["SHA"],
    "targets":   targets,
    "block":     os.environ["BLOCK_VAL"],
}
print(json.dumps(out, indent=2, ensure_ascii=False))
PY
  exit 0
fi

# --- 2. Replace block in each target between sentinels ----------------------

sync_file() {
  local file="$1"

  if [[ ! -f "$file" ]]; then
    echo "SKIP: $file not found" >&2
    return 0
  fi

  if ! grep -qF "$BEGIN_MARK" "$file" || ! grep -qF "$END_MARK" "$file"; then
    echo "ERROR: $file missing $BEGIN_MARK / $END_MARK sentinels" >&2
    return 1
  fi

  local tmp
  tmp="$(mktemp)"

  awk -v begin="$BEGIN_MARK" -v end="$END_MARK" -v block="$GENERATED" '
    BEGIN { inside=0 }
    {
      if (index($0, begin)) {
        print begin
        print ""
        print block
        print ""
        print end
        inside=1
        next
      }
      if (inside && index($0, end)) { inside=0; next }
      if (!inside) print
    }
  ' "$file" > "$tmp"

  if [[ "$CHECK_ONLY" -eq 1 ]]; then
    local h1 h2
    h1="$(sha256sum < "$file" | awk '{print $1}')"
    h2="$(sha256sum < "$tmp"  | awk '{print $1}')"
    if [[ "$h1" != "$h2" ]]; then
      echo "DRIFT: $file is out of sync with README.md — run scripts/sync-install-from-readme.sh" >&2
      rm -f "$tmp"
      return 1
    fi
    rm -f "$tmp"
    echo "OK:    $file in sync"
  else
    mv "$tmp" "$file"
    echo "WROTE: $file"
  fi
}

rc=0
for f in "${TARGETS[@]}"; do
  sync_file "$f" || rc=1
done

exit $rc
