#!/usr/bin/env bash
# ----------------------------------------------------------------------
# init.sh - one-shot repo init & preflight validation for movie CLI
# ----------------------------------------------------------------------

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DRY_RUN=0

print_help() {
  cat <<'EOF'
init.sh - One-shot repo init & preflight validation for movie-cli.

Usage:
  ./init.sh             # run toolchain check and installer dry-run validation
  ./init.sh --dry-run   # preview steps without environment impact
  ./init.sh -h | --help
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) print_help; exit 0 ;;
    *) echo "init: ERROR unknown flag '$1'" >&2; exit 1 ;;
  esac
done

echo
echo "==> [movie-cli init] Starting preflight validation..."

# Step 1: Deploy manifest
if [ -f "$SCRIPT_DIR/deploy-manifest.json" ]; then
  echo "  [OK] deploy-manifest.json found."
else
  echo "  [WARN] deploy-manifest.json missing."
fi

# Step 2: Go toolchain
if command -v go >/dev/null 2>&1; then
  echo "  [OK] $(go version)"
else
  echo "  [WARN] 'go' command not found on PATH."
fi

# Step 3: Installer dry-run
if [ -f "$SCRIPT_DIR/install.sh" ]; then
  echo "  [TEST] Running install.sh --dry-run..."
  bash "$SCRIPT_DIR/install.sh" --dry-run
  rc=$?
  if [ $rc -eq 0 ]; then
    echo "  [OK] install.sh dry-run contract passed."
  else
    echo "  [FAIL] install.sh dry-run failed with exit code $rc" >&2
    exit $rc
  fi
fi

echo
echo "==> [movie-cli init] Preflight validation completed successfully."
exit 0
