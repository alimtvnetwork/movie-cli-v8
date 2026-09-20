# CI/CD Task: Stale go.mod and Linux Trash Unused Import

## Source
- Runner job: Release (#35457736498) / Lint (#35457736430) / E2E (#35457736396)
- Error type: FAIL
- Detected at: 2026-09-19 17:19:35 UTC

## Error Summary
`pkg/trashbin/trash_linux.go:10:2: "strings" imported and not used`
`go.mod/go.sum out of date — run 'go mod tidy' before tagging.`
`cmd/movie_rest_staged.go: database.GetMediaByID and ActionSimpleInput signature alignment`

## Required Fix
Remove unused imports from `pkg/trashbin/trash_linux.go` and `cmd/movie_rest_staged.go`, align database method calls, and run `go mod tidy` to sync `go.mod`.

## Acceptance Criteria
- [ ] `go vet ./...` passes on both Windows and Linux
- [ ] `go mod tidy` is clean with zero diff
- [ ] Local runner passes with exit code 0
- [ ] No regressions in any job

## Status
- [x] resolved
