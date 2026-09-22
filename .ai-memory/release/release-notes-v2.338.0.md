## Quick Install v2.338.0

### Windows (PowerShell)

```powershell
irm https://github.com/alimtvnetwork/movie-cli-v8/releases/download/v2.338.0/install.ps1 | iex
```

### Unix / Linux / macOS (Bash)

```bash
curl -fsSL https://github.com/alimtvnetwork/movie-cli-v8/releases/download/v2.338.0/install.sh | bash
```

---

## What's Changed in v2.338.0

### Added / Changed — Automated gofmt formatting on bump, release notes generator, and GitMap database ignore

- **Automated `gofmt` Formatting in Bump Lifecycle**: Enhanced `03-ai-scripts/37-bump-version.py` to automatically execute `gofmt -w` on `version/info.go` after writing, ensuring zero formatting drift or CI linter failures on version bump commits.
- **Dedicated Release Notes Generator**: Added automated `.ai-memory/release/release-notes-vX.Y.Z.md` generation with prominent PowerShell (`irm ... | iex`) and Bash (`curl -fsSL ... | bash`) Quick Install one-liners and structured changelog sections.
- **Release Orchestrator Boolean & Path Hygiene**: Fixed boolean polarity checks and updated release candidate staging in `03-ai-scripts/29-release-orchestrator.py` to target `.ai-memory/release/`.
- **GitMap Telemetry Hygiene**: Added `.gitmap/data/` to `.gitignore` to prevent local GitMap SQLite database caches (`sql.db`) from appearing in git status or being tracked.
- **Test Inventory Synchronization**: Updated test inventory timestamps across 147 cataloged tests.
- **Clean CI Quality Gates**: All 15 local CI/CD quality gates verified 100% green.
