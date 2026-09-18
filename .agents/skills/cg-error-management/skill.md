---
name: cg-error-management
description: "Master and enforce structured error management, *appfault.AppError handling, and retrospective compliance across all packages."
---

# Error Management Skill (`cg-error-management`)

## Overview

Enforce repository error management standards, structured `*appfault.AppError` generation, contextual wrapping, and retrospective failure prevention.

## Core Rules

1. **Structured AppError:** Use `*appfault.AppError` for all Go errors requiring structured failure metadata. Never return raw Go `error` or bare void from mutation functions.
2. **Never Swallow Errors (CODE RED):** Every error caught or received must be logged with context or wrapped and returned.
3. **Context Wrapping:** Enrich errors with path context (`.WithPath(path)`) and variable context (`.WithVar(key, val)`).
4. **Universal Response Envelope:** APIs and services return standardized envelopes: `{ "Data": ..., "Errors": [...], "Meta": ... }`.
5. **Retrospective Compliance:** Prior to resolving bugs, consult `02-spec/03-error-manage/01-error-resolution/03-retrospectives/` to ensure past failure modes are not resurrected.
