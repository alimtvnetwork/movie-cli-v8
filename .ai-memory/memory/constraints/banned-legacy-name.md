---
name: Banned legacy binary name
description: The previous project codename is permanently banned everywhere in the repo. Binary is 'movie'.
type: constraint
---

The user-facing binary is **`movie`**. The previous legacy codename (and any
derivatives such as `/tmp/<legacy>`, `<LEGACY>_DB`, `.<legacy>/`,
`<legacy>-cli`, `<legacy> v2.x.y`) MUST NEVER reappear anywhere in the repo:
not in code, docs, specs, examples, comments, scripts, configs, or memory.

**Why:** The legacy name was retired. Reintroducing it breaks brand
consistency, confuses users, and fails CI.

**Enforcement:** `.github/workflows/ci.yml` → Lint job → "Forbidden term
guard" step greps the entire repo (excluding `.git`, `.release`,
`node_modules`, `CHANGELOG.md`, and `ci.yml` itself) and fails the build on
any case-insensitive match.

**How to apply:**
- Always write `movie` as the binary name in examples, version banners, env
  vars (`MOVIE_DB`, not `<LEGACY>_DB`), paths (`/tmp/movie`), user-agents
  (`movie-cli/2.x`), and folder conventions (`.movie/`).
- If you find a legacy reference anywhere, fix it immediately — do not defer.
- Never re-suggest the legacy name under any spelling or casing.
