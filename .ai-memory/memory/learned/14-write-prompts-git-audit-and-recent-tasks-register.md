# Write Prompts 30-Commit Git Audit & Recent 20-Task Register Architecture

> **Type:** Institutional Knowledge & Learned Architecture  
> **Status:** Active & Canonical  
> **Date:** 2026-09-13  
> **Reference Plan:** `05-changes-history/26-write-prompts-git-audit-and-recent-tasks-register/01-transaction-log.md`

---

## 1. Executive Overview

This document formalizes the architectural decisions, user directives, and operational protocols governing the upgrade of the memory-writing workflows (`01-write-antigravity.md` and `03-write-memory.md`) to Version 2.2.0. The upgrade mandates an audit of the last 30 Git commits before authoring memory, institutes a compact 20-task recent completion register in `.ai-memory/plans/01-index.md`, and establishes a 19-box verification checklist before turn completion.

---

## 2. Mandatory Directives & User Directives

### 2.1. Mandatory 30-Commit Git History Audit

> [!IMPORTANT]
> **User Directive (Verbatim):**  
> *"improve the write prompts, memory write prompts, along with the, let's say, last 30 Git histories of summaries, what it has done, uh, what are the things it could learn, and also check the main task list, what is completed, so that it knows what is completed recently, so that in a compact sense, it would know what is done, what is not. Remember to include that in the what to read and also the task list. It would refer to that file to understand, like, what task is done. Um, so this is something, like, I want, um, you to improve in the both of the write prompts, okay? In a way that, uh, if I run this, then last, uh, let's say, five, 10 conversation or task, or let's say 20 task. 20 task, let's say. 20 task should be, uh, written inside, uh, .loggable folder in the plans or somewhere so that it refers back from the index file and also the what to read file, so that, uh, whenever we, we create a new memory, it knows what is done recently. It has the recent memory, also the recent things"*

- **Principle:** AI agents cannot reliably author memory or plan subsequent tasks based solely on ephemeral chat context. Ephemeral context drops details across turns, leading to repeated work, resurrected anti-patterns, and blind spots.
- **Rule:** Before authoring or updating memory files, the AI MUST execute:
  ```sh
  git log -n 30 --oneline
  ```
  (and `git log -n 30 --stat` where needed) to inspect the last 30 commits.
- **Extraction Protocol:** From the commit log, the agent must extract:
  1. What code, docs, and configurations were modified, added, or deleted.
  2. What explicit user directives and architectural guidelines were enforced.
  3. What bugs, CI/CD failures, or lint issues were diagnosed and resolved.
  4. What institutional lessons and anti-patterns must be permanently recorded.

---

### 2.2. Recent 20-Task Tracking & Compact Task Register

- **Problem:** As codebases grow, the number of historical tasks and completed plans multiplies into hundreds of files. Scanning every historical file overwhelms context tokens, while skimming produces blind spots.
- **Solution:** A rolling, compact **Recent Completed Tasks Register (Last 20 Tasks)** embedded directly in `.ai-memory/plans/01-index.md`.
- **Requirements:**
  1. The register tracks the 20 most recent tasks and completed plans, sorted in reverse chronological order.
  2. Each entry includes the task number, title, relative path link to `05-changes-history/` or `plans/completed/`, completion date, and a concise summary.
  3. `.ai-memory/what-to-read.md` references this register under "Before any task (always)".
  4. Every new memory write updates this register, retiring the oldest entry beyond 20 to maintain a strictly bounded window.

---

## 3. Workflow Integration & Pre-Flight Sequence

### 3.1. Pre-Flight Step 0

Both write prompts now define **Step 0** in their Pre-Flight sequence:
```markdown
0. `git log -n 30 --oneline` — inspect the last 30 commits to understand recent file changes, what code/docs were touched, recent bug fixes, and what the AI can learn from recent history before starting memory capture.
```
Step 0 executes *before* opening or modifying any file, ensuring the agent's internal model is aligned with ground truth on disk.

### 3.2. Phase 1 Session Audit Structure

Phase 1 (Internal Session Audit) now explicitly requires answering:
- **Git History Audit (Last 30 Commits):** Trajectory, recent bug fixes, directives, and learned patterns.
- **Recent Tasks Status (Last 20 Tasks):** Compact review of the 20 tasks in `.ai-memory/plans/01-index.md` vs remaining pending tasks.
- **Done / Pending / Learned / Avoid / Ambiguities / Suggestions / User Commands.**

### 3.3. Standardized 19-Box Verification Checklist

Both write prompts enforce an identical 19-box verification checklist covering:
1. 30-commit git history audit execution.
2. 20-task status register audit in `.ai-memory/plans/01-index.md`.
3. Pre-flight recursive walk of `.ai-memory/`.
4. Topic folder nesting (no memory files at memory root).
5. Synchronized index updates (`01-index.md`).
6. Plan lifecycle transitions (`pending/` -> `completed/` via `mv`).
7. Append-only `strictly-avoid.md`.
8. Verbatim capture of user directives without softening.
9. Zero consolidation or shortening of detailed specs.
10. Strict lowercase `readme.md` verification.
11. Real numbers in completion summaries (zero `[N]` or `[X]` placeholders).

---

## 4. Cross-Reference Map

- **Master Memory Index:** `.ai-memory/memory/01-index.md`
- **Master Plans Index:** `.ai-memory/plans/01-index.md`
- **Canonical Reading Sequence:** `.ai-memory/what-to-read.md`
- **Write Antigravity Prompt:** `01-prompts/03-read-write/01-write-antigravity.md`
- **Write Memory Prompt:** `01-prompts/03-read-write/03-write-memory.md`
- **Write Antigravity Skill:** `.agents/skills/write-antigravity/skill.md`
- **Write Memory Skill:** `.agents/skills/write-memory/skill.md`
- **Task Transaction Log:** `05-changes-history/26-write-prompts-git-audit-and-recent-tasks-register/01-transaction-log.md`
