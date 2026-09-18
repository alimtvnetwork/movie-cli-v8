# Consolidated Subtasks: Guideline Prompt & Installer Upgrade

> **Parent Plan:** [`.ai-memory/plans/pending/04-guideline-prompt-and-installer-upgrade.md`](.ai-memory/plans/pending/04-guideline-prompt-and-installer-upgrade.md)
> **Status:** `PENDING`
> **Consolidated Tasks Merged:** `01-format-overview-prompt.md`, `02-upgrade-installer-scripts.md`, `03-generate-50-improvements.md`, `04-release.md`

## 1. Subtask 1: Format Overview as Prompt

1. Read `02-spec/02-coding-guidelines/00-overview.md` (or the equivalent).
2. Rewrite it to be an AI prompt with explicitly "must follow" instructions.
3. Extract the rule about method arguments (more than 3 args or > 100 chars -> one per line) and create `03-coding-style-checklist.md` in the root of the repo as a prompt instruction file.
4. Ensure indexes are updated properly.

## 2. Subtask 2: Upgrade Installers

- **Status:** ✅ Done
- **Started:** 2026-09-05T10:04:00+08:00
- **Completed:** 2026-09-05T10:12:00+08:00

1. Edit `scripts/generate-bundle-installers.mjs`.
2. Enhance the PowerShell and Bash scripts logic for `version.json` merge:
   - For `codingGuideline`, calculate `installedFiles` array (by listing files extracted from the bundle).
   - Before extracting, read `codingGuideline.installedFiles` from the existing `version.json`. Any file in the old list that is NOT in the new list should be explicitly deleted.
   - Output an `install-summary.json` file inside `.ai-memory/` detailing versions updated and files removed.
3. Generate bundles (`npm run bundles:generate`).

## 3. Subtask 3: 50 Improvements for Error Manage

1. Read the `02-spec/03-error-manage/` folder contents.
2. Create `02-spec/03-error-manage/02-improvements.md`.
3. Draft exactly 50 concrete, numbered improvements covering logging, API structure, code mapping, testing, etc.
4. Link this file in the `00-overview.md` index of error manage.

## 4. Subtask 4: Release

1. Run `node scripts/release.mjs --tier minor --scope "Upgrade installer scripts and coding guideline prompt" --skip-slides`.
2. Fix sync drift.
3. Generate bundles.
4. Update memory pins.
5. Commit and push v6.29.0.
