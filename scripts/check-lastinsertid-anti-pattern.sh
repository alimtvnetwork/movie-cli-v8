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
# Implementation note: we strip line comments (`// …`) before matching so
# that doc/banner comments warning future contributors NOT to use
# LastInsertId don't themselves trip the guard. Block comments
# (`/* … */`) are rare in this codebase and intentionally NOT stripped —
# anyone using one is opting in to the lint.

set -e

ROOT="${1:-db}"

if [ ! -d "$ROOT" ]; then
  echo "::warning::$ROOT/ not found — skipping LastInsertId anti-pattern guard."
  exit 0
fi

violations=0
while IFS= read -r -d '' file; do
  # Strip Go line comments so warning banners that mention LastInsertId
  # don't false-positive. sed pattern removes everything from `//` to EOL.
  stripped=$(sed -E 's://.*$::' "$file")

  has_unsafe_sql=$(printf '%s' "$stripped" | grep -ciE 'INSERT OR IGNORE|ON CONFLICT' || true)
  has_lastinsert=$(printf '%s' "$stripped" | grep -c 'LastInsertId' || true)

  if [ "$has_unsafe_sql" -gt 0 ] && [ "$has_lastinsert" -gt 0 ]; then
    echo "::error file=${file}::Stale-LastInsertId anti-pattern: file uses INSERT OR IGNORE / ON CONFLICT AND calls LastInsertId() in real code. SELECT the canonical PK back instead. See spec/09-app-issues/09-director-fk-stale-lastinsertid.md."
    grep -nE 'INSERT OR IGNORE|ON CONFLICT|LastInsertId' "$file" \
      | grep -vE '^\s*[0-9]+:\s*//' \
      | sed "s|^|  ${file}:|"
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
