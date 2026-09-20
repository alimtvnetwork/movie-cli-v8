#!/usr/bin/env bash
#
# Build, deploy, and run movie CLI from the repo root.
#
# Usage:
#   ./run.sh                          # pull, build, deploy
#   ./run.sh --no-pull                # skip git pull
#   ./run.sh --force-pull             # discard local changes + pull (no prompt)
#   ./run.sh --no-deploy              # skip deploy step
#   ./run.sh -r scan                  # build + scan parent folder
#   ./run.sh -r scan ~/movies         # build + scan specific path
#   ./run.sh -r help                  # build + show help
#   ./run.sh -t                       # run all unit tests with reports

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")" && pwd)"
BIN_DIR="$REPO_ROOT/bin"
BIN_NAME="movie"

# -- Defaults --------------------------------------------------
NO_PULL=false
NO_DEPLOY=false
FORCE_PULL=false
DEPLOY_PATH=""
RUN=false
TEST=false
RUN_ARGS=()

# -- Parse arguments -------------------------------------------
while [[ $# -gt 0 ]]; do
    case "$1" in
        --no-pull)             NO_PULL=true; shift ;;
        --force-pull)          FORCE_PULL=true; shift ;;
        --no-deploy)           NO_DEPLOY=true; shift ;;
        --deploy-path)         DEPLOY_PATH="$2"; shift 2 ;;
        -r|--run)              RUN=true; shift
            while [[ $# -gt 0 ]]; do
                RUN_ARGS+=("$1"); shift
            done
            ;;
        -t|--test)             TEST=true; shift ;;
        *)
            echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

step() {
    printf "\n\033[35m [%s]\033[0m \033[1m%s\033[0m\n" "$1" "$2"
}

ok() {
    printf "  \033[32m✔ %s\033[0m\n" "$1"
}

# 1. Pull
if [ "$NO_PULL" = false ]; then
    step "git" "Updating repository..."
    if [ "$FORCE_PULL" = true ]; then
        git reset --hard HEAD
        git clean -fd
    fi
    git pull --ff-only || true
    ok "Git up to date"
fi

# 2. Test
if [ "$TEST" = true ]; then
    step "test" "Running unit tests..."
    go test ./...
    ok "Tests passed"
fi

# 3. Build
step "build" "Compiling $BIN_NAME..."
mkdir -p "$BIN_DIR"
BUILD_COMMIT="$(git rev-parse HEAD 2>/dev/null || echo "unknown")"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
go build -ldflags "-s -w -X 'github.com/alimtvnetwork/movie-cli-v8/version.GitCommit=${BUILD_COMMIT}' -X 'github.com/alimtvnetwork/movie-cli-v8/version.BuildDate=${BUILD_DATE}'" -o "$BIN_DIR/$BIN_NAME" .
ok "Built $BIN_DIR/$BIN_NAME"

# 4. Deploy
if [ "$NO_DEPLOY" = false ]; then
    if [ -z "$DEPLOY_PATH" ]; then
        DEPLOY_PATH="${HOME}/.local/bin"
    fi
    step "deploy" "Deploying to $DEPLOY_PATH..."
    mkdir -p "$DEPLOY_PATH"
    cp -f "$BIN_DIR/$BIN_NAME" "$DEPLOY_PATH/$BIN_NAME"
    chmod +x "$DEPLOY_PATH/$BIN_NAME"
    ok "Deployed to $DEPLOY_PATH/$BIN_NAME"
fi

# 5. Run
if [ "$RUN" = true ]; then
    step "run" "Executing $BIN_NAME ${RUN_ARGS[*]}..."
    "$BIN_DIR/$BIN_NAME" "${RUN_ARGS[@]}"
fi
