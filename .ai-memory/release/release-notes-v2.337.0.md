## Quick Install v2.337.0

### Windows (PowerShell)

```powershell
irm https://github.com/alimtvnetwork/movie-cli-v8/releases/download/v2.337.0/install.ps1 | iex
```

### Unix / Linux / macOS (Bash)

```bash
curl -fsSL https://github.com/alimtvnetwork/movie-cli-v8/releases/download/v2.337.0/install.sh | bash
```

---

## What's Changed in v2.337.0

### Added / Changed — GitMap CI/CD runner parity, zero-storage workflows, and canonical prompts/skills sync

- **GitMap CI/CD Runner Parity**: Added `run-smart`, `smart`, `--smart`, `run-incremental`, `incremental`, `--fast`, and `--commits` shortcuts with dynamic package discovery based on changed git files to `03-ai-scripts/06-cicd-local-runner.py`.
- **Zero-Storage GitHub Actions Workflows**: Fully eliminated `actions/upload-artifact@v4` across `.github/workflows/ci.yml` in compliance with Rule 7, rendering reports directly via `$GITHUB_STEP_SUMMARY` and console output, and added `.github/workflows/purge-actions-artifacts.yml`.
- **Canonical Prompts & Specs Synchronization**: Synced `01-prompts/` (105 files) and `02-spec/` (711 files) from coding guidelines while strictly preserving all domain-specific `movie-cli` specs (`21-app`, `22-app-issues`, `23-app-db`, `24-app-ui-design-system`, `19-main-worker-service/98-changelog.md`).
- **Antigravity Skills Fleet Expansion**: Synced 13 new GitMap & system skills while preserving all 10 `movie-cli-*` domain skills, and updated `03-ai-scripts/38-sync-prompts-skills-scripts.py`.
- **Code Quality & Struct Alignment**: Optimized `Client` struct memory layout in `tmdb/client.go` to 80 bytes via `fieldalignment`, fixed `gocritic` slice append and value-copy warnings, removed unused imports, and updated test inventory to 147 tests.
- **Clean CI Quality Gates**: All 15 local CI/CD quality gates verified 100% green.
