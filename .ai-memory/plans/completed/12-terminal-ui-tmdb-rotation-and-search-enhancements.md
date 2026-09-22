# Master Architectural Plan: Terminal UI Overhaul, TMDb Token Rotation, Multi-Engine Search Fallback & SQLite Concurrency

> **Plan Slug:** `12-terminal-ui-tmdb-rotation-and-search-enhancements`  
> **Status:** Complete  
> **Date:** 2026-09-22  
> **Budget:** N = 200 (PHASE_1 = 100, PHASE_2 = 100)  

---

## User Request (Verbatim)

```text
  [45/206] ⭐ 3.7  Chaury Paatham (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\chaury-paatham-2025.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Iron Lung' (year 2026) after fallback chain — inserted with local data only
  [46/206] ⭐ 0.0  Iron Lung (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\iron-lung-2026.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'IF' (year 2024) after fallback chain — inserted with local data only
  [47/206] ⭐ 0.0  IF (2024)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\if-2024.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Heart Eyes' (year 2025) after fallback chain — inserted with local data only
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Jalsa' (year 2022) after fallback chain — inserted with local data only
⚠️  Could not write error to DB: insert error log: database is locked (5) (SQLITE_BUSY)
  [48/206] ⭐ 0.0  Heart Eyes (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\heart-eyes-2025.json
  [49/206] ⭐ 0.0  Jalsa (2022)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\jalsa-2022.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Jatadhara' (year 2025) after fallback chain — inserted with local data only
  [50/206] ⭐ 0.0  Jatadhara (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\jatadhara-2025.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Julie' (year 2004) after fallback chain — inserted with local data only
  [51/206] ⭐ 0.0  Julie (2004)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\julie-2004.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Kaalidas 2' (year 2026) after fallback chain — inserted with local data only
  [52/206] ⭐ 0.0  Kaalidas 2 (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\kaalidas-2-2026.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Kaalidhar Laapata' (year 2025) after fallback chain — inserted with local data only
  [53/206] ⭐ 0.0  Kaalidhar Laapata (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\kaalidhar-laapata-2025.json
     🖼️  Thumbnail saved
     ⭐ 9.2  Batman: Knightfall Part 1: Knightfall
  [54/206] ⭐ 9.2  Batman Knightfall Part 1 Knightfall (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\batman-knightfall-part-1-knightfall-2026.json
     🖼️  Thumbnail saved
     ⭐ 6.4  Ikkis
  [55/206] ⭐ 6.4  Ikkis (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\ikkis-2026.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Hamnet' (year 2025) after fallback chain — inserted with local data only
  [56/206] ⭐ 0.0  Hamnet (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\hamnet-2025.json
     🖼️  Thumbnail saved
     ⭐ 7.1  28 Years Later: The Bone Temple
  [57/206] ⭐ 7.1  28 Years Later The Bone Temple (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\28-years-later-the-bone-temple-2026.json
     🖼️  Thumbnail saved
     ⭐ 5.0  Babita Singh Reporting
  [58/206] ⭐ 5.0  Babita Singh Reporting (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\babita-singh-reporting-2026.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Kartvya' (year 2026) after fallback chain — inserted with local data only
  [59/206] ⭐ 0.0  Kartvya (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\kartvya-2026.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Kuberaa' (year 2025) after fallback chain — inserted with local data only
  [60/206] ⭐ 0.0  Kuberaa (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\kuberaa-2025.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Kuntilanak 3' (year 2022) after fallback chain — inserted with local data only
  [61/206] ⭐ 0.0  Kuntilanak 3 (2022)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\kuntilanak-3-2022.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Laal Rang' (year 2016) after fallback chain — inserted with local data only
  [62/206] ⭐ 0.0  Laal Rang (2016)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\laal-rang-2016.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Ladies First' (year 2026) after fallback chain — inserted with local data only
  [63/206] ⭐ 0.0  Ladies First (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\ladies-first-2026.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Kaantha' (year 2025) after fallback chain — inserted with local data only
  [64/206] ⭐ 0.0  Kaantha (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\kaantha-2025.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Last Shift' (year 2014) after fallback chain — inserted with local data only
  [65/206] ⭐ 0.0  Last Shift (2014)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\last-shift-2014.json
     🖼️  Thumbnail saved
     ⭐ 0.0  We All Wish It Were Easier
  [66/206] ⭐ 0.0  All of You (2024)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\all-of-you-2024.json
     🖼️  Thumbnail saved
     ⭐ 7.3  Dhurandhar: The Revenge
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Locked' (year 2025) after fallback chain — inserted with local data only
  [67/206] ⭐ 7.3  Dhurandhar The Revenge Raw and Undekha (2026)
⚠️  Could not write error to DB: insert error log: database is locked (5) (SQLITE_BUSY)
     🖼️  Thumbnail saved
     🖼️  Thumbnail saved
     ⭐ 3.5  I Know Exactly How You Die
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\dhurandhar-the-revenge-raw-and-undekha-2026.json
  [68/206] ⭐ 0.0  Locked (2025)
     ⭐ 7.5  The Black Phone
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\locked-2025.json
  [69/206] ⭐ 3.5  I Know Exactly How You Die (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\i-know-exactly-how-you-die-2026.json
  [70/206] ⭐ 7.5  Black Phone 2 (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\black-phone-2-2025.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Maanaadu' (year 2021) after fallback chain — inserted with local data only
  [71/206] ⭐ 0.0  Maanaadu (2021)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\maanaadu-2021.json
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Materialists' (year 2025) after fallback chain — inserted with local data only
  [72/206] ⭐ 0.0  Materialists (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\materialists-2025.json
     🖼️  Thumbnail saved
     ⭐ 5.7  Bambi: The Reckoning
  [73/206] ⭐ 5.7  Bambi The Reckoning (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\bambi-the-reckoning-2025.json
     🖼️  Thumbnail saved
     ⭐ 7.5  Dead of Winter
  [74/206] ⭐ 7.5  Dead Of Winter (2025)
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Mercy' (year 2026) after fallback chain — inserted with local data only
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Logout' (year 2025) after fallback chain — inserted with local data only
     🖼️  Thumbnail saved
     ⭐ 5.8  Ek Deewane Ki Deewaniyat
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Lucky Day' (year 2019) after fallback chain — inserted with local data only
     🖼️  Thumbnail saved
     ⭐ 7.9  How to Train Your Dragon
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Mahavatar Narsimha' (year 2025) after fallback chain — inserted with local data only
     🖼️  Thumbnail saved
     ⭐ 7.9  Dept. Q
     🖼️  Thumbnail saved
     ⭐ 5.8  Fleshpot on 42nd Street
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Motor City' (year 2026) after fallback chain — inserted with local data only
     🖼️  Thumbnail saved
     ⭐ 7.4  Pirates of the Caribbean: Dead Man's Chest
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Nobody 2' (year 2025) after fallback chain — inserted with local data only
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Norimberga' (year 2025) after fallback chain — inserted with local data only
     🖼️  Thumbnail saved
     ⭐ 7.7  Evil Dead Burn
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Normal' (year 2026) after fallback chain — inserted with local data only
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\tv\dead-of-winter-2025.json
  [75/206] ⭐ 0.0  Mercy (2026)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\mercy-2026.json
  [76/206] ⭐ 0.0  Logout (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\logout-2025.json
  [77/206] ⭐ 5.8  Ek Deewane Ki Deewaniyat (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\ek-deewane-ki-deewaniyat-2025.json
  [78/206] ⭐ 0.0  Lucky Day (2019)
     🖼️  Thumbnail saved
     ⭐ 6.6  Is This Thing On?
⚠️  [WARN] movie_scan_process.go:167: no TMDb match for 'Mud' (year 2012) after fallback chain — inserted with local data only
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\lucky-day-2019.json
  [79/206] ⭐ 7.9  How to Train Your Dragon (2025)
     📝 JSON metadata saved: Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\how-to-train-your-dragon-2025.json
```

