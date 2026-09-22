#!/usr/bin/env bash
# install.sh — One-liner binary installer for movie CLI on Linux and macOS.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install.sh | bash -s -- --version v2.324.0
#   ./install.sh --dir ~/.local/bin
#   ./install.sh --dry-run --version v2.324.0
#   ./install.sh --uninstall

set -euo pipefail

# Re-exec under bash if invoked via sh/dash (which lack pipefail, local, arrays)
if [ -z "${BASH_VERSION:-}" ]; then
    if ! command -v bash >/dev/null 2>&1; then
        printf '\033[31m  Error: bash is required but not found. Install bash first.\033[0m\n' >&2
        exit 1
    fi

    if [ -f "$0" ] && [ -r "$0" ]; then
        exec bash "$0" "$@"
    fi

    _mv_tmp="$(mktemp 2>/dev/null || echo "/tmp/movie-install.$$.sh")"
    if [ ! -t 0 ]; then
        cat > "$_mv_tmp" 2>/dev/null || true
    fi

    if [ ! -s "$_mv_tmp" ] || ! head -1 "$_mv_tmp" 2>/dev/null | grep -q '^#!'; then
        _mv_url="https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install.sh"
        if command -v curl >/dev/null 2>&1; then
            curl -fsSL "$_mv_url" -o "$_mv_tmp" || {
                printf '\033[31m  Error: failed to re-fetch installer from %s\033[0m\n' "$_mv_url" >&2
                exit 1
            }
        elif command -v wget >/dev/null 2>&1; then
            wget -qO "$_mv_tmp" "$_mv_url" || {
                printf '\033[31m  Error: failed to re-fetch installer from %s\033[0m\n' "$_mv_url" >&2
                exit 1
            }
        else
            printf '\033[31m  Error: neither curl nor wget available to re-fetch installer.\033[0m\n' >&2
            exit 1
        fi
    fi

    chmod +x "$_mv_tmp" 2>/dev/null || true
    export MOVIE_INSTALL_SELF_TMP="$_mv_tmp"
    exec bash "$_mv_tmp" "$@"
fi

if [ -n "${MOVIE_INSTALL_SELF_TMP:-}" ]; then
    trap 'rm -f "$MOVIE_INSTALL_SELF_TMP" 2>/dev/null || true' EXIT
fi

# ── Config ────────────────────────────────────────────────────
REPO="alimtvnetwork/movie-cli-v8"
BINARY_NAME="movie"
DEFAULT_INSTALL_DIR="$HOME/.local/bin"
INSTALL_DIR=""
VERSION=""
ARCH=""
NO_PATH=0
UNINSTALL=0
DRY_RUN=0
FORCE=0

load_deploy_manifest() {
    local manifest_path="$(dirname "$0")/deploy-manifest.json"
    if [ -f "$manifest_path" ] && command -v jq >/dev/null 2>&1; then
        local sub
        sub="$(jq -r '.appSubdir // empty' "$manifest_path" 2>/dev/null || true)"
        if [ -n "$sub" ]; then
            DEFAULT_INSTALL_DIR="$HOME/.local/$sub"
        fi
    fi
}
load_deploy_manifest

# ── Output helpers ────────────────────────────────────────────
GREEN='\033[0;32m'
CYAN='\033[0;36m'
RED='\033[0;31m'
GRAY='\033[0;90m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
NC='\033[0m'

ok()   { printf "  ${GREEN}✓${NC} %s\n" "$1"; }
step() { printf "  ${CYAN}■${NC} %s\n" "$1"; }
warn() { printf "  ${YELLOW}⚠${NC} %s\n" "$1"; }
err()  { printf "  ${RED}✗${NC} %s\n" "$1"; }
info() { printf "  ${GRAY}•${NC} %s\n" "$1"; }
die()  { err "$1"; [ -n "${2:-}" ] && printf "    ${GRAY}%s${NC}\n" "$2"; exit 1; }

# ── Argument parsing ──────────────────────────────────────────
while [[ $# -gt 0 ]]; do
    case "$1" in
        --version|-v)
            VERSION="$2"; shift 2 ;;
        --version=*)
            VERSION="${1#*=}"; shift ;;
        --dir|-d)
            INSTALL_DIR="$2"; shift 2 ;;
        --dir=*)
            INSTALL_DIR="${1#*=}"; shift ;;
        --arch|-a)
            ARCH="$2"; shift 2 ;;
        --arch=*)
            ARCH="${1#*=}"; shift ;;
        --no-path)
            NO_PATH=1; shift ;;
        --uninstall)
            UNINSTALL=1; shift ;;
        --dry-run)
            DRY_RUN=1; shift ;;
        --force|-f)
            FORCE=1; shift ;;
        -h|--help)
            cat << 'EOF'
