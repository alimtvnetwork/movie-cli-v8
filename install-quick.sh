#!/usr/bin/env bash
# install-quick.sh — Short installer for movie CLI on Linux and macOS.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install-quick.sh | bash
#   ./install-quick.sh
#   ./install-quick.sh --dir ~/.local/bin --version v2.324.0

set -euo pipefail

REPO="alimtvnetwork/movie-cli-v8"
INSTALLER_URL="https://raw.githubusercontent.com/${REPO}/main/install.sh"

if [ "$(id -u 2>/dev/null || echo 1)" -eq 0 ]; then
    DEFAULT_DIR="/usr/local/bin"
else
    DEFAULT_DIR="${HOME:-~}/.local/bin"
fi

INSTALL_DIR="$DEFAULT_DIR"
VERSION=""
EXTRA_ARGS=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dir|-d)
            INSTALL_DIR="$2"; shift 2 ;;
        --dir=*)
            INSTALL_DIR="${1#*=}"; shift ;;
        --version|-v)
            VERSION="$2"; shift 2 ;;
        --version=*)
            VERSION="${1#*=}"; shift ;;
        *)
            EXTRA_ARGS+=("$1"); shift ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd || echo "")"
LOCAL_INSTALLER="$SCRIPT_DIR/install.sh"

if [ -f "$LOCAL_INSTALLER" ]; then
    bash "$LOCAL_INSTALLER" --dir "$INSTALL_DIR" ${VERSION:+--version "$VERSION"} "${EXTRA_ARGS[@]}"
else
    printf "  \033[36m→\033[0m Fetching canonical installer from %s...\n" "$INSTALLER_URL"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$INSTALLER_URL" | bash -s -- --dir "$INSTALL_DIR" ${VERSION:+--version "$VERSION"} "${EXTRA_ARGS[@]}"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$INSTALLER_URL" | bash -s -- --dir "$INSTALL_DIR" ${VERSION:+--version "$VERSION"} "${EXTRA_ARGS[@]}"
    else
        printf "\033[31m  Error: curl or wget required.\033[0m\n" >&2
        exit 1
    fi
fi
