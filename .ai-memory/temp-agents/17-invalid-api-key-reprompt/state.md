# State: TMDb API Key Re-Prompt & Minor Release (17-invalid-api-key-reprompt)

- Task: 17-invalid-api-key-reprompt-and-release
- Target Version: v2.328.0
- Status: READY_FOR_RELEASE
- Progress: 90%

## Target Deliverables
1. [x] Create master plan `.ai-memory/plans/pending/17-invalid-api-key-reprompt-and-release.md`.
2. [x] Add `VerifyAuth() error` to `tmdb.Client`.
3. [x] Implement interactive verification and re-prompt loop in `cmd/movie_tmdb.go`.
4. [x] Implement mid-scan auth failure recovery and error suppression in `cmd/movie_scan_process.go`.
5. [x] Wire `ensureValidTmdbClient` into `search`, `suggest`, `discover`, `info`.
6. [x] Add unit test for `VerifyAuth` with mock HTTP server.
7. [x] Pass all 20 quality gates in `06-cicd-local-runner.py --all-paths --run-tests`.
8. [ ] Execute 5-step release branching lifecycle for `v2.328.0` and monitor remote pipeline with GitMap dynamic waiting.