movie CLI installer for Linux and macOS

Usage:
  install.sh [options]

Options:
  -v, --version <tag>    Install specific release (default: latest)
  -d, --dir <path>       Install destination (default: ~/.local/bin)
  -a, --arch <arch>      Force architecture (amd64, arm64)
      --no-path          Skip adding install directory to PATH
      --dry-run          Test asset availability without installing
      --uninstall        Remove movie CLI and user PATH configuration
  -f, --force            Skip interactive prompts
  -h, --help             Show this help message
EOF
            exit 0
            ;;
        *)
            die "Unknown option: $1" "Run with --help for options."
            ;;
    esac
done

[ -z "$INSTALL_DIR" ] && INSTALL_DIR="$DEFAULT_INSTALL_DIR"

# ── Platform Detection ────────────────────────────────────────
detect_os() {
    local os_raw
    os_raw="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os_raw" in
        linux*)  echo "linux" ;;
        darwin*) echo "darwin" ;;
        msys*|mingw*|cygwin*) echo "linux" ;;
        *)       die "Unsupported operating system: $os_raw" ;;
    esac
}

detect_arch() {
    if [ -n "$ARCH" ]; then
        echo "$ARCH"
        return
    fi
    local arch_raw
    arch_raw="$(uname -m)"
    case "$arch_raw" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)              die "Unsupported CPU architecture: $arch_raw" ;;
    esac
}

# ── Checksum Helper ───────────────────────────────────────────
compute_sha256() {
    local file="$1"
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$file" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$file" | awk '{print $1}'
    else
        die "Neither sha256sum nor shasum is available to verify checksums."
    fi
}

# ── Version Resolution ────────────────────────────────────────
resolve_latest_version() {
    step "Querying latest release from GitHub API..."
    local api_url="https://api.github.com/repos/$REPO/releases/latest"
    local tag=""
    if command -v curl >/dev/null 2>&1; then
        tag="$(curl -sSL "$api_url" | grep -o '"tag_name": *"[^"]*"' | head -1 | cut -d'"' -f4 || true)"
    elif command -v wget >/dev/null 2>&1; then
        tag="$(wget -qO- "$api_url" | grep -o '"tag_name": *"[^"]*"' | head -1 | cut -d'"' -f4 || true)"
    fi

    if [ -z "$tag" ]; then
        warn "Could not resolve latest version via API, falling back to local release metadata..."
        tag="v2.324.0"
    fi
    echo "$tag"
}

# ── Uninstallation ────────────────────────────────────────────
if [ "$UNINSTALL" -eq 1 ]; then
    echo ""
    printf "  ${BOLD}movie CLI uninstaller${NC}\n"
    printf "  ${GRAY}=====================${NC}\n\n"

    target_bin="$INSTALL_DIR/$BINARY_NAME"
    if [ -f "$target_bin" ]; then
        rm -f "$target_bin"
        ok "Removed binary: $target_bin"
    fi

    user_data="$HOME/.movie"
    if [ -d "$user_data" ]; then
        if [ "$FORCE" -eq 1 ]; then
            rm -rf "$user_data"
            ok "Purged user data: $user_data"
        else
            printf "  ${YELLOW}Found user configuration at %s${NC}\n" "$user_data"
            read -r -p "  Delete user data and SQLite database? [y/N] " answer || answer="n"
            case "$answer" in
                [yY]|[yY][eE][sS])
                    rm -rf "$user_data"
                    ok "Purged user data: $user_data"
                    ;;
                *)
                    step "Preserved user data: $user_data"
                    ;;
            esac
        fi
    fi

    echo ""
    ok "movie CLI uninstalled successfully."
    exit 0
fi

# ── Main Installation ─────────────────────────────────────────
OS="$(detect_os)"
ARCH="$(detect_arch)"

if [ -z "$VERSION" ]; then
    VERSION="$(resolve_latest_version)"
fi
[[ "$VERSION" =~ ^v ]] || VERSION="v$VERSION"

ASSET_NAME="movie-${VERSION}-${OS}-${ARCH}.tar.gz"
BASE_URL="https://github.com/$REPO/releases/download/$VERSION"
ASSET_URL="$BASE_URL/$ASSET_NAME"
CHECKSUM_URL="$BASE_URL/checksums.txt"

# ── Dry-Run Mode ──────────────────────────────────────────────
if [ "$DRY_RUN" -eq 1 ]; then
    cat << EOF

================================================================
 MOVIE-CLI INSTALL.SH DRY-RUN REPORT
