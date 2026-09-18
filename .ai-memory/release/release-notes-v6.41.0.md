## Quick Install v6.41.0

### Windows (PowerShell)

```powershell
Invoke-WebRequest -Uri https://raw.githubusercontent.com/alimtvnetwork/coding-guidelines-v24/v6.41.0/install.ps1 -OutFile install.ps1; .\install.ps1 -TargetDir ".ai-memory/prompts" -Version "v6.41.0"
```

### Unix / Linux / macOS (Bash)

```bash
curl -sL https://raw.githubusercontent.com/alimtvnetwork/coding-guidelines-v24/v6.41.0/install.sh | bash -s -- ".ai-memory/prompts" "v6.41.0"
```

---

## What's Changed in v6.41.0

### Added — Nuclear Package Modularization, 5-Day Test Inventory Freshness & Verbatim Task Extraction

- **Nuclear Package Modularization Prompt:** Added [`25-nuclear-package-modularization-and-unit-test-optimization.md`](https://github.com/alimtvnetwork/coding-guidelines-v24/blob/v6.41.0/01-prompts/15-cg-execute/25-nuclear-package-modularization-and-unit-test-optimization.md) mandating Directed Acyclic Graph (DAG) package architectures, zero-dependency leaf package segregation (`pkg/constants`, `pkg/model`, `pkg/appfault`, `pkg/fsutil`, `pkg/cliexit`), blackbox heavy test isolation (`tests/heavy_test/` under `package heavy_test`), and centralized `.ai-memory/test-inventory.json` synchronization.
- **5-Day Cache Freshness Engine:** Enhanced `03-ai-scripts/33-test-inventory-generator.py` with `--check-age` and `--max-age-days` to audit test inventory freshness and skip redundant test execution on fresh timing profiles.
- **Verbatim Task Capture Mandate:** Updated parent execution prompts and skills (`02-execute-parent-task-with-n-steps.md`, `06-execute-parent-task-with-n-steps-v2.md`, `07-execute-batched-loop-v2.md`) to require Phase 1 Step 0 verbatim prompt capture under `## User Request (Verbatim)` and actionable deliverables extraction under `## Extracted Actionable Task List`.
- **Per-Task Isolated Workspaces:** Enforced `.ai-memory/temp-agents/xx-<task-name>/` with progress tracking via `state.md`.
- **Catalog & Index Synchronization:** Registered Prompt 25 in `01-prompts/15-cg-execute/01-index.md` and regenerated `.ai-memory/prompts.md` with 96 indexed prompts.
- **Clean CI Quality Gates:** All 36 local CI/CD quality gates verified 100% green.
