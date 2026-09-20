#!/usr/bin/env bash
# Short interactive installer for movie CLI on Linux / macOS.
#
# DUAL-MODE EXECUTION (Modeled after GitMap Spec 108):
#
#   1. Eval-mode (RECOMMENDED — auto-activates PATH in current shell):
#        eval "$(curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install-quick.sh)"
#
#      Runs directly inside the interactive shell so the trailing source
#      mutates the current PATH immediately. No "open a new terminal" needed.
#
#   2. Pipe-mode (LEGACY — child process, prints reload banner):
#        curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/movie-cli-v8/main/install-quick.sh | bash
#
#   3. Local-file mode:
#        ./install-quick.sh
#        ./install-quick.sh --dir ~/.local/bin
#        ./install-quick.sh --no-discovery

REPO="alimtvnetwork/movie-cli-v8"
INSTALLER_URL="https://raw.githubusercontent.com/${REPO}/main/install.sh"
if [ "$(id -u 2>/dev/null || echo 1)" -eq 0 ]; then
    DEFAULT_DIR="/usr/local/bin"
else
    DEFAULT_DIR="${HOME:-~}/.local/bin"
fi

cleanup_corrupted_install_dirs() {
    local scan_roots=()
    [ -n "${PWD:-}" ] && scan_roots+=("${PWD}")
    [ -n "${HOME:-}" ] && scan_roots+=("${HOME}")
    [ -d "${HOME:-}/.local" ] && scan_roots+=("${HOME}/.local" "${HOME}/.local/bin")
    [ -d "/tmp" ] && scan_roots+=("/tmp")

    if command -v python3 >/dev/null 2>&1; then
        python3 - "${scan_roots[@]}" << 'PYEOF' 2>/dev/null || true
import os, sys, shutil

esc = chr(27)
for root in sys.argv[1:]:
    if not os.path.isdir(root):
        continue
    try:
        for name in os.listdir(root):
            p = os.path.join(root, name)
            if not os.path.isdir(p) or os.path.islink(p):
                continue
            if p in ("/", os.path.expanduser("~")):
                continue

            is_bad = False
            if name == "~":
                is_bad = True
            elif esc in name or "\x1b" in name or "\033" in name:
                is_bad = True
            elif "\n" in name or "\r" in name:
                is_bad = True
            elif "quick installer" in name.lower() or ("movie" in name.lower() and "installer" in name.lower()):
                is_bad = True

            if is_bad:
                shutil.rmtree(p, ignore_errors=True)
    except Exception:
        pass
PYEOF
        return 0
    fi
}

sanitize_install_dir() {
    local raw="$1"
    if [ -z "${raw}" ]; then
        echo ""
        return 0
    fi

    local cleaned
    cleaned="$(printf '%s' "${raw}" | sed -E 's/\x1B\[[0-9;]*[a-zA-Z]//g' | tr -d '\r')"
    cleaned="$(printf '%s' "${cleaned}" | sed -e 's/^[[:space:]"'"'"']*//' -e 's/[[:space:]"'"'"']*$//')"

    case "${cleaned}" in
        *$'\n'*|*$'\r'*|*"installer"*|*"Install path"*|*"Default:"*)
            echo ""
            return 0
            ;;
    esac

    if [ "${cleaned}" = "~" ]; then
        cleaned="${HOME:-/tmp}"
    elif [[ "${cleaned}" == "~/"* ]]; then
        cleaned="${HOME:-/tmp}/${cleaned#\~/}"
    fi

    if [ "${cleaned}" != "/" ]; then
        cleaned="${cleaned%/}"
    fi

    echo "${cleaned}"
}

__movie_detect_eval_mode() {
    if [ -n "${BASH_EXECUTION_STRING:-}" ]; then
        echo 0
        return
    fi
    local shell_name
    shell_name="$(basename "${SHELL:-/bin/bash}")"
    if [ "$0" = "${shell_name}" ] || [ "$0" = "-${shell_name}" ] || [ -z "${BASH_SOURCE[0]:-}" ]; then
        echo 1
        return
    fi
    echo 0
}

