---
name: coding-guidelines
description: "Use this skill to audit, review, and enforce coding guidelines across all languages."
---

# Coding Guidelines Enforcement Skill

## Overview

Enforce repository coding guidelines across Go, TypeScript, Python, and shell scripts without regression or compromise.

## Core Rules

1. **Positive Booleans Only:** Identifier prefix must be `is` or `has` (`Is` / `Has` for PascalCase). Negative prefixes and alternative verbs (`can`, `should`, `was`) are prohibited.
2. **Implicit Boolean Evaluation:** Total ban on `== true` or `=== true`.
3. **No Mixed Polarity:** Do not combine positive and negative conditions in the same `if` statement (`if isA && !isB` is forbidden).
4. **Structured Errors:** Use `*appfault.AppError` in Go packages. Do not swallow errors.
5. **Vertical Line Gaps:** Blank line before `if`, after `}`, and before `return`.
6. **Bounded Functions & Files:** Functions max 15 lines (preferred 8). Files max 300 lines.
7. **No Loose Parameters:** Use `*Params` structs for functions with > 2-3 arguments.
8. **Relative Git Paths:** All markdown references and paths must be relative from repo root.

## Execution Workflow

1. **Explore & Map:** Identify target files and trace dependencies before editing.
2. **Micro-Batching:** Limit refactoring passes to bounded batches of 5-8 files.
3. **Verify Compliance:** Run targeted linters and validation scripts from `03-ai-scripts/`.
