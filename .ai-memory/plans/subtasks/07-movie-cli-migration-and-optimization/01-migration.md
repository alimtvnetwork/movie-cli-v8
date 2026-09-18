# Subtask 1: Migration of spec and .lovable

## Goal
Migrate the folder structure in `D:\work\movie-cli-v8` to the modern one from `D:\work\coding-guidelines`, preserving app specs.

## Instructions
1. Run powershell commands to copy `D:\work\coding-guidelines\02-spec`, `.ai-memory`, `01-prompts`, `03-ai-scripts`, and `agents.md` to `D:\work\movie-cli-v8\`.
2. Move contents of `D:\work\movie-cli-v8\spec\08-app` into `D:\work\movie-cli-v8\02-spec\21-app\` if `spec\08-app` exists. (Keep the files!). Do the same for `09-app-issues` -> `22-app-issues`.
3. If `D:\work\movie-cli-v8\.lovable\memory` exists, copy its contents into `D:\work\movie-cli-v8\.ai-memory\memory\`.
4. Delete the old `.lovable` and `spec` folders and `AGENTS.md` in `D:\work\movie-cli-v8` after moving their important contents.
5. Record changes in `.ai-memory/temp-agents/07-movie-cli-migration-and-optimization/state.md`.
