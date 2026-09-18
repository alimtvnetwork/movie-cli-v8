---
name: no-tmdb-account-sync
description: NEVER suggest watchlist sync with TMDb user accounts (pull/push). Watchlist export/import uses local JSON only.
type: constraint
---
# Constraint: No TMDb account sync for watchlist

NEVER propose, plan, or implement "Watchlist TMDb sync — pull/push with TMDb accounts" or any feature that signs the user into a TMDb account to sync watchlist data.

**Why:** User explicitly forbade this. Existing `movie watch export` / `movie watch import` (local JSON, no TMDb auth) is the final design.

**How to apply:**
- Do NOT add this idea to plans, suggestions, AC docs, or backlog files.
- If audited code or specs reference it, remove the reference.
- If asked to implement, refuse and point to the existing local JSON sync.
