---
name: Acronyms in Go identifiers must be MixedCaps, not all-caps
description: All exported and unexported Go identifiers must use Imdb/Tmdb/Api/Http/Url casing — never IMDb/TMDb/API/HTTP/URL — except for trailing-initialism locals (imdbID, tmdbID) and DB column names which already follow this rule
type: constraint
---

# Acronym MixedCaps Rule

In Go identifiers we **always** write multi-letter acronyms as MixedCaps,
not as their original all-caps form. This applies to types, fields, methods,
functions, constants, and test names.

## Rules

| Acronym | ❌ Wrong | ✅ Right |
|---------|---------|---------|
| IMDb | `IMDbCache`, `LookupByIMDbID`, `fetchIMDbIDFromDuckDuckGo` | `ImdbCache`, `LookupByImdbId`, `fetchImdbIdFromDuckDuckGo` |
| TMDb | `initTMDbClient`, `readTMDbConfigValue`, `tmdbCredentials.APIKey` | `initTmdbClient`, `readTmdbConfigValue`, `tmdbCredentials.ApiKey` |
| API | `APIKey`, `APIURL` | `ApiKey`, `ApiUrl` |
| HTTP | `HTTPClient`, `HTTPError` | `HttpClient`, `HttpError` |
| URL | `URLPattern`, `BaseURL` (mid-word) | `UrlPattern`, `BaseUrl` |

## Intentional exceptions

These are **kept as-is** and are NOT violations:

1. **Trailing-initialism short locals**: `imdbID`, `tmdbID`, `imgURL`,
   `reqURL`, `posterURL`. The `ID`/`URL` is the very last token of an
   unexported short-lived local — Go-idiomatic and easier to read.
2. **DB column names** stored in SQL strings (`ImdbId`, `TmdbId`,
   `LookupKey`) — these are already MixedCaps and do not change.
3. **Prose in comments / Markdown** referring to the *product*: "the TMDb
   API", "IMDb id", "HTTP request" — these are English, not identifiers.
4. **Environment variables** (`TMDB_API_KEY`, `TMDB_TOKEN`) and
   user-facing config keys (`TmdbApiKey`, `TmdbToken`) — external contracts.

## Why

- Consistent with the rest of the codebase (`db.SetImdbLookup`,
  `attachImdbCacheUnless`, `cmd/imdb_cache_adapter.go`).
- Makes future renames trivial — one identifier shape per acronym.
- Stops the recurring back-and-forth where some files use `IMDb` and
  others `Imdb`.

## How to enforce

When adding a new identifier:
- If the name contains `IMDb`, `TMDb`, `API`, `HTTP`, `URL`, `JSON`, `XML`,
  `SQL`, `HTML` mid-word → use `Imdb`, `Tmdb`, `Api`, `Http`, `Url`,
  `Json`, `Xml`, `Sql`, `Html` instead.
- If the acronym is the trailing token of an unexported local — leave it
  in the original all-caps form per Go idiom.

## History

- **v2.107.0** — first piecemeal rename: `IMDbCacheUnless`→`ImdbCacheUnless`.
- **v2.115.0** — wholesale sed sweep across 17 Go files. Spec authored.
- **v2.128.3** — first regression after spec: `doctor/json.go` (`JSONReport`/`JSONFinding`/`PrintJSON`/`toJSON`/`toJSONFindings`) + `cmd/doctor.go` (`doctorJSON`/`emitJSON`). Fixed.

## See also

- Spec: `spec/01-coding-guidelines/03-coding-guidelines-spec/03-golang/09-acronym-naming.md`
- Issue file: `spec/12-ci-cd-pipeline/05-ci-cd-issues/05-acronym-mixedcaps.md`
- Playbook: `mem://ci-cd/01-build-fixes-playbook`

