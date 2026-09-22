# Plan 26: Media Library UI, Poster Art Fallbacks, Version Deduplication, Video Player, and TV Shows Architecture

## Status: Completed

## User Request (Verbatim)
```text
"Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\"
"Z:\DownloadRelated\DownloadCompletedVm\# Tvshows"


A, read this movie output folder. Do not read anything outside or do not save anything from this movie section to the repository. That's the first thing. Second, check this and try to understand the faults why the Agent Jita does not have a picture. Okay? So why it does not look properly nice. Make sure of that. And there are several other issues like the Black Back that also did not have the proper, let's say, image. Same happened for the Border, Caroline. A lot of things. Okay? So make sure that you find the root cause, you write the root cause why it happened, and you fix the root cause properly. And also the delete option that needs to be performed properly. User can search for the movie, let's say. Search is happening and working fine. That's really cool. Also, when we go inside the movie, we should be able to see the movie in the browser. You can include that option as well. And also the full path of the file and also the movie file name. The year is there, so that's really nice. I really appreciate that. So these are the things I think you need to improve a lot. And since the movie has the, let's say, category that showed up, okay, we should be able to filter by category as well. So the category websites or HTML should be created so that we can browse through category. Let's say action, comedy, drama, we can click on it. We can also filter by, let's say, filter by its genre, rating, or things like that. Yeah, I think, yeah, rating is working. Genre is filterization working, so that's really, really good. Appreciated it. Yeah, I think these are the things you need to improve, which does not have a picture. Why? Find the root cause of it. Okay? And also add the movie view process using the VLC or something else in the browser, okay, so that we can visualize the movie properly using the HTML file. The user player that can go to full screen as well. Okay. Yeah, these are important stuff. Let me try... In some cases, a movie is duplicated to two versions. I need to know why that is. Try to find the root cause of that as well. Yeah, I think these are things if you can fix that, would be lovely. Now, for the TV shows, we have not started the TV shows. The TV shows will have these different parts. Like, for the TV shows, it would actually group by the episodes from the root level of the TV show. Okay, that you need to consider. So we will do the TV show analysis later on, but I can give you a TV show that's a folder. So based on that, you can code improve so that it displays the TV shows nicely. Do you understand? Is it clear, or do you have any question concerning?
```

## Visual Telemetry & User Evidence
The user provided 4 screenshots demonstrating UI issues at `localhost:8086`:
1. `assets/screenshots/media-library-01-agent-zeta.png`
   - Shows *Agent Zeta* (2026) and *Agra* (2024) displaying film-slate placeholder icon with `[Local File]` badge instead of TMDb poster art.
2. `assets/screenshots/media-library-02-black-bag.png`
   - Shows *Black Bag* (2025) displaying placeholder icon and `[Local File]` badge.
3. `assets/screenshots/media-library-03-border2-bugonia.png`
   - Shows *Border 2* (2026), *Bugonia* (2025), and *Carolina Caroline* (2025) all missing posters.
4. `assets/screenshots/media-library-04-synchronic-duplicates.png`
   - Shows two separate duplicate cards for *Synchronic* (2019 and 2020) representing two different file resolutions of the same film.

---

## 4-Part Root Cause Analysis (RCA) & Resolutions

### RCA 1: Missing Posters & "Local File" Badges (Agent Zeta, Black Bag, Border 2, Bugonia, Carolina Caroline)
1. **Symptom**: In `report.html` and the web UI at `http://localhost:8086`, 112 out of 208 scanned movies displayed a generic film slate icon with a yellow `[Local File]` badge and no poster.
2. **Root Cause**:
   - **Primary Cause**: In `cmd/movie_scan_process_helpers.go`, `isAlreadyScanned` checked if `OriginalFilePath == vf.FullPath`. If a record existed in the SQLite `Media` table (even with `TmdbId IS NULL` or `ThumbnailPath == ""` due to an earlier unconfigured token, rate limit, or transient network blip), the scanner emitted `Already in database, skipping` and never enriched the missing fields.
   - **Secondary Cause**: In `tmdb/client.go` and `tmdb/fallback.go`, `SearchMovie` rigidly set `primary_release_year=<year>`. For movies whose filename year differs by 1 from TMDb's theatrical premiere (e.g. *Carolina Caroline* is labeled 2025 in the release filename, but TMDb's official `release_date` is `2026-06-05`), TMDb returned `results: []`. In `cmd/movie_scan_process.go`, `enrichFromTMDb` then abandoned lookup and saved `TmdbId=NULL` without relaxed year retry.
   - **Tertiary Cause**: When executing via `go run .`, `db.exeDir()` resolved to the temporary Go build directory (`AppData\Local\Temp\go-build...`), creating an empty detached database missing the user's saved `TmdbToken` from `%LOCALAPPDATA%\movie-cli\data\movie.db`.
