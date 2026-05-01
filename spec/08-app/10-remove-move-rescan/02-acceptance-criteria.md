# Acceptance Criteria

## Remove command (`movie rm` | `remove` | `delete`)

| # | GIVEN | WHEN | THEN |
|---|-------|------|------|
| R1 | a media row "Inception" exists | user runs `movie rm Inception` | row is soft-deleted (`IsDeleted=1`, `MediaStatusId=Removed`); JSON sidecar removed; `report.html` regenerated; one `ActionHistory` row with `FileActionId=3 (Delete)` |
| R2 | `inception.mkv` exists on disk and in DB | `movie delete inception.mkv` | same outcome as R1, matched by `CurrentFilePath` basename |
| R3 | 12 movies have `TmdbRating < 5` | `movie rm "rating < 5"` | all 12 are soft-deleted in one batch; one summary row printed; 12 `ActionHistory` rows share one `BatchId` |
| R4 | no movie matches the target | `movie rm "Inceptioon"` | exit 1, error "no media matched 'Inceptioon'", no DB write |
| R5 | help is requested | `movie rm --help` | help lists all 3 aliases plus 4 examples (name / filename / rating / genre) |
| R6 | a removal exists | `movie undo` | last `Delete` action is reverted (`IsDeleted=0`, `MediaStatusId=Active`); JSON sidecar regenerated |

## Move command (`movie move`)

| # | GIVEN | WHEN | THEN |
|---|-------|------|------|
| M1 | existing interactive flow | `movie move` (no args) | unchanged: interactive picker fires |
| M2 | a thriller exists in `~/Downloads` | `movie move thriller ~/Movies/Thriller` | every Active media tagged Thriller is moved on disk; `Media.CurrentFilePath` updated; one `MoveHistory` row per file; `report.html` regenerated |
| M3 | condition selector | `movie move "rating > 8" ~/Movies/Best` | matching files moved; `MoveHistory` written |
| M4 | sugar flag form | `movie move -g thriller ~/Movies/Thriller` | expands internally to `move "genre = thriller" ~/Movies/Thriller` (M2 outcome) |
| M5 | destination missing | `movie move thriller /nope` | parent created if missing; on permission error abort with clear message; **zero** files moved (atomic batch) |

## Smart rescan / reconciliation

| # | GIVEN | WHEN | THEN |
|---|-------|------|------|
| S1 | `.movie-output/` exists, SQLite empty | `movie scan .` | DB hydrated from JSON sidecars first; only files NOT in JSON hit TMDb; one `ReconciliationHistory` row per hydrated item with `Name=HydratedFromJson` |
| S2 | JSON references a file that no longer exists on disk | `movie scan .` | matching JSON sidecar deleted; SQLite row marked `MediaStatusId=Missing`; row logged with `Name=RemovedMissing` |
| S3 | a brand-new file appears in the folder | `movie scan .` | only that file is processed via TMDb; `ReconciliationHistory` row with `Name=AddedNew` |
| S4 | nothing changed since last scan | `movie scan .` | zero TMDb calls; final summary line prints `Converged: N items, 0 reprocessed` |
| S5 | first-ever scan, no `.movie-output/` | `movie scan .` | falls through to current full-scan flow (no regression) |

## Database

| # | GIVEN | WHEN | THEN |
|---|-------|------|------|
| D1 | fresh DB | `movie scan` runs migrations | `MediaStatus`, `ReconciliationActionType`, `ReconciliationHistory` tables exist; `Media.IsDeleted` and `Media.MediaStatusId` columns exist |
| D2 | enum tables | inspected | `MediaStatus` seeded with 4 rows; `ReconciliationActionType` seeded with 4 rows |
| D3 | every PK | inspected | `Int auto increment`, named `<TableName>Id` |
