---
name: LastInsertId anti-pattern lint guard
description: v2.306.0 CI guard scripts/check-lastinsertid-anti-pattern.sh fails if any db/ file mixes INSERT OR IGNORE / ON CONFLICT with res.LastInsertId() in real code; comment-aware (strips `// …`); wired into .github/workflows/ci.yml right after Forbidden term guard
type: feature
---
**Rule (now CI-enforced)**: In `db/`, any file that uses
`INSERT OR IGNORE` or `ON CONFLICT` MUST NOT call `res.LastInsertId()`.
Always SELECT the canonical PK back via the natural unique key.
Reference impl: `db/genre.go::EnsureGenre`.

**Why**: Under parallel scans, mattn/go-sqlite3's `LastInsertId()` after
a no-op `INSERT OR IGNORE` returns the last successful rowid on the
connection — possibly from another table. Caused MediaDirector FK 787
(issue 05) and would have caused the same bug in
Season/Episode (issue 06).

**Script**: `scripts/check-lastinsertid-anti-pattern.sh [root=db]`
- Strips `// …` line comments before matching so warning banners
  that mention LastInsertId don't false-positive.
- Returns exit 1 with `::error file=…::` annotations on violation.
- Verified: catches deliberate regression; passes on current `db/`.

**CI**: wired into `.github/workflows/ci.yml` lint job, right after
the Forbidden term guard.
