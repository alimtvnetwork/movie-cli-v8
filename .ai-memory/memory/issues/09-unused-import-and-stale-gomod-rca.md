# RCA: Unused Import in Linux Trash Implementation and Stale go.mod Direct Requirement

- **Issue Slug:** `09-unused-import-and-stale-gomod-rca`
- **Severity:** High (Blocked Release, Lint, and E2E GitHub Actions Workflows)
- **Detected In:** GitHub Actions Pipeline Run #35457736498, #35457736430, #35457736396
- **Date:** 2026-09-20

---

## 1. Why It Happened

During release `v2.323.0`, two distinct defects slipped past local pre-commit checks:

1. **Unused Import in Platform-Specific File (`pkg/trashbin/trash_linux.go`):**
   `strings` was imported in `pkg/trashbin/trash_linux.go` during initial development of path sanitization for XDG trash formatting, but was never referenced in the final implementation. Because development and local validation took place on a Windows host (`GOOS=windows`), files with `//go:build linux` build tags were ignored by default Go compiler and `go vet` passes on the developer workstation. When the CI pipeline ran on `ubuntu-latest` (`GOOS=linux`), `go vet ./...` and `go test ./...` immediately failed with:
   ```text
   pkg/trashbin/trash_linux.go:10:2: "strings" imported and not used
   ```

2. **Out of Date `go.mod` and `go.sum` Direct Requirement:**
   The colorful terminal help formatter (`cmd/help_formatter.go`) introduced a direct import of `github.com/mattn/go-isatty`. Previously, `github.com/mattn/go-isatty` was listed in `go.mod` as an indirect dependency (`// indirect`). Adding direct usage in package code requires `go mod tidy` to promote it to the primary `require (...)` block. Because `go mod tidy` was not executed prior to tagging, Step 3 of the pre-release migration audit in `.github/workflows/release.yml` detected a diff between the checked-in `go.mod` and the evaluated `go mod tidy` output, terminating the release workflow.

---

## 2. How It Happened

1. **Host-Target Build Tag Asymmetry:**
   `pkg/trashbin` was intentionally designed with OS-specific implementations: `trash_windows.go`, `trash_darwin.go`, and `trash_linux.go`.
   Local tests and checks run on Windows without explicit cross-compilation flags (e.g. `GOOS=linux go vet ./...`) only analyze Windows-tagged files and agnostic files, silently skipping `trash_linux.go`.
2. **Missing Pre-Tag Module Hygiene Check:**
   When adding the `github.com/mattn/go-isatty` import to `cmd/help_formatter.go`, the code compiled locally without errors because the module was already cached in `go.sum` from indirect dependencies. However, the release workflow specifically runs:
   ```bash
   cp go.mod /tmp/go.mod.before
   cp go.sum /tmp/go.sum.before
   go mod tidy
   diff -q /tmp/go.mod.before go.mod
   ```
   Because `go mod tidy` was omitted prior to creating release branch `release/v2.323.0`, the CI runner failed this strict consistency assertion.

---

## 3. Root Cause

1. **`pkg/trashbin/trash_linux.go` Root Cause:**
   Extraneous `"strings"` import remaining in `trash_linux.go` after drafting XDG trash spec path handling. Lack of cross-platform matrix vetting in local tooling allowed the unreferenced import to escape detection until executed in the Ubuntu runner environment.
2. **`go.mod` Root Cause:**
   Direct dependency upgrade of `github.com/mattn/go-isatty` without running `go mod tidy` to re-synchronize `go.mod` and `go.sum`.

---

## 4. Code Fix

1. **Surgically Removed Unused Import:**
   In `pkg/trashbin/trash_linux.go`, removed `"strings"` from the import declaration:
   ```go
   import (
       "fmt"
       "os"
       "os/exec"
       "path/filepath"
       "time"

       "github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
   )
   ```
2. **Synchronized Go Module Dependencies:**
   Executed `go mod tidy` at repository root, promoting `github.com/mattn/go-isatty v0.0.16` into the explicit `require (...)` block in `go.mod` and pruning obsolete indirect tags.
3. **Cross-Platform Local Verification:**
   Verified `go vet` both for native host and simulated `GOOS=linux` target to guarantee zero compilation or vetting regressions across platforms.