resolve_effective_repo() {
    local base_repo="$1"
    local window="${2:-20}"

    if [[ ! "$base_repo" =~ ^([^/]+)/(.+)-v([0-9]+)$ ]]; then
        echo "$base_repo"
        return
    fi

    local owner="${BASH_REMATCH[1]}"
    local stem="${BASH_REMATCH[2]}"
    local baseline="${BASH_REMATCH[3]}"
    local effective="$baseline"

    local max_m=$((baseline + window))
    for ((m = baseline + 1; m <= max_m; m++)); do
        local probe_url="https://github.com/$owner/$stem-v$m"
        if curl -fsIL --connect-timeout 2 -m 4 "$probe_url" >/dev/null 2>&1; then
            effective="$m"
        fi
    done

    if [ "$effective" -eq "$baseline" ]; then
        echo "$base_repo"
    else
        echo "$owner/$stem-v$effective"
    fi
}

__movie_quick_install_main() {
    set -euo pipefail

    local install_dir="$DEFAULT_DIR"
    local version=""
    local no_discovery=false
    local extra_args=()

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dir|-d)
                install_dir="$2"; shift 2 ;;
            --dir=*)
                install_dir="${1#*=}"; shift ;;
            --version|-v)
                version="$2"; shift 2 ;;
            --version=*)
                version="${1#*=}"; shift ;;
            --no-discovery)
                no_discovery=true; shift ;;
            *)
                extra_args+=("$1"); shift ;;
        esac
    done

    cleanup_corrupted_install_dirs
    install_dir="$(sanitize_install_dir "$install_dir")"
    install_dir="${install_dir:-$DEFAULT_DIR}"

    local effective_repo="$REPO"
    if [ "$no_discovery" = "false" ] && [ -z "$version" ]; then
        effective_repo="$(resolve_effective_repo "$REPO" 20)"
    fi

    local installer_url="https://raw.githubusercontent.com/${effective_repo}/main/install.sh"
    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" 2>/dev/null && pwd || echo "")"
    local local_installer="$script_dir/install.sh"

    mkdir -p "$install_dir"
    local profile_hint_file="$install_dir/.movie-last-profile"
    rm -f "$profile_hint_file"

    if [ -f "$local_installer" ] && [ "$effective_repo" = "$REPO" ]; then
        bash "$local_installer" --dir "$install_dir" ${version:+--version "$version"} "${extra_args[@]}"
    else
        printf "  \033[36m→\033[0m Fetching installer from %s...\n" "$installer_url"
        curl -fsSL "$installer_url" | bash -s -- --dir "$install_dir" ${version:+--version "$version"} "${extra_args[@]}"
    fi

    local shell_rc
    case "$(basename "${SHELL:-/bin/bash}")" in
        zsh)  shell_rc="$HOME/.zshrc" ;;
        fish) shell_rc="$HOME/.config/fish/config.fish" ;;
        *)    shell_rc="$HOME/.bashrc" ;;
    esac
    echo "$shell_rc" > "$profile_hint_file"
}

# ── Entrypoint Dispatcher ────────────────────────────────────────────────
__movie_eval_mode="$(__movie_detect_eval_mode)"

if [ "$__movie_eval_mode" -eq 1 ]; then
    # Eval mode: execute inside a subshell to avoid polluting user shell state
    ( __movie_quick_install_main "$@" )
    __movie_rc_code=$?
    if [ $__movie_rc_code -eq 0 ]; then
        # Auto-source the resolved profile in the current interactive session
        __movie_hint_file="${DEFAULT_DIR}/.movie-last-profile"
        if [ -f "$__movie_hint_file" ]; then
            __movie_prof="$(cat "$__movie_hint_file" 2>/dev/null || echo "")"
            if [ -n "$__movie_prof" ] && [ -f "$__movie_prof" ]; then
                # shellcheck source=/dev/null
                . "$__movie_prof" 2>/dev/null || true
            fi
            rm -f "$__movie_hint_file"
        fi
        if command -v movie >/dev/null 2>&1; then
            printf "\n  \033[1;32m✓\033[0m \033[1mmovie\033[0m activated on PATH for this shell: %s\n\n" "$(command -v movie)"
        fi
    fi
else
    # Pipe mode / local execution
    __movie_quick_install_main "$@"
fi
