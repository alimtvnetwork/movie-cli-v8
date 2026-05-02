#!/usr/bin/env bash
# get.sh — Smart installer for movie-cli.
#
# Tries the GitHub Release asset first; if no release is published, falls back
# to source-build from main with a clear message either way.
#
# Resolution order:
#   1. GitHub Release  → releases/latest/download/install.sh   (pre-built binary)
#   2. Source-build    → raw.githubusercontent.com/…/main/install.sh (needs Git+Go)
#
# Invoked via:
#   curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/get.sh | bash

set -euo pipefail

OWNER='alimtvnetwork'
REPO='movie-cli-v8'
BRANCH='main'
RELEASE_URL="https://github.com/${OWNER}/${REPO}/releases/latest/download/install.sh"
SOURCE_URL="https://raw.githubusercontent.com/${OWNER}/${REPO}/${BRANCH}/install.sh"
RELEASES_UI="https://github.com/${OWNER}/${REPO}/releases"
PROBE_TIMEOUT=10

# ── Colors ────────────────────────────────────────────────────
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
GRAY='\033[0;90m'
DCYAN='\033[0;36m'
NC='\033[0m'

step()  { printf "  ${CYAN}->${NC} ${GRAY}%s${NC}\n" "$*"; }
ok()    { printf "  ${GREEN}OK${NC} ${GREEN}%s${NC}\n" "$*"; }
warn()  { printf "  ${YELLOW}!!${NC} ${YELLOW}%s${NC}\n" "$*"; }
note()  { printf "     ${GRAY}%s${NC}\n" "$*"; }
fail()  { printf "  ${RED}XX${NC} ${RED}%s${NC}\n" "$*"; }

echo ""
printf " ${DCYAN}+======================================+${NC}\n"
printf " ${DCYAN}|${NC} ${CYAN}movie smart installer${NC}                ${DCYAN}|${NC}\n"
printf " ${DCYAN}+======================================+${NC}\n"
echo ""

# ── 1. Probe the GitHub Release asset ─────────────────────────
step "Checking for a published GitHub Release..."

# -L follows the /releases/latest/ redirect; -I does HEAD; -o /dev/null discards
# body; -w prints the final HTTP code. Any non-200 → no release.
release_code=$(curl -sIL -o /dev/null -w '%{http_code}' \
                    --max-time "$PROBE_TIMEOUT" \
                    "$RELEASE_URL" 2>/dev/null || echo "000")

if [[ "$release_code" == "200" ]]; then
    ok "Release found — installing pre-built binary"
    note "Source: $RELEASE_URL"
    echo ""
    curl -fsSL --max-time 30 "$RELEASE_URL" | bash
    exit $?
fi

# ── 2. Fall back to source-build ──────────────────────────────
warn "No published GitHub Release found for ${OWNER}/${REPO} (HTTP ${release_code})."
note "Falling back to source-build from branch '${BRANCH}'."
note "(This needs Git + Go 1.22+ on PATH. Build takes ~30s.)"
echo ""
note "Tip for maintainers: publish a release at"
note "  ${RELEASES_UI}"
note "so future installs grab a pre-built binary instead of compiling locally."
echo ""
step "Downloading source installer: $SOURCE_URL"

source_code=$(curl -sIL -o /dev/null -w '%{http_code}' \
                   --max-time "$PROBE_TIMEOUT" \
                   "$SOURCE_URL" 2>/dev/null || echo "000")

if [[ "$source_code" != "200" ]]; then
    echo ""
    fail "Source installer is also unreachable (HTTP ${source_code})."
    note "URL: $SOURCE_URL"
    echo ""
    note "What to try next:"
    note "  1. Check your internet connection / corporate proxy."
    note "  2. Open ${RELEASES_UI} in a browser to see if releases exist."
    note "  3. Clone manually:"
    note "       git clone https://github.com/${OWNER}/${REPO}.git"
    note "       cd ${REPO} && bash ./install.sh"
    exit 1
fi

curl -fsSL --max-time 60 "$SOURCE_URL" | bash
exit $?
