---
name: read-memory-enhanced
description: "Load project identity, specifications, conventions, active plans, git commit history, and RCA records into agent context."
---

# Read Memory Enhanced Skill

## Overview

Executes the Prompt Architect Memory Retrieval, Git Commit History & Project Context Ingestion workflow. Loads the single source of truth from disk and prevents hallucination.

## Workflow Phases

### Phase 1: Git History Audit
- Inspect the last 10-30 commits via `git log -n 10 --stat` and `git log -n 30 --oneline`.
- Identify recently touched files, architectural shifts, and resolved issues.

### Phase 2: Spec & Memory Ingestion
- Read `.ai-memory/what-to-read.md` first.
- Ingest root `readme.md`, `AGENTS.md`, and `.ai-memory/coding-guidelines.md`.
- Read all learned memories in `.ai-memory/memory/learned/` and active plans in `.ai-memory/plans/`.
- Review retrospectives and RCA post-mortems in `02-spec/03-error-manage/` and `.ai-memory/cicd-issues/`.

### Phase 3: Verification & Alignment
- Verify repository identity, database schemas, command sets, and toolchain versions.
- Ensure strict lowercase filenames across the repository.
- Verify that no banned patterns exist in active context.
