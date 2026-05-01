# `movie rm` — Remove Command Spec

## Aliases

`movie rm <target>` · `movie remove <target>` · `movie delete <target>`

All three resolve to the same cobra command; `--help` lists all aliases.

## Target resolution order

1. **Numeric ID** — if `<target>` parses as int, treat as `MediaId`.
2. **Exact title** match (case-insensitive) on `Media.Title`.
3. **Fuzzy / prefix** match on `Media.CleanTitle`.
4. **Filename** match on `Media.CurrentFilePath` basename.
5. **Condition expression** — if step 2–4 fail AND target contains an
   operator (`< > = != AND OR`), fall through to the expression parser
   from `03-condition-expression-grammar.md`.

This order delegates to the existing `resolveMediaByQuery` helper for
steps 1–4 (already SHARED in `cmd/movie_resolve.go`); step 5 is new.

## Side effects (per match)

1. `Media.IsDeleted = 1`, `Media.MediaStatusId = Removed`.
2. Delete the per-movie JSON sidecar
   (`<scanDir>/.movie-output/json/<type>/<slug>.json`).
3. Append one `ActionHistory` row:
   - `FileActionId = 3 (Delete)`
   - `MediaId = <id>`
   - `MediaSnapshot = <full Media row JSON>`  ← lets `undo` restore it
   - `Detail = "reason=user-rm;target=<original arg>"`
   - `BatchId = <one ID shared across the whole rm batch>`
4. After the batch, regenerate `report.html` for every affected scan
   folder (deduped by parent dir).
5. Print summary: `Removed N media · undo with: movie undo`.

## Flags

| Flag         | Default | Effect |
|--------------|---------|--------|
| `--dry-run`  | false   | Print matches and exit; touch nothing. |
| `--yes`, `-y`| false   | Skip the `[y/N]` confirm when matches > 5. |
| `--purge`    | false   | **Future**: also delete the on-disk video file. Not implemented in this phase. |

## Files to create

- `cmd/movie_rm.go` — cobra command + dispatcher (aliases registered)
- `cmd/movie_rm_resolve.go` — picks resolution path (delegate vs expr)
- `cmd/movie_rm_apply.go` — performs the soft-delete + JSON unlink + audit
- `cmd/movie_condition.go` — tokeniser + SQL builder (SHARED with `move`)
- `cmd/movie_condition_test.go` — table-driven grammar tests

Each file ≤100 lines, each function ≤8 lines per the user's stricter rule
for new code.

## Undo / redo

Already covered by the existing `movie undo` / `movie redo` flow because
we use `ActionHistory` + `FileActionDelete` (which `redo_handlers.go`
knows about). No new undo wiring needed — verify in QA.
