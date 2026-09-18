---
name: Flat command syntax
description: All movie CLI commands are FLAT under `movie` — never write `movie movie <cmd>`. Common AI mistake.
type: constraint
---

# Flat Command Syntax — No `movie movie` Nesting

## Rule
Every command is a direct subcommand of the `movie` binary. There is **no** `movie` parent group inside the `movie` binary.

## Correct
- `movie scan /path`
- `movie config set tmdb_api_key XYZ`
- `movie ls`
- `movie info 123`
- `movie watch add 123`
- `movie tag add 1 favorite`

## WRONG (do not generate, do not document)
- ❌ `movie movie scan /path`
- ❌ `movie movie config set ...`
- ❌ `movie movie ls`

## Why this rule exists
Earlier project memory described a nested tree (`movie → movie → scan`) which caused the AI to generate `movie movie <cmd>` examples in chat, READMEs, specs, and UI copy. The actual cobra command tree is flat. The user explicitly flagged this as a recurring error that must never happen again.

## How to apply
- When writing CLI examples in chat, README.md, spec files, UI strings, comments, or anywhere else: ALWAYS use `movie <cmd>`.
- When reading old docs, treat any `movie movie <cmd>` as a bug to fix on sight.
- When updating the command tree in `.lovable/memory/01-project-overview.md`, keep it flat.