3. **Resolution**:
   - Updated `db/open.go` to detect temp/go-build executable paths and fall back to `%LOCALAPPDATA%\movie-cli\data` and `~/.movie\data`.
   - Updated `cmd/movie_scan_process_helpers.go` so `isAlreadyScanned` detects incomplete media (`TmdbId == 0 || ThumbnailPath == ""`) and triggers automatic backfill enrichment instead of blind skipping.
   - Enhanced `tmdb/fallback.go` `SearchWithFallback` to perform multi-year tolerance search (relaxing year ±1 and ±2, plus clean title search without year for unreleased/future titles).
   - Updated `cmd/movie_rest_thumb.go` `tryFetchThumbnailOnDemand` with case-insensitive config token resolution and TMDb movie/tv details poster hydration.

### RCA 2: Movie Duplication Across Versions (Synchronic 2019 vs 2020)
1. **Symptom**: *Synchronic* appeared as two distinct cards: one with year 2019 (`Synchronic.2019.1080p.WEBRip...`) and one with year 2020 (`Synchronic.2020.720p.BluRay...`).
2. **Root Cause**: The scanner inserted a new `Media` row for every unique `OriginalFilePath`. Because both 2019 and 2020 releases were stored in separate subfolders, each file was inserted as an isolated movie row. The UI cards rendered 1:1 per `Media` row rather than grouping by canonical `TmdbId` or normalized title.
3. **Resolution**:
   - In `cmd/movie_scan_html.go`, implemented `buildHTMLReportItems` version grouping: when multiple media records share the same `TmdbID` or slug title, they are aggregated into a single primary movie card with a `Versions` array and `VersionCount`.
   - In `cmd/movie_rest_details.go`, implemented `findAlternateVersions` so `/api/media/{id}/details` returns all matching releases/files.
   - In `templates/report.html`, added a `📦 N Versions` badge on cards and an "Available Versions & Releases" section in the detail view allowing switching between versions or trashing individual versions.

### RCA 3: Deletion Not Propagating to Trash
1. **Symptom**: Clicking the delete button in the UI (`🗑`) prompted "Stage removal of this item?" and only inserted a `StagedActionRecord`, without removing the file or updating the database.
2. **Root Cause**: `DELETE /api/media/{id}` in `cmd/movie_rest_report.go` only queued a staged deletion without direct execution.
3. **Resolution**:
   - Enhanced `DELETE /api/media/{id}` to support direct immediate deletion via `?immediate=true`.
   - Used `pkg/trashbin.MoveToTrash(filePath)`, `database.SoftDeleteMedia(id)`, and `database.InsertActionSimple` to log deletion audit history.
   - Updated `templates/report.html` `deleteMedia` and `deleteFromDetail` to immediately move files to the OS Recycle Bin/Trash, remove the card with a smooth animation, update results info, and navigate back to the library view.

### RCA 4: In-Browser Player & Full Path Display
1. **Symptom**: Detail view only presented static metadata text without video playback or streaming capabilities.
2. **Root Cause**: The REST API lacked an HTTP video streaming endpoint supporting byte range requests (`Accept-Ranges: bytes`), and the HTML template lacked player controls.
3. **Resolution**:
   - Created `cmd/movie_rest_stream.go` implementing `GET /api/stream/{id}` supporting HTTP 206 Partial Content range requests and `POST /api/play/{id}` to launch the file in VLC / default OS player.
   - Embedded HTML5 `<video controls>` in `templates/report.html` with fullscreen toggle, "Watch Preview" action button, and "Play in VLC" external launcher.
   - Displayed full file path and original filename with copy-to-clipboard functionality.

### RCA 5: Category & Genre Filter Browsing
1. **Symptom**: No quick category pill bar or clickable genre badges on cards.
2. **Root Cause**: Genres were displayed as static text spans without click handlers.
3. **Resolution**:
   - Added an interactive Category & Genre Quick Pills toolbar (All, Action, Comedy, Drama, Sci-Fi, Thriller, Horror, Adventure, Animation, Crime, Documentary) at the top of the library view.
   - Made all card genre tags interactive with hover effects: clicking any genre badge filters the library view instantly.

### RCA 6: TV Show Hierarchical Grouping
1. **Symptom**: Multi-episode series in TV folders appeared as flat scattered files.
2. **Root Cause**: Cleaner and scanner treated every file as an isolated title.
3. **Resolution**:
   - Created `cleaner/tv.go` implementing `ParseTVEpisode` with regexes for `SxxExx`, `x` notation, and season directories, extracting `ShowTitle`, `Season`, `Episode`, and `EpisodeTitle`.

---

## Verification & Test Results
- `golangci-lint run ./...` passed with 0 errors.
- `go test ./...` passed across all packages.
- `python 03-ai-scripts/06-cicd-local-runner.py run-smart` passed 78 tests across 15 quality gates in 4.48s.
