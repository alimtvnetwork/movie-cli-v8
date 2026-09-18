# Consolidated Plan: Multi-Repo Folder Structure Migration & Gitmap Fixes

**Started From**: User requested to fix the folder structure across gitmap, wp-onboarding, and 10 other child repositories, migrating `.lovable` to `.ai-memory`, `spec` to `02-spec`, leaving no trace of `.lovable`, and diagnosing/fixing missing `gitmap agy pipeline-fix` helptext and releasing gitmap.
**Loops/Steps Taken**: 1 comprehensive orchestrator loop (planning, gitmap fixes, release v6.259.1, automated multi-repo migration across 12 repositories).

---

## User Request (Verbatim)

```text
Is it done?

Hi there. Um, so first of all, uh, from the, uh, coding guideline, I want you to fix, uh, the Git map, uh, folder structure, uh, .lovable folder and everything. So start with that big step. Uh, that should be the first

You need to read the code base, uh, the lovable folder, the AI memory folder, how it is restructured than previously. Uh, you need to understand the current folder structure. The same way you need to, uh, fix the folder structure in Git map, uh, WP onboarding, and, uh, uh, let's say other places. Uh, what do you mean by other places? That means in the, uh, scripts folder, in the Cat Malaysia, uh, CatMy UI V12, uh, AGM, uh, then Stata, the scripts fixture. Okay, all of these, and also the, uh, Let's Mark, uh, Let's Mark UI, WP link manager. Then we have, uh, the Laravel, Lara licensing, yeah, Lara licensing, Laravel automation. Um, yeah, so all of these projects, uh, actually have the old folder structure of a spec.lovablefolder. I want you to understand first how this is done properly. So that's the first thing I want you to understand. Once you understand that, then your next job is to fix on those code bases, okay, properly, and leave no trace of .lovablefolder. That means-- What, what do I mean by that? Because in past, inside the prompts, the .lovablefolder is mentioned several times. So first you understand the coding guideline, and wherever there is a difference in the Git map, you fix that. Once you learn that, then you fix into the other code repositories. Do you understand what I'm saying?

also in gitmap check and fix and release gitmap
|
Okay, um, what I don't understand is that, uh, we actually requested for the git map, uh, git pipeline AGY fix. So these commands are not there. The help is not there, so nothing is there. So please find those error fix with, uh, Antigravity. What are the commands? These are the commands. Fix those. Why the help text is not there. Explain me those, okay? So your job is to fix it

please list out all tasks as step 1
```

---

## Extracted Actionable Task List

- [x] **Task 1 (Architecture & Coding Guideline Ingestion):** Read and internalize the modern meta-repository structure (`.lovable/` -> `.ai-memory/`, `spec/` -> `02-spec/`, relative git path rules, and prompt updates).
- [x] **Task 2 (Gitmap AGY Pipeline Fix & Help Text):** Identify why `gitmap agy fix pipeline` and `gitmap pipeline fix` were missing, why help text was absent, fix the command routing and markdown help rendering, and add automated test coverage.
- [x] **Task 3 (Gitmap Verification & Release):** Run the release orchestrator, bump version to `v6.259.1`, tag, push, and publish the GitHub release with clean release notes.
- [x] **Task 4 (Child Repository Structure Migration - 11 Repositories + Gitmap):** Pre-pull each repository (`git pull origin main`), rename `.lovable/` -> `.ai-memory/`, rename legacy `spec/` -> `02-spec/`, update `.gitignore`, and perform exhaustive search-and-replace to eradicate any occurrences/references to `.lovable`.
- [x] **Task 5 (Atomic Git History & Remote Push):** Verify working tree cleanliness across all 12 repositories, commit atomically per repo (no per-file commits), and push directly to remote GitHub repositories.
- [x] **Task 6 (Plan Consolidation & Memory Indexing):** Consolidate subtasks into a single completed plan document in `.ai-memory/plans/completed/`, clean up pending directories, and update the memory index.

