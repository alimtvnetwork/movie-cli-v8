# Subtask 2: DRY Optimization Plan

## Goal
Analyze the Go codebase in `D:\work\movie-cli-v8` and make a plan to use common packages (like `pkg/appfault` or centralized tools) to make it DRY.

## Instructions
1. Use `run_command` to list `.go` files in `movie-cli-v8` (e.g., in `errlog`, `apperror`, `tmdb`).
2. Write a planning markdown file `D:\work\movie-cli-v8\02-spec\21-app\99-dry-optimization-plan.md` detailing how to use `coding-guidelines` standardized error wrapping (`*appfault.AppError`), unified boolean flags, or extracted shared structures to reduce the codebase size.
3. Apply any quick coding guideline fixes (e.g., line lengths, variable names) if necessary.
4. Record changes in `.ai-memory/temp-agents/07-movie-cli-migration-and-optimization/state.md`.
