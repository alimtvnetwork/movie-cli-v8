#!/usr/bin/env bash
# bootstrap.sh — Version-discovery bootstrap (Bash equivalent of bootstrap.ps1).
#
# Implements spec/03-general/05-install-latest-sibling-repo.md.
#
# Given the starting repo URL embedded below, probes sibling repos
# (-v<N+k>) on the same GitHub owner, picks the highest existing one,
# and delegates install to that repo's install.sh.
#
# Invoked by users via:
#   curl -fsSL https://raw.githubusercontent.com/<owner>/<base>-v<N>/main/bootstrap.sh | bash
#
# Each sibling repo ships its own copy — they all behave identically
# because the algorithm probes by base name, not by version.
#
# No retries. No caching. No GitHub API. 5s timeout per probe.

set -euo pipefail

# ── Inputs ────────────────────────────────────────────────────
# Default starting URL — overridable via first positional argument.
REPO_URL="${1:-https://github.com/alimtvnetwork/movie-cli-v8}"

# ── Constants (spec §3) ───────────────────────────────────────
MAX_LOOKAHEAD=25
PROBE_TIMEOUT_SEC=5
PROBE_BRANCH="main"

LOG_FILE="/tmp/movie-bootstrap.log"

# ── Colors ────────────────────────────────────────────────────
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
GRAY='\033[0;90m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# ── Logging ───────────────────────────────────────────────────
log() {
    local color="$1"; shift
    local msg="$*"
    printf "${color}%s${NC}\n" "$msg"
    printf "%s\n" "$msg" >> "$LOG_FILE" 2>/dev/null || true
}
log_inline() {
    local color="$1"; shift
    printf "${color}%s${NC}" "$*"
}

# ── URL parsing (spec §2) ─────────────────────────────────────
# Sets globals: PARSED_OWNER, PARSED_BASE, PARSED_N. Returns 0 on match.
parse_repo_url() {
    local url="$1"
    # Strip trailing slash and .git
    url="${url%/}"
    url="${url%.git}"
    # Reject /tree/ or /blob/
    if [[ "$url" == */tree/* || "$url" == */blob/* ]]; then
        log "$RED" "[bootstrap] error: only bare repo URLs supported (no /tree/ or /blob/)"
        return 1
    fi
    if [[ "$url" =~ ^https://github\.com/([^/]+)/(.+)-v([0-9]+)$ ]]; then
        PARSED_OWNER="${BASH_REMATCH[1]}"
        PARSED_BASE="${BASH_REMATCH[2]}"
        PARSED_N="${BASH_REMATCH[3]}"
        PARSED_CLEAN="$url"
        return 0
    fi
    return 1
}

# ── Probe one candidate (spec §4 step 3) ──────────────────────
# Echoes "HIT <url>" or "MISS <reason>".
probe_install_script() {
    local owner="$1" base="$2" version="$3"
    local url="https://raw.githubusercontent.com/${owner}/${base}-v${version}/${PROBE_BRANCH}/install.sh"
    local http_code
    # -f makes curl exit non-zero on 4xx/5xx, but we still want the code
    # captured by -w. Drop -f so the body is suppressed (-o /dev/null) but
    # the status line is always written.
    http_code=$(curl -sS -o /dev/null -w '%{http_code}' \
                     --max-time "$PROBE_TIMEOUT_SEC" \
                     "$url" 2>/dev/null || echo "000")
    if [[ "$http_code" == "200" ]]; then
        echo "HIT $url"
        return 0
    fi
    case "$http_code" in
        000) echo "MISS timeout/dns" ;;
        404) echo "MISS 404" ;;
        *)   echo "MISS http_$http_code" ;;
    esac
    return 1
}

# ── Find latest sibling (spec §4 step 2-3) ────────────────────
# Sets globals: WINNER_VERSION, WINNER_URL. Returns 0 if found.
find_latest_sibling() {
    local k v result
    for (( k=MAX_LOOKAHEAD; k>=0; k-- )); do
        v=$(( PARSED_N + k ))
        log_inline "$GRAY" "[bootstrap] probing v${v} ... "
        result=$(probe_install_script "$PARSED_OWNER" "$PARSED_BASE" "$v")
        if [[ "$result" == HIT* ]]; then
            log "$GREEN" "HIT"
            WINNER_VERSION="$v"
            WINNER_URL="${result#HIT }"
            return 0
        fi
        log "$GRAY" "miss (${result#MISS })"
    done
    return 1
}

# ── Delegate to the winner's install.sh (spec §4 step 4) ──────
delegate_install() {
    local install_url="$1"
    log "$MAGENTA" "[bootstrap] delegating to: $install_url"
    # Pipe the remote script straight to bash, propagate its exit code
    curl -fsSL --max-time 30 "$install_url" | bash
}

# ── Entry point ───────────────────────────────────────────────
log "$CYAN" "[bootstrap] starting URL: $REPO_URL"

if ! parse_repo_url "$REPO_URL"; then
    log "$YELLOW" "[bootstrap] URL has no -v<N> suffix; installing as-is"
    fallback_url="${REPO_URL%/}"
    fallback_url="${fallback_url//github.com/raw.githubusercontent.com}/${PROBE_BRANCH}/install.sh"
    delegate_install "$fallback_url"
    exit $?
fi

log "$CYAN" "[bootstrap] parsed: owner=${PARSED_OWNER} base=${PARSED_BASE} current=v${PARSED_N}"

if ! find_latest_sibling; then
    log "$YELLOW" "[bootstrap] no -v<N..N+${MAX_LOOKAHEAD}> repo found; falling back to starting URL"
    fallback_url="https://raw.githubusercontent.com/${PARSED_OWNER}/${PARSED_BASE}-v${PARSED_N}/${PROBE_BRANCH}/install.sh"
    delegate_install "$fallback_url"
    exit $?
fi

log "$CYAN" "[bootstrap] selected: https://github.com/${PARSED_OWNER}/${PARSED_BASE}-v${WINNER_VERSION}"
if [[ "$WINNER_VERSION" != "$PARSED_N" ]]; then
    log "$CYAN" "[bootstrap] auto-upgrade: v${PARSED_N} -> v${WINNER_VERSION}"
fi
delegate_install "$WINNER_URL"
exit $?
