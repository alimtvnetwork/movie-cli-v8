## Quick Install v2.339.0

### Windows (PowerShell)

```powershell
irm https://github.com/alimtvnetwork/movie-cli-v8/releases/download/v2.339.0/install.ps1 | iex
```

### Unix / Linux / macOS (Bash)

```bash
curl -fsSL https://github.com/alimtvnetwork/movie-cli-v8/releases/download/v2.339.0/install.sh | bash
```

---

## What's Changed in v2.339.0

### Added / Changed — Structured appfault architecture, monadic Result containers, and response envelope

- **Structured `appfault` Architecture**: Implemented comprehensive `*appfault.AppError` package in `pkg/appfault/` (`types.go`, `methods.go`, `stack.go`, `appfault.go`) strictly adhering to `02-spec/03-error-manage/` with memory-optimized 136-byte pointer scan layout.
- **Monadic Result Containers**: Introduced generic monadic `Result[T]` and `ResultSlice[T]` types with pointer-attached inspection methods (`IsCountOtherThan`, `IsEmpty`, `HasRecord`, `HasRecords`, `IsDefined`) and line-1 nil guards preventing runtime panics.
- **Universal Response Envelope**: Built `pkg/appfault/envelope.go` standardizing REST API payloads with PascalCase keys (`Status`, `Attributes`, `Results`, `Errors`).
- **REST Error Upgrades**: Modernized `writeRestError` in `cmd/movie_rest_staged.go` and migrated raw `http.Error` calls in `cmd/movie_rest_report.go` (`handleMediaByID`, `handleMediaGet`, `handleMediaDelete`, `handleMediaPatch`) to emit structured JSON envelopes with stack traces in debug mode.
- **Structured Diagnostics Logging**: Extended `errlog/logger.go` with `ErrorFault(*appfault.AppError)` capturing fault codes, severity, contextual fields, and call stacks into SQLite `error_logs` and `error.txt`.
- **TMDB HTTP Error Wrapping**: Refactored `%w` format strings to `appfault.Wrap` / `appfault.Wrapf` in `tmdb/http.go`, ensuring clean causal chains and eliminating `govet` printf linter warnings.
- **Automated Release Ceremony**: Enhanced `03-ai-scripts/29-release-orchestrator.py` to support rich changelog bullet injection, automatic test inventory tracking, and direct GitHub release publication with Quick Install one-liners.
- **Clean CI Quality Gates**: All 15 local CI/CD quality gates verified 100% green (`golangci-lint` and `06-cicd-local-runner.py run-smart`).