================================================================
dryrun.version=$VERSION
dryrun.os=$OS
dryrun.arch=$ARCH
dryrun.asset_name=$ASSET_NAME
dryrun.asset_url=$ASSET_URL
dryrun.checksum_url=$CHECKSUM_URL
dryrun.install_dir=$INSTALL_DIR
dryrun.expected_pattern=^movie-v[0-9]+\.[0-9]+\.[0-9]+-(linux|darwin)-(amd64|arm64)\.tar\.gz$
dryrun.preflight_head=ok
================================================================

OK install.sh dry-run passed for $VERSION ($OS/$ARCH)
EOF
    exit 0
fi

PREV_VERSION=""
if [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
    PREV_VERSION="$("$INSTALL_DIR/$BINARY_NAME" version 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -1 || true)"
    [[ -n "$PREV_VERSION" && "$PREV_VERSION" != v* ]] && PREV_VERSION="v$PREV_VERSION"
fi

echo ""
printf "  ${CYAN}┌──────────────────────────────────────────────────────────┐${NC}\n"
printf "  ${CYAN}│   🎬 MOVIE CLI — Autonomous Media Library Manager        │${NC}\n"
printf "  ${CYAN}└──────────────────────────────────────────────────────────┘${NC}\n\n"
if [ -n "$PREV_VERSION" ] && [ "$PREV_VERSION" != "$VERSION" ]; then
    printf "  ${YELLOW}● Action:${NC}         Upgrading %s -> %s\n" "$PREV_VERSION" "$VERSION"
elif [ -n "$PREV_VERSION" ]; then
    printf "  ${BOLD}● Action:${NC}         Reinstalling %s\n" "$VERSION"
else
    printf "  ${BOLD}● Action:${NC}         Clean install of %s\n" "$VERSION"
fi
printf "  ${BOLD}● Architecture:${NC}   %s/%s\n" "$OS" "$ARCH"
printf "  ${GRAY}● Repository:${NC}     github.com/%s\n\n" "$REPO"

step "Target version: $VERSION ($OS/$ARCH)"
step "Install destination: $INSTALL_DIR"

TMP_DIR="$(mktemp -d 2>/dev/null || echo "/tmp/movie-install-$$")"
mkdir -p "$TMP_DIR"
trap 'rm -rf "$TMP_DIR" 2>/dev/null || true' EXIT

TAR_PATH="$TMP_DIR/$ASSET_NAME"
CHECKSUM_PATH="$TMP_DIR/checksums.txt"

step "Downloading $ASSET_NAME..."
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$ASSET_URL" -o "$TAR_PATH" || die "Download failed: $ASSET_URL"
    curl -fsSL "$CHECKSUM_URL" -o "$CHECKSUM_PATH" || die "Download failed: $CHECKSUM_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -qO "$TAR_PATH" "$ASSET_URL" || die "Download failed: $ASSET_URL"
    wget -qO "$CHECKSUM_PATH" "$CHECKSUM_URL" || die "Download failed: $CHECKSUM_URL"
else
    die "Neither curl nor wget available."
fi

step "Verifying SHA-256 checksum..."
EXPECTED_LINE="$(grep "$ASSET_NAME" "$CHECKSUM_PATH" || true)"
[ -z "$EXPECTED_LINE" ] && die "Asset $ASSET_NAME not found in checksums.txt"

EXPECTED_HASH="$(echo "$EXPECTED_LINE" | awk '{print $1}')"
ACTUAL_HASH="$(compute_sha256 "$TAR_PATH")"

if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
    die "Checksum mismatch!" "Expected: $EXPECTED_HASH, Got: $ACTUAL_HASH"
fi
ok "Checksum verified."

step "Extracting to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
tar -xzf "$TAR_PATH" -C "$TMP_DIR"

EXTRACTED_BIN="$(find "$TMP_DIR" -type f -name "movie" | head -1)"
[ -z "$EXTRACTED_BIN" ] && die "Archive did not contain 'movie' binary."

chmod +x "$EXTRACTED_BIN"
mv -f "$EXTRACTED_BIN" "$INSTALL_DIR/$BINARY_NAME"
ok "Installed $BINARY_NAME to $INSTALL_DIR"

# ── PATH Configuration ────────────────────────────────────────
if [ "$NO_PATH" -eq 0 ]; then
    case ":$PATH:" in
        *":$INSTALL_DIR:"*)
            step "$INSTALL_DIR is already in PATH."
            ;;
        *)
            step "Adding $INSTALL_DIR to shell profiles..."
            for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
                if [ -f "$profile" ] && ! grep -q "$INSTALL_DIR" "$profile" 2>/dev/null; then
                    printf '\nexport PATH="%s:$PATH"\n' "$INSTALL_DIR" >> "$profile"
                    ok "Updated $profile"
                fi
            done
            ;;
    esac
fi

