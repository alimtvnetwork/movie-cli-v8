# DRY Optimization Plan

## Goal
To reduce codebase size and improve compliance with the `coding-guidelines` through DRY optimization, specifically by standardizing error wrapping, unifying boolean flags, and extracting shared parameter structs.

## 1. Standardize Error Wrapping
**Current State:** The CLI uses an internal wrapper (`github.com/alimtvnetwork/movie-cli-v8/apperror`) for error management (e.g. `apperror.Wrap`, `apperror.Wrapf`, `apperror.New`).
**Target State:** Transition to the `*appfault.AppError` standard specified in `coding-guidelines`. 
**Action Plan:**
- Run a project-wide search-and-replace to update all imports from `github.com/alimtvnetwork/movie-cli-v8/apperror` to `pkg/appfault`.
- Update functions returning structured failure metadata to return `*appfault.AppError` instead of standard `error`.
- Refactor calls across `errlog/logger.go`, `tmdb/client.go` and `tmdb/http.go` from `apperror` utilities to `appfault` utilities. Remove `apperror/apperror.go`.

## 2. Unify Boolean Flags (Strict Boolean Standard)
**Current State:** Several fields in option structs violate the "Strict Boolean Standard: is and has only" rule. For example:
- `UseTable bool` (found in `cmd/types.go` and `cmd/movie_scan_process.go`)
- `UseJson bool` (found in `cmd/types.go`)
- `UserProvidedPath bool` (found in `cmd/path_scope.go`)
**Target State:** Boolean variables must strictly use `is` or `has` prefixes (e.g., `can`, `should`, `use`, `will` are banned).
**Action Plan:**
- Rename `UseTable` to `IsTableOutput`.
- Rename `UseJson` to `IsJsonOutput`.
- Rename `UserProvidedPath` to `IsUserProvidedPath`.
- Update all associated CLI flag bindings in `cmd` packages.

## 3. Extract Shared Structures
**Current State:** `cmd/types.go` contains nearly 30 ad-hoc option structs with 1 to 5 fields each. Many share similar subsets of parameters (e.g., Database, Client, output format flags). Examples:
- `ScanOutputOpts { UseTable bool; UseJson bool }`
- `ScanLoopConfig { UseJson bool; UseTable bool; Client; ... }`
- Many input structs in `cmd/types.go` have redundant properties.
**Target State:** Consolidate fragmented structs into a set of embedded core configuration structs.
**Action Plan:**
- Create an `OutputFormatOpts` struct containing `IsJsonOutput` and `IsTableOutput`, and embed it in `ScanLoopConfig` and `ScanOutputOpts`.
- Identify functions with overlapping parameter lists (e.g. `Client *tmdb.Client`, `Database *db.DB`) and unify them into a generic `AppContext` or `ServiceContext` struct.
- Review generic `XxxInput` structs in `types.go` and combine them into a single polymorphic context object if applicable.

## 4. Line Length & Quick Formatting Fixes
**Current State:** Code in `tmdb/client.go`, `errlog/logger.go` has chains and literals exceeding 100 characters.
**Target State:** Strictly conform to < 100 character limits on logic lines and multi-line parameter formatting.
**Action Plan:**
- Wrap long lines in `tmdb/client.go` (e.g., `GetRecommendations` and `storeImdbCache`).
- Apply the mandatory multi-line formatting to `apperror.Wrap`/`apperror.New` calls with 2 or more arguments.
