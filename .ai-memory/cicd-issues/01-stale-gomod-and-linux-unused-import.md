# CI/CD Issue: Stale go.mod Direct Requirement and Linux Trash Unused Import

- **Job:** Release (`#35457736498`), Lint (`#35457736430`), E2E (`#35457736396`)
- **Type:** FAIL
- **Detected:** 2026-09-19 17:19:35 UTC
- **Status:** resolved

## Error
```text
[Release Step: Pre-release migration audit checklist]
go.mod/go.sum out of date — run 'go mod tidy' before tagging.
5a6
> 	github.com/mattn/go-isatty v0.0.16
15d15
< 	github.com/mattn/go-isatty v0.0.16 // indirect

[Lint Step: Go vet]
pkg/trashbin/trash_linux.go:10:2: "strings" imported and not used

[E2E Step: Run Go integration test]
pkg/trashbin/trash_linux.go:10:2: "strings" imported and not used
FAIL	github.com/alimtvnetwork/movie-cli-v8/cmd [build failed]
```

## Root Cause
1. `pkg/trashbin/trash_linux.go` imported `"strings"` without calling any functions from the package.
2. `cmd/help_formatter.go` directly imported `github.com/mattn/go-isatty`, requiring `go mod tidy` to promote it from indirect to direct dependency in `go.mod`.

## Fix Applied
1. Removed `"strings"` from `pkg/trashbin/trash_linux.go`.
2. Ran `go mod tidy` to update `go.mod` and `go.sum`.
3. Validated via `go vet` and local runner.

## Plan Task
Enqueued at `.ai-memory/plans/pending/13-cicd-gomod-and-trash-linux-import.md`
RCA documented at `.ai-memory/memory/issues/09-unused-import-and-stale-gomod-rca.md`
