# Condition Expression Grammar

Used by `movie rm "<expr>"` and `movie move "<expr>" <dest>`.

## Grammar (EBNF)

```
expr        = term { ("AND" | "OR") term } ;
term        = field op value ;
field       = "Rating" | "Year" | "Genre" | "Duration" | "Size" | "Resolution"
            | "r" | "y" | "g" | "d" | "s" | "res" ;       (* short aliases *)
op          = "<" | "<=" | "=" | ">=" | ">" | "!=" ;
value       = number | quoted-string | bare-word ;
```

Case-insensitive on `field`, `AND`, `OR`. Whitespace is ignored.

## Field → SQL column map

| Public field | SQL column                          | Type   |
|--------------|-------------------------------------|--------|
| Rating / r   | `Media.TmdbRating`                  | REAL   |
| Year / y     | `Media.Year`                        | INT    |
| Duration / d | `Media.Runtime` (minutes)           | INT    |
| Size / s     | `Media.FileSize` (bytes; suffix `MB`/`GB` accepted) | INT |
| Resolution / res | `Media.Resolution` (`720p`, `1080p`, `2160p`) | TEXT |
| Genre / g    | join `MediaGenre`+`Genre`, compares `Genre.Name` | TEXT |

## Translation

The CLI tokenises the expression into `[]ConditionToken`, then builds a
**parameterised** SQL `WHERE` clause — never string-concats user input.

Example:

```
movie rm "rating < 5 AND year >= 2010"
```

becomes

```sql
SELECT MediaId FROM Media
WHERE TmdbRating < ?      -- ?1 = 5
  AND Year       >= ?     -- ?2 = 2010
  AND IsDeleted = 0
```

`Genre` is special — it generates an `EXISTS (SELECT 1 FROM MediaGenre …)`
sub-query, also parameterised.

## Safety rules

1. Reject any unknown field → `apperror.Wrap("unknown field: %s", name)`.
2. Reject any unknown operator → same.
3. Reject empty expression.
4. Cap matches at 10 000 rows per command (configurable via
   `Config.RemoveMaxMatches`) to prevent accidental library-wipe.
5. Always show `--dry-run` output and a `[y/N]` confirm prompt when
   matches > 5, unless `--yes` flag is passed.
