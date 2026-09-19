---
name: Release Architecture Map
description: Architectural map of how versioning and releases work in movie-cli (where version lives, how it propagates, sync pipeline, and release ceremony).
type: standard
---

# Release Architecture Map

**Repository:** alimtvnetwork/movie-cli-v8
**Canonical Version Source:** version.json (master manifest) + version/info.go (Go binary version) + package.json (npm/scripts)
**Reading Queue:** Enqueued in .ai-memory/what-to-read.md and indexed in .ai-memory/memory/01-index.md

## 1. Overview of Release Architecture

This repository uses an automated versioning and synchronization pipeline where version increments propagate deterministically across Go source files, npm configuration, manifest files, documentation, and release changelogs.

```text
                  ┌────────────────────────┐
                  │ 03-ai-scripts/         │
                  │ 37-bump-version.py     │
                  └───────────┬────────────┘
                              │
     ┌────────────────────────┼────────────────────────┐
     │                        │                        │
┌────▼──────────┐      ┌──────▼───────┐        ┌───────▼────────┐
│ version.json  │      │version/info.go│       │  package.json  │
└───────────────┘      └──────────────┘        └────────────────┘
     │                        │                        │
     └────────────────────────┼────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              │                               │
       ┌──────▼────────┐              ┌───────▼────────┐
       │ CHANGELOG.md  │              │   readme.md    │
       └───────────────┘              └────────────────┘
```

## 2. Where the Version Lives

1. **version.json**:
   The canonical manifest and single source of truth at the repository root. Conforms to `02-spec/01-spec-authoring-guide/14-version-schema.md`. Contains:
   - `version`: Global repository SemVer (e.g. `"2.323.0"`)
   - `backend`: `"inherit"`
   - `frontend`: `"inherit"`
   - `changelog`: `"CHANGELOG.md"`
   - `author`: Attribution metadata (`"Md. Alim Ul Karim"`)
   - `repository`: `"alimtvnetwork/movie-cli-v8"`

2. **version/info.go**:
   The Go source file establishing the runtime version string compiled into the binary:
   ```go
   package version

   var (
       Version   = "v2.323.0"
       BuildTime = "unset"
       GitCommit = "unset"
   )
   ```

3. **package.json**:
   Standard npm configuration maintaining `"version": "2.323.0"` for scripts and package managers.

4. **readme.md & CHANGELOG.md**:
   - `readme.md`: Header badges and pinned version snippets.
   - `CHANGELOG.md`: Detailed release notes, categorized changes, and reproducible install commands.

## 3. How Version Propagates (Sync Pipeline)

1. **03-ai-scripts/37-bump-version.py**:
   Executes automated multi-file synchronization for `--tier <major|minor|patch>`:
   - Computes next SemVer based on current version tags and files.
   - Updates `version.json`, `version/info.go`, `package.json`, `readme.md`, and prepends new release notes to `CHANGELOG.md` and `02-spec/19-main-worker-service/98-changelog.md`.
   - Injects canonical Unix and Windows installation one-liners into `readme.md` and `CHANGELOG.md`.

2. **03-ai-scripts/14-version-sync-checker.py**:
   Validates version alignment across all targets (`version.json`, `package.json`, `version/info.go`). Fails CI if drift occurs.

## 4. Release Branching Ceremony (5-Step Protocol)

Releases MUST follow this strict 5-step sequence:

1. **Step 1: Release Branch Creation**:
   ```bash
   git checkout -b release/vX.Y.Z
   ```
   *Never perform release bumps directly on `main`.*

2. **Step 2: Automated Bump Execution**:
   ```bash
   python 03-ai-scripts/37-bump-version.py --tier minor --scope "<Scope description>"
   python 03-ai-scripts/14-version-sync-checker.py
   ```

3. **Step 3: Commit on Release Branch**:
   ```bash
   git add -A
   git commit -m "release: vX.Y.Z <scope description>"
   ```

4. **Step 4: Create Annotated Git Tag**:
   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   ```

5. **Step 5: Merge to Main & Push to Remote**:
   ```bash
   git checkout main
   git merge release/vX.Y.Z
   git push origin main release/vX.Y.Z
   git push origin vX.Y.Z
   ```

## 5. Non-Negotiable AI Rules

- 🔴 **Zero Manual Timestamp / Date Editing:** Always utilize `03-ai-scripts/37-bump-version.py` or script runners.
- 🔴 **Relative Paths Only:** Never embed machine-specific absolute file paths or `file:///` URIs.
- 🔴 **No Bypassing CI/CD Checks:** Never comment out or disable validation workflows.
- 🔴 **Zero Routine Test Execution:** Never trigger heavy unit test suites unless explicitly commanded by the repository owner.
