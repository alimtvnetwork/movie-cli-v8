---
name: execute-pending-tasks
description: "Autonomously orchestrate and execute pending tasks in continuous self-loops with bounded micro-tasking."
---

# Execute Pending Tasks Skill

## Overview

Executes pending architectural and implementation plans from `.ai-memory/plans/pending/` and `.ai-memory/plans/subtasks/` using bounded micro-tasking and continuous self-loops.

## Execution Directives

1. **Pre-Flight Inspection:** Read the target plan and associated subtasks before touching code.
2. **Micro-Tasking:** Execute tasks in small, isolated steps (5-8 files maximum per batch).
3. **Continuous Self-Loop:** Check off completed items in plan files, log transaction records, and move completed plans to `.ai-memory/plans/completed/`.
4. **Tool Reuse:** Consult `03-ai-scripts/01-index.md` to reuse repository automation scripts instead of creating redundant temporary scripts.
5. **Atomic Commit & Push:** Group related changes into a single atomic commit and push immediately to remote.