---

## 1. Summary of Completed Deliverables

### Subtask 01: Coding Guidelines & Folder Structure Grounding
- Internalized the architectural transformation:
  - Renaming `.lovable/` to `.ai-memory/` across all repositories.
  - Renaming `spec/` to `02-spec/` for standardized numeric prefixes.
  - Removing duplicate `.ai-memory/prompts` and `.ai-memory/ai-fix-scripts` when root `01-prompts` or `03-ai-scripts` exist.
  - Replacing all internal file references, links, prompts, documentation, and configs to point to `.ai-memory/`.

### Subtask 02: Gitmap AGY Pipeline Fix Command & Helptext
- Root Cause Identified:
  - `renderAgyHelp` in `cli/cmdagy/agy_help.go` fell back to generic Cobra `cmd.Usage()` on subcommands, ignoring `cli/helptext/agy-fix-pipeline.md`.
  - `agyFixPipelineCmd` lacked a custom `SetHelpFunc` and `RunPipelineFixAgyCLI` ignored `--help` flags.
  - `IsPipelineFixAgyArgs` failed to match `fix` when passed standalone or with `--help`.
  - `printPipelineHelp` did not document `fix` / `fix errors agy`.
- Fixes Applied:
  - Wired `helptext.PrintWithMode("agy-fix-pipeline", render.PrettyAuto)` to `agyFixPipelineCmd.SetHelpFunc`.
  - Intercepted help flags in `RunPipelineFixAgyCLI`.
  - Added `renderSubcommandHelp` in `agy_help.go`.
  - Enhanced argument matching in `cmdpipeline/pipeline_fix_agy_runner.go`.
  - Added `fix` and examples to `cmdpipeline/pipeline.go`.
  - Added comprehensive test cases in `cmdpipeline/pipeline_fix_agy_test.go`.

### Subtask 03: Gitmap Verification & Release
- Staged and committed changes atomically with `20bd910f`.
- Executed release orchestrator `03-ai-scripts/29-release-orchestrator.py`.
- Published GitHub Release `v6.259.1`: https://github.com/alimtvnetwork/gitmap-v28/releases/tag/v6.259.1.
- Synced `main` branch with remote.

### Subtask 04 & 05: Multi-Repo Migration Ledger
All 12 repositories were pulled before modification, migrated to the new folder structure, all file references updated, verified clean, committed atomically, and pushed to remote:

| Repository | Path | `.lovable` Removed | `.ai-memory` Active | `spec` -> `02-spec` | Remote Status |
|---|---|:---:|:---:|:---:|:---:|
| **gitmap** | `../gitmap` | Yes | Yes | Yes | Released v6.259.1 |
| **wp-onboarding** | `../wp-onboarding` | Yes | Yes | Yes | Pushed (clean) |
| **scripts-fixer** | `../scripts-fixer` | Yes | Yes | Yes | Pushed (clean) |
| **cat-my** | `../cat-my` | Yes | Yes | Yes | Pushed (clean) |
| **ui-prompts-cat** | `../ui-prompts-cat` | Yes | Yes | N/A | Pushed (clean) |
| **Antigravity-Manager** | `../Antigravity-Manager` | Yes | Yes | Yes | Pushed (clean) |
| **alim-status-sample** | `../alim-status-sample` | Yes | Yes | Yes | Pushed (clean) |
| **letsmarknow** | `../letsmarknow` | Yes | Yes | Yes | Pushed (clean) |
| **letsmarknow-ui** | `../letsmarknow-ui` | Yes | Yes | Yes | Pushed (clean) |
| **wp-link-manager** | `../wp-link-manager` | Yes | Yes | Yes | Pushed (clean) |
| **lara-licensing** | `../lara-licensing` | Yes | Yes | Yes | Pushed (clean) |
| **laravel-automation** | `../laravel-automation` | Yes | Yes | Yes | Pushed (clean) |
