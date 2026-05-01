# `movie move` — Selector Mode Spec

## Backward-compat dispatch

```
argc == 0 or 1   → existing interactive flow (unchanged)
argc == 2        → NEW selector mode: move <selector> <destination>
```

Selector grammar matches `03-condition-expression-grammar.md`. A bare
word that matches a row in `Genre.Name` is shorthand for `genre = <word>`.

## Sugar flag

`movie move -g <genre> <dest>`  →  internally becomes
`movie move "genre = <genre>" <dest>`. Same for `-r`, `-y`, etc.
Implemented as a pre-parse hook so the rest of the pipeline is identical.

## Side effects (per match, batched)

1. Resolve absolute destination path (create parent if missing).
2. Move file via existing `MoveFile` helper (handles cross-drive EXDEV).
3. Update `Media.CurrentFilePath`.
4. Insert `MoveHistory` row (already exists in schema).
5. Insert `ActionHistory` row with `FileActionId = 1 (Move)` for undo.
6. Regenerate `report.html` per affected scan folder.

## Atomicity

If ANY file in the batch fails the pre-flight check (unwritable dest,
not enough space, name collision), abort the entire batch BEFORE any
disk move. Either all-or-nothing.

## Files to create

- `cmd/movie_move_selector.go` — argc dispatcher + sugar flag expander
- `cmd/movie_move_apply.go` — batched mover with pre-flight
- (re-uses) `cmd/movie_condition.go`, `cmd/movie_move_helpers.go`

## Help examples

```
movie move                                  # interactive picker (unchanged)
movie move thriller ~/Movies/Thriller       # genre shorthand
movie move "rating > 8"  ~/Movies/Best
movie move "year < 2000" ~/Movies/Classics
movie move -g horror ~/Movies/Horror
```
