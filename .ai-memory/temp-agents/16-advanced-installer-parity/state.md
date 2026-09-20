# State: Advanced Installer & Release Process Parity (16-advanced-installer-parity)

- Task: 16-advanced-installer-and-release-process-parity
- Target Version: v2.327.0
- Status: READY_FOR_RELEASE
- Progress: 90%

## Target Deliverables
1. [x] Create master plan `.ai-memory/plans/pending/16-advanced-installer-and-release-process-parity.md` with verbatim prompt and actionable checklist.
2. [x] Decompose into lean subtasks under `.ai-memory/plans/subtasks/16-advanced-installer-and-release-process-parity/`.
3. [x] Implement Dual-Mode execution (Eval-mode auto-source + Pipe-mode) in `install-quick.sh` (GitMap Spec 108).
4. [x] Implement Versioned Repo Discovery in `install-quick.ps1` and `install-quick.sh` (GitMap Spec 95).
5. [x] Implement Deploy Path persistence (`deployPath` in `powershell.json` / `movie.json`) in `install-quick.ps1`.
6. [x] Implement `movie self-uninstall` (and `movie uninstall` alias) in Go CLI with Windows temp-handoff self-deletion (GitMap Spec 90).
7. [x] Refactor `03-ai-scripts/16-installer-smoke-tester.py` to differentiate root installers vs quick wrappers vs uninstallers, and register in `02-shared-engine.py`.
8. [x] Update `readme.md` with dual-mode installation one-liners.
9. [x] Verify all quality gates pass 100% green (`06-cicd-local-runner.py --all-paths --run-tests`).
10. [ ] Execute 5-step release branching lifecycle for `v2.327.0`, monitor remote pipeline via GitMap dynamic waiting, and consolidate plan.