The terminal looks very terrible, and there should be a scan help. And also at the end, it should say, if you want to force run, you can use this flag. If you want to open the browser using the terminal, use this movie UI command, and that will automatically open that report. So these type of things are already missing. You didn't fix it. And the terminal looks like crap. You need to fix it, have some padding. Jason say where, I don't think that the full path is necessary. If it just mentioned from the movie output, that should be enough. Okay. And, yeah. In lot of cases, the HTML that you create, that needs to be changed. I asked several times. You didn't do it. And at the end, you do not show the movie UI command. That is also terrible. You need to suggest to the user so that it's a friendly one. Okay? And also, I want you to include multiple, let's say, API token for the TMDB so that it can switch automatically. Also, you should actually request or have the feature for the DuckDuckGo to search, if the movie is not found there. And also the Google search is required too. So these are the things I think very important that you need to fix it. And also, I do see that you have some wording, which also you need to fix, I believe. Okay. So why the warning is happening? Okay, so are you populating the SQLite database? You have to confirm because this is how we are going to save later on. Yeah, a lot of things are not very good, so I think you need to improve. I asked you several times
```

---

## Subtask Decomposition & Execution Register

| Subtask File | Scope & Tasks | Status |
| :--- | :--- | :--- |
| `01-terminal-ui-and-summary-guidance.md` | Task-01 & Task-02: Padding, relative JSON paths, serialized printing, end-of-scan guidance box, --force flag | Complete |
| `02-tmdb-token-rotation.md` | Task-03: Multiple TMDb API key/token pool, automatic rotation on 429/401 | Complete |
| `03-search-fallbacks-and-matching.md` | Task-04: Relaxed TMDb matching, DuckDuckGo & Google scrapers, persistent caching | Complete |
| `04-sqlite-concurrency-and-lock-fix.md` | Task-05 & Task-07: Single-connection serialization, exponential backoff, DB population integrity | Complete |
| `05-html-report-overhaul.md` | Task-06: Modern glassmorphism report, interactive `movie ui` callout banner & copy action | Complete |

---

## Technical Implementations & Architecture

### 1. Terminal Output Decoupling & Padding Architecture
- Removed all uncoordinated stdout prints from parallel worker goroutines (`enrichFromTMDb`, `downloadScanThumbnail`, `applyTMDbResult`).
- Moved thumbnail confirmation and JSON path printing exclusively into the serialized `commitEnrichedFile` tail, eliminating interleaved console output.
- Formatted JSON metadata paths relative to `.movie-output/...` via `formatScanDisplayPath`.
- Progress printer distinguishes TMDb-matched items (`[k/N] ⭐ rating Title (Year)`) from local-only media (`[k/N] ⚠️  0.0 Title (Year) (local info only)`).
- Added vertical line padding (`\n`) between scanned entries for visual breathing room.
- Rendered high-visibility ASCII/Unicode guidance box at the conclusion of scans showing:
  - `movie ui` (interactive web dashboard)
  - `movie rest --open` (API server + browser)
  - `movie scan --force` (bypass cache and re-enrich all files)
  - `movie rescan`, `movie search`, and `movie ls`
- Added `--force` (`-f`) flag to `movie scan` command.

### 2. Multi-Token Pool & Auto-Rotation
- Implemented `Credential` pool in `tmdb.Client` with `SetCredentials`, `RotateCredential`, and `CredentialCount`.
- Automatic failover rotation upon HTTP 429 (rate limited) and HTTP 401/403 (unauthorized), rebuilding request URLs and retrying seamlessly.
- Config and environment variable discovery supporting `tmdb_api_keys`, `tmdb_tokens`, `TMDB_API_KEYS`, `TMDB_TOKENS`, `TMDB_TOKEN_1`..`9`, `TMDB_API_KEY_1`..`9`.
- Integrated across scan, rescan, and failed rescan flows.

### 3. Multi-Engine Search Matching & Fallbacks
- Added `SearchMovie` and `SearchTV` with explicit release year parameters on `tmdb.Client`.
- Relaxed year matching to exact year and year ± 1 in `SearchWithFallback`.
- SearchMulti with clean title alone, ranked by proximity to release year.
- Progressive word trimming for multi-word titles down to 1 word.
- DuckDuckGo HTML and Lite scrapers with desktop browser user-agent and headers.
- Google Search fallback scraper (`https://www.google.com/search?q={query}+imdb`) to extract IMDb ID (`tt\d{7,10}`).
- Automatic caching in `ImdbLookupCache` table.

### 4. SQLite Concurrency & Lock Elimination
- Enforced `conn.SetMaxOpenConns(1)` in `openAndConfigureDB` to serialize database operations and eliminate multi-connection lock contention.
- Increased `PRAGMA busy_timeout = 10000;` (10 seconds).
- Wrapped `InsertErrorLog` in a 5-attempt retry loop with exponential backoff on `SQLITE_BUSY`.
- Guarded `initScanLogger` DB writer callback with mutex.
- Verified that all media records (matched and local-only) insert into `Media` table with all fields.

### 5. Modern HTML Report Overhaul
- Replaced unstyled `.rest-banner` with a glassmorphism `.hero-guidance-card`.
- Featured prominent `movie ui` launch command with a 1-click clipboard copy button and fallback.
- Added pulsing live indicator dot and aggregate stat badges.
- Added `Local File` status badges for offline or local-only records.
