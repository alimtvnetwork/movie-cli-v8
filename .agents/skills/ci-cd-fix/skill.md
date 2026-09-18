---
name: ci-cd-fix
description: "Diagnose, fix, and verify CI/CD pipelines and GitHub Actions workflows with zero-storage adherence."
---

# CI/CD Fix Skill

## Overview

Diagnose, fix, and verify CI/CD pipelines, workflows, and build systems. Adhere strictly to the Zero-Storage GitHub Actions Mandate and never disable tests or checks to force a pass.

## Core Directives

1. **Never Disable CI/CD:** Comments, exclusions, or deletes of workflow steps are strictly forbidden. Fix the root cause in application or build code.
2. **Zero-Storage Compliance:** Ensure no routine steps use `actions/upload-artifact`. Diagnostic output must flow to `$GITHUB_STEP_SUMMARY` or stdout.
3. **Targeted Verification:** Test targeted commands locally before pushing. Avoid running full comprehensive test suites unless explicitly commanded by repository owner.
4. **Log Analysis:** Trace failure messages directly from workflow logs to identify precise file and line numbers.
