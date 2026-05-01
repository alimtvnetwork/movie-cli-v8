---
name: Director FK error 787 (stale LastInsertId)
description: ensureDirector trusted LastInsertId after INSERT OR IGNORE; on UNIQUE conflict the driver returned a stale rowid from another table, producing MediaDirector FK 787. Fixed by always SELECT-after-IGNORE (Genre pattern). v2.304.0.
type: issue
---

Root cause: `db/director.go ensureDirector` used `res.LastInsertId()` to short-circuit. With `INSERT OR IGNORE` and a UNIQUE(Name) conflict, mattn/go-sqlite3 returns the last successful rowid on the connection — could be from Genre/Media/etc. That stale id was inserted into MediaDirector.DirectorId → SQLite extended error 787 (FK violation).

Fix: Match `db/genre.go EnsureGenre` pattern — INSERT OR IGNORE then unconditional SELECT. Also guard `LinkMediaDirectors` against id<=0.

Rule: Never use LastInsertId() as an existence check after INSERT OR IGNORE. Always SELECT the canonical PK.

Spec: spec/09-app-issues/09-director-fk-stale-lastinsertid.md
