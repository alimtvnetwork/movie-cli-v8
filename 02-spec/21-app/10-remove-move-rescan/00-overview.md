# Remove · Move · Smart Rescan — Spec Overview

> **Sequence**: spec/08-app/10-remove-move-rescan
> Inferred next available slot under `spec/08-app/`. The user's verbatim
> proposed `/spec/YY-movie-cli/`; we placed it inside the existing app
> spec tree to keep all `movie` CLI specs co-located.

## Goals

1. Add `movie rm` (aliases `remove`, `delete`) for removing movies by name,
   filename, or condition expression.
2. Extend `movie move` so it accepts a *selector* (genre, name, condition)
   plus a destination — without breaking the existing interactive form.
3. Make `movie scan` **smart**: reconcile Disk ⇄ JSON ⇄ SQLite so previously
   processed items are not re-fetched from TMDb.

## Mapping onto the existing schema (no parallel `Movie` table)

The user's verbatim proposes a fresh `Movie` table. The shipped project
already has an equivalent **PascalCase `Media` table** (auto-inc `MediaId`),
N-M `MediaGenre` join, `MoveHistory`, and a 15-row `FileAction` enum that
already includes `Delete`, `Restore`, and `Compact`.

To avoid a multi-day breaking migration, this spec maps the verbatim
1:1 onto the existing schema:

| Verbatim                  | Maps to (existing)                                     |
| ------------------------- | ------------------------------------------------------ |
| `Movie`                   | `Media`                                                |
| `MovieId`                 | `MediaId`                                              |
| `Genre` / `MovieGenre`    | `Genre` / `MediaGenre` (already N-M)                   |
| `MovieStatus` lookup      | **NEW** `MediaStatus` lookup + `Media.MediaStatusId`   |
| `Movie.IsDeleted`         | **NEW** `Media.IsDeleted` boolean column               |
| `RemovalHistory`          | Reuse `ActionHistory` row, `FileActionId = Delete (3)` |
| `MoveHistory`             | Already exists                                         |
| `ReconciliationHistory`   | **NEW** table + `ReconciliationActionType` lookup      |

Three new artefacts only:
1. `MediaStatus` lookup (Active / Removed / Moved / Missing)
2. `Media.IsDeleted` + `Media.MediaStatusId` columns (`migrate_v4.go`)
3. `ReconciliationHistory` + `ReconciliationActionType` (`migrate_v5.go`)

## Files in this spec

- `00-overview.md` — this file
- `01-erd.mmd` — Mermaid ERD covering new + impacted tables
- `02-acceptance-criteria.md` — full GIVEN / WHEN / THEN matrix
- `03-condition-expression-grammar.md` — accepted operators / fields / SQL translation
- `remove-command/01-spec.md`
- `move-command/01-spec.md`
- `rescan-reconciliation/01-spec.md`

## Non-goals

- Hard-deleting the on-disk video file (`movie rm` leaves the file on disk;
  a future `--purge` flag is tracked in the plan).
- Cross-folder JSON cache sharing (each scan folder owns its own
  `.movie-output/` cache).
- Replacing the existing interactive `movie move` UX (we keep both forms
  — interactive when 0–1 args, selector mode when 2 args).

## Error handling

Every new code path follows `spec/02-error-manage-spec/04-runtime-error-handling.md`:
- Wrap with `apperror.Wrap(...)` — never `fmt.Errorf`.
- Log via `errlog` so the row also lands in `ErrorLog` table.
- Reconciliation failures of a single item must NOT abort the whole scan.
