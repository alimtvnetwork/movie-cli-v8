# Plan: Regex Centralization, Core Harvesting, Generic DB Engine & Code Generator

> **Tracking ID:** `12-regex-centralization-and-generic-dbengine`
> **Status:** ✅ Completed
> **Created:** 2026-09-13
> **Completed:** 2026-09-13

---

## 1. Executive Summary

Successfully accomplished full regex pattern harvesting, lazy regex engine audit, call-site refactoring across `gitmap/cli`, porting of the generic `dbengine` database wrapper into `coding-guidelines/04-code/golang/pkg/dbengine` with standard `*appfault.AppError` and typed `Result[T]` envelopes, and porting of the code generator script `35-db-struct-enum-generator.py`.

---

## 2. Completed Subtasks Summary

### Subtask 01: Lazy-Regex Audit & Core Regex Harvesting

- **Audit & Polarity Fix:** Inspected `gitmap/cli/lazyregex/lazyregex.go`, audited for boolean principles, and eliminated mixed-polarity conditionals in `Compile`, `CompileMust`, and `compiledRegex`.
- **Harvesting from `03-aukgo/core`:**
  - Harvested 80+ regex constants and lazy precompiled instances from `03-aukgo/core/regconsts/regconsts.go` and `03-aukgo/core/regexnew/regexes-compiled.go`.
  - Created `gitmap/cli/lazyregex/regconsts.go` and `gitmap/cli/lazyregex/regexes_compiled.go`.
  - Created `coding-guidelines/04-code/golang/pkg/regexnew/regconsts.go` and `coding-guidelines/04-code/golang/pkg/regexnew/regexes_compiled.go`.
  - Standardized `UUIDAny` alongside version-specific `UUID4` to support arbitrary valid UUIDs without version mismatch errors.
- **Verification:**
  - `gitmap/cli/lazyregex`: 11/11 tests pass (0.181s).
  - `coding-guidelines/04-code/golang/pkg/regexnew`: 14/14 tests pass (0.967s).

### Subtask 02: GitMap Common Regex Centralization & Refactor

- **Pattern Mining:** Scanned all Go source files across `gitmap/cli` and extracted 60+ repository-specific regular expressions.
- **Lazy Engine Byte Support:** Added `Find(b []byte) []byte`, `FindBytes(b []byte) []byte`, and `IsMatchBytes(b []byte) bool` to both `lazyregex.LazyRegexp` and `regexnew.LazyRegex`.
- **Caller Refactoring:** Refactored high-frequency callers across `gitmap/cli` to eliminate ad-hoc recompilations:
  - `cmd/filemanipulator.go`: `lazyregex.NumberPrefixRegex.Compiled()`
  - `cmd/sequence_cmd.go`: `lazyregex.NumberPrefixRegex.Compiled()`
  - `cmd/folder/metadata.go`: `lazyregex.NumberPrefixRegex`
  - `pipelinedb/pipeline_split_db.go`: `lazyregex.SlugSanitizeRegex`
  - `store/schedule_split_db.go`: `lazyregex.SlugSanitizeRegex`
  - `clonenext/version.go`: `lazyregex.RepoVersionSuffixRegex`
  - `cmdpipeline/pipeline_error_extract.go`: `lazyregex.AnsiEscapeRegex`
  - `cmdagy/agy_conv_scanner.go`: `lazyregex.FileUriRegex`
- **Verification:** All modified packages in `gitmap/cli` compile cleanly and pass 100%.

### Subtask 03: Generic Database Wrapper Package (`pkg/dbengine`)

- **Package Implementation:** Created clean, generic, redistributable database wrapper under `coding-guidelines/04-code/golang/pkg/dbengine/`:
  - `dialect.go`: Multi-dialect SQL abstraction (SQLite, PostgreSQL, MySQL, MariaDB, MSSQL, Oracle, MongoDB).
  - `operators.go`: Strongly-typed SQL comparison and logical operators with JSON serialization.
  - `compiler_sql.go`: Dialect compilers for PostgreSQL, MySQL, MSSQL, Oracle, MongoDB.
  - `compiler_sqlite.go`: SQLite-specific dialect compiler, view definitions, and column inspection.
  - `executor.go`: `SqlExecutor` unified interface and `TxWrapper` implementation.
  - `result_types.go`: Typed monadic result envelopes (`Uint64Result`, `Int64Result`, `StringResult`, `BoolResult`, `RowsAffectedResult`, `EntityResult[T]`, `ListResult[T]`, `CompiledQueryResult`) wrapping `appfault.Result[T]`.
  - `scan_helpers.go`: Type-safe scan conversion utilities (`ScanString`, `ScanInt`, `ScanInt64`, `ScanUint`, `ScanUint64`, `ScanBool`, `ScanFloat64`).
  - `wrapper.go`: Connection pooling, transaction isolation, view caching via `__dbengine_view_meta`, and schema inspection.
  - `query.go`: Fluent generic `QueryBuilder[T, F]` and `JoinBuilder[T, F]` supporting CTEs, joins, group by, having, locate, order by, and deterministic SHA-256 query caching.
  - `query_cache.go`: Thread-safe query compilation cache.
  - `repository.go`: Generic `Repository[T, F]` with strongly-typed CRUD queries.
- **Zero External Dependencies:** Built solely on standard library `database/sql` and `pkg/appfault`.
- **Unit Tests:** Created `scan_helpers_test.go` and `dbengine_test.go` covering operators, compilers, query builder compilation, query caching, mock executor repository operations, and scan helpers.
- **Verification:** `go test -v -count=1 ./pkg/dbengine/...` passes 100% (0.746s).

### Subtask 04: Python Database Struct & Enum Code Generator

- **Script Implementation:** Created `coding-guidelines/03-ai-scripts/35-db-struct-enum-generator.py`.
- **Dynamic Module Resolution:** Implemented `find_go_module_root()` to automatically detect Go module boundaries from `go.mod` and resolve proper package import paths across projects.
- **Repo Standards Compliance:** Adheres to Python 3.10+ typing, `02-shared-engine.py` architecture, strict lowercase naming, and boolean conventions.
- **Catalog Update:** Registered scripts 34 and 35 in `coding-guidelines/03-ai-scripts/01-index.md`.
- **Verification:** Tested via `python 03-ai-scripts/35-db-struct-enum-generator.py --dry-run`.

---

## 3. Verification Commands Executed

| Package / Target | Command | Result |
| :--- | :--- | :--- |
| `coding-guidelines/common/pkg/dbengine` | `go test -v -count=1 ./pkg/dbengine/...` | ✅ 100% Pass |
| `coding-guidelines/common/pkg/regexnew` | `go test -v -count=1 ./pkg/regexnew/...` | ✅ 100% Pass |
| `gitmap/cli/lazyregex` | `go test -v -count=1 ./lazyregex/...` | ✅ 100% Pass |
| `gitmap/cli/dbengine` | `go test -v -count=1 ./dbengine/...` | ✅ 100% Pass |
| `35-db-struct-enum-generator.py` | `python 03-ai-scripts/35-db-struct-enum-generator.py --dry-run` | ✅ 100% Pass |
