#!/usr/bin/env bash
# uninstall-quick.sh — One-liner uninstaller for movie CLI on Linux and macOS.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/uninstall-quick.sh | bash
#   ./uninstall-quick.sh
#   ./uninstall-quick.sh --yes --purge-data

set -euo pipefail

INSTALL_DIR="${HOME:-~}/.local/bin"
BINARY_NAME="movie"
KEEP_DATA=0
PURGE_DATA=0
YES=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dir|-d)
            INSTALL_DIR="$2"; shift 2 ;;
        --keep-data)
            KEEP_DATA=1; shift ;;
        --purge-data)
            PURGE_DATA=1; shift ;;
        --yes|-y)
            YES=1; shift ;;
        *)
            shift ;;
    esac
done

GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
NC='\033[0m'

echo ""
printf "  ${BOLD}movie CLI Quick Uninstaller${NC}\n"
printf "  ===========================\n\n"

TARGET_BIN="$INSTALL_DIR/$BINARY_NAME"
if [ -f "$TARGET_BIN" ]; then
    rm -f "$TARGET_BIN"
    printf "  ${GREEN}✓${NC} Removed binary: %s\n" "$TARGET_BIN"
else
    # Also check /usr/local/bin
    if [ -f "/usr/local/bin/$BINARY_NAME" ]; then
        rm -f "/usr/local/bin/$BINARY_NAME" 2>/dev/null || true
        printf "  ${GREEN}✓${NC} Removed binary: /usr/local/bin/%s\n" "$BINARY_NAME"
    fi
fi

USER_DATA="${HOME:-~}/.movie"
if [ -d "$USER_DATA" ]; then
    if [ "$PURGE_DATA" -eq 1 ] || { [ "$YES" -eq 1 ] && [ "$KEEP_DATA" -eq 0 ]; }; then
        rm -rf "$USER_DATA"
        printf "  ${GREEN}✓${NC} Removed user data: %s\n" "$USER_DATA"
    elif [ "$KEEP_DATA" -eq 1 ]; then
        printf "  ${CYAN}→${NC} Preserved user data: %s\n" "$USER_DATA"
    else
        printf "  ${YELLOW}Found user configuration & database at %s${NC}\n" "$USER_DATA"
        read -r -p "  Delete ~/.movie user data? [y/N] " answer || answer="n"
        case "$answer" in
            [yY]|[yY][eE][sS])
                rm -rf "$USER_DATA"
                printf "  ${GREEN}✓${NC} Removed user data: %s\n" "$USER_DATA"
                ;;
            *)
                printf "  ${CYAN}→${NC} Preserved user data: %s\n" "$USER_DATA"
                ;;
        esac
    fi
fi

echo ""
printf "  ${GREEN}✓${NC} Uninstall complete.\n\n"
