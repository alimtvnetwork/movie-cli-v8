---
name: release-management
description: "Orchestrate version releases, release notes, git tags, and release assets without breaking CI/CD."
---

# Release Management Skill

## Overview

Manage semantic versioning, release notes, git tagging, and distribution binaries for `movie-cli-v8`.

## Release Workflow

1. **Version Update:** Bump version in canonical version metadata and sync references across specs.
2. **Changelog & Notes:** Generate release notes summarizing features, fixes, and architectural upgrades.
3. **Git Tagging:** Create annotated semantic tags (`vX.Y.Z`).
4. **Release Assets:** Attach binaries directly to GitHub Releases via `gh release create` / `gh release upload`, keeping workflow storage ephemeral and compliant with zero-storage quotas.
