#!/usr/bin/env bash
# check-lastinsertid-anti-pattern.sh
#
# CI guard: forbid the stale-LastInsertId anti-pattern in db/.
#
# Root cause (see spec/09-app-issues/09-director-fk-stale-lastinsertid.md):
# After an `INSERT OR IGNORE` or `INSERT … ON CONFLICT DO UPDATE` whose
# constraint actually fires, mattn/go-sqlite3's res.LastInsertId() returns
# the last successful rowid on the connection — possibly from another table
# under parallel scans. Any helper that uses that value as a foreign key
# can blow up with FK error 787 or, worse, silently attach rows to the
# wrong parent.
#
# Rule: in db/, any function whose Exec call uses INSERT OR IGNORE or
# ON CONFLICT must NOT call res.LastInsertId(). Always SELECT the
# canonical PK back via the natural unique key, like db/genre.go::EnsureGenre.
#
# This script greps every .go file in db/ for the dangerous combo and
# fails the build if it finds one. It deliberately scans per-file rather
# than per-function so it is dialect-free and impossible to game with
# helper indirection.

set -e

ROOT="${1:-db}"

if [ ! -d "$ROOT" ]; then
  echo "::warning::$ROOT/ not found — skipping LastInsertId anti-pattern guard."
  exit 0
fi

violations=0
while IFS= read -r -d '' file; do
  # Only flag files that contain BOTH the danger source and the danger sink.
  if grep -qiE '(INSERT OR IGNORE|ON CONFLICT)' "$file" \
     && grep -q 'LastInsertId' "$file"; then
    echo "::error file=${file}::Stale-LastInsertId anti-pattern: file uses INSERT OR IGNORE / ON CONFLICT AND calls LastInsertId(). SELECT the canonical PK back instead. See spec/09-app-issues/09-director-fk-stale-lastinsertid.md."
    grep -nE 'INSERT OR IGNORE|ON CONFLICT|LastInsertId' "$file" | sed "s|^|  ${file}:|"
    violations=$((violations + 1))
  fi
done < <(find "$ROOT" -type f -name '*.go' -print0)

if [ "$violations" -gt 0 ]; then
  echo ""
  echo "Found $violations file(s) mixing INSERT OR IGNORE / ON CONFLICT with LastInsertId()."
  echo "Fix: drop res.LastInsertId() and SELECT the canonical PK back via the natural unique key."
  echo "Reference impl: db/genre.go::EnsureGenre"
  exit 1
fi

echo "✅ No stale-LastInsertId anti-pattern in $ROOT/."