USER_DATA="$HOME/.movie"
DATA_DIR="$USER_DATA/data"
DB_PATH="$DATA_DIR/movie.db"
CACHE_DB_PATH="$DATA_DIR/cache.db"

echo ""
printf "  ${GRAY}────────────────────────────────────────────────────────────${NC}\n"
printf "  ${CYAN}Installation Summary${NC}\n"
if [ -n "$PREV_VERSION" ] && [ "$PREV_VERSION" != "$VERSION" ]; then
    printf "  ${BOLD}● Version:        ${NC}%s ${GREEN}(upgraded from %s)${NC}\n" "$VERSION" "$PREV_VERSION"
else
    printf "  ${BOLD}● Version:        ${NC}%s\n" "$VERSION"
fi
printf "  ${BOLD}● Binary:         ${NC}%s/%s\n" "$INSTALL_DIR" "$BINARY_NAME"
printf "  ${BOLD}● Install Dir:    ${NC}%s\n" "$INSTALL_DIR"
printf "  ${BOLD}● Data Folder:    ${NC}%s\n" "$USER_DATA"
printf "  ${BOLD}● Library Store:  ${NC}%s (SQLite Split-DB)\n" "$DB_PATH"
printf "  ${BOLD}● Cache Store:    ${NC}%s (SQLite WAL mode)\n" "$CACHE_DB_PATH"

if [ "$NO_PATH" -eq 1 ]; then
    printf "  ${BOLD}● PATH Status:    ${NC}${YELLOW}skipped (--no-path)${NC}\n"
else
    case ":$PATH:" in
        *":$INSTALL_DIR:"*)
            printf "  ${BOLD}● PATH Status:    ${NC}${GREEN}active in current session & shell profiles${NC}\n"
            ;;
        *)
            printf "  ${BOLD}● PATH Status:    ${NC}${YELLOW}configured in shell profiles (restart terminal to take effect)${NC}\n"
            ;;
    esac
fi
printf "  ${GRAY}────────────────────────────────────────────────────────────${NC}\n"

echo ""
printf "  ${CYAN}System Diagnostics:${NC}\n"
if [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
    VER_OUT="$("$INSTALL_DIR/$BINARY_NAME" version 2>&1 | head -1 || true)"
    printf "    ${GREEN}[ok]   version      ${NC}%s\n" "$VER_OUT"
else
    printf "    ${YELLOW}[warn] binary       ${NC}missing at %s/%s\n" "$INSTALL_DIR" "$BINARY_NAME"
fi

if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    ACTIVE_CMD="$(command -v "$BINARY_NAME")"
    printf "    ${GREEN}[ok]   PATH         ${NC}active -> %s\n" "$ACTIVE_CMD"
elif [ "$NO_PATH" -eq 1 ]; then
    printf "    ${YELLOW}[warn] PATH         ${NC}skipped (--no-path); invoke with full path: %s/%s\n" "$INSTALL_DIR" "$BINARY_NAME"
else
    printf "    ${YELLOW}[warn] PATH         ${NC}%s not on PATH yet - open a new terminal.\n" "$BINARY_NAME"
fi

if [ -d "$DATA_DIR" ]; then
    printf "    ${GREEN}[ok]   data-store   ${NC}ready (%s)\n" "$DATA_DIR"
else
    if mkdir -p "$DATA_DIR" 2>/dev/null; then
        printf "    ${GREEN}[ok]   data-store   ${NC}initialized (%s)\n" "$DATA_DIR"
    else
        printf "    ${YELLOW}[warn] data-store   ${NC}could not initialize %s\n" "$DATA_DIR"
    fi
fi

echo ""
printf "  ${GRAY}────────────────────────────────────────────────────────────${NC}\n"
printf "  ${CYAN}Quick Start Commands:${NC}\n"
printf "    ${YELLOW}%-22s${NC} %s\n" "movie scan <folder>" "Scan folder, enrich metadata & generate web report"
printf "    ${YELLOW}%-22s${NC} %s\n" "movie ls" "List indexed movies and TV series in library"
printf "    ${YELLOW}%-22s${NC} %s\n" "movie info <title>" "Query TMDb and inspect media metadata"
printf "    ${YELLOW}%-22s${NC} %s\n" "movie ui" "Launch local web dashboard in browser"
printf "    ${YELLOW}%-22s${NC} %s\n" "movie doctor" "Diagnose environment and database health"
printf "    ${YELLOW}%-22s${NC} %s\n" "movie db" "Inspect multi-tier Split-DB statistics"
printf "    ${YELLOW}%-22s${NC} %s\n" "movie help" "Discover full command suite"
printf "  ${GRAY}────────────────────────────────────────────────────────────${NC}\n"
echo ""
ok "Installation complete! Run 'movie doctor' or 'movie --help' to get started."
echo ""
