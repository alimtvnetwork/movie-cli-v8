# Error Handling Principles

> **Scope:** Repository-wide (Go packages, CLI commands, scripts)  
> **Source:** `02-spec/03-error-manage/01-index.md` & `AGENTS.md`

---

## 1. Structured Go AppError Type (`*appfault.AppError`)

- In all Go packages, functions returning structured failure metadata MUST use `*appfault.AppError` as their return type (e.g. `func validate() *appfault.AppError`).
- Import path is strictly `github.com/alimtvnetwork/movie-cli-v8/04-code/golang/pkg/appfault` (or repo equivalent `pkg/appfault`).
- Do not use package-stutter: use `appfault.AppError` (never `apperror.AppError`).
- Use monadic containers `Result[T]`, `ResultSlice[T]`, and `ResultMap[K, V]` with `.AppError()` or `.Fault()` accessors.

## 2. No Swallowed Errors (CODE RED)

- Every error caught or received MUST be wrapped with operational context or logged explicitly before returning.
- Silent error suppression, bare `return nil`, empty catch blocks, or ignoring error returns is strictly forbidden.
- Always enrich errors with path context (`.WithPath(path)`) and variable context (`.WithVar(name, val)`).

## 3. No Bare Void in Go

- Functions performing business logic or mutations MUST return either `Result[T]` or `*appfault.AppError`.
- Bare void functions (`func doWork()`) that hide error outcomes are banned.

## 4. Retrospective Verification

- Before closing bug-fix or error remediation turns, verify the error pattern against existing retrospectives in `02-spec/03-error-manage/01-error-resolution/03-retrospectives/` and `.ai-memory/cicd-issues/`.
