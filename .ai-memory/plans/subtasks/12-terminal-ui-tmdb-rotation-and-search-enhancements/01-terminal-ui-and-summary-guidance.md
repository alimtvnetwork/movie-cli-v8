# Subtask 01: Terminal UI Formatting, Line Padding & End-of-Scan Guidance

> **Parent Plan:** `.ai-memory/plans/pending/12-terminal-ui-tmdb-rotation-and-search-enhancements.md`  
> **Status:** Complete  
> **Target Files:**
> - `cmd/movie_scan_json.go`
> - `cmd/movie_scan_progress.go`
> - `cmd/movie_scan_pool.go`
> - `cmd/movie_scan_process.go`
> - `cmd/movie_scan_helpers.go`
> - `cmd/movie_scan_helpers_print.go`

---

## 1. Problem Statement

1. Scanned JSON metadata paths are printed as long absolute filesystem paths (e.g. `Z:\DownloadRelated\DownloadCompletedVm\# Movies\.movie-output\json\movie\...`), which wrap awkwardly and fill the terminal window with noise.
2. In parallel worker mode, `applyTMDbResult` in worker goroutines prints `🖼️ Thumbnail saved` and `⭐ rating title` concurrently while `commitEnrichedFile` in the main goroutine prints `[idx/total]` and `📝 JSON metadata saved: ...`. This causes interleaved, duplicate, and fragmented output.
3. Item output lacks vertical breathing room/padding between entries.
4. When `movie scan` completes, there is no guidance box showing users how to launch the interactive web dashboard (`movie ui` or `movie rest --open`), how to force re-scan (`movie scan --force` / `movie rescan`), or helpful search tips.

---

## 2. Proposed Changes

### A. Short Relative JSON Path (`cmd/movie_scan_json.go`)
- Clean path printing helper: extract `.movie-output/...` from `jsonPath` if present, or format relative to scan root.
- Print clean formatted line: `fmt.Printf("     📝 JSON: %s\n", shortPath)`.

### B. Worker Serialization & Decoupling (`cmd/movie_scan_process.go`, `cmd/movie_scan_pool.go`)
- Remove asynchronous `fmt.Printf` / `fmt.Println` calls from worker goroutines in `applyTMDbResult` and `downloadScanThumbnail`.
- Record enrichment status (`HasThumbnail`, `IsTmdbMatch`) on `enrichedFile`.
- In `commitEnrichedFile` (which executes serially in `drainScanResults`):
  - Print the clean progress header: `[idx/total] ⭐ rating title (year)`.
  - If thumbnail saved: print `     🖼️  Poster saved`.
  - If no TMDb match: print `  [idx/total] ⚠️  title (year) [local info only]`.
  - Print the short relative JSON path.
  - Print a clean vertical spacer line (`fmt.Println()`) between items for visual clarity.

### C. End-of-Scan Guidance Card (`cmd/movie_scan_helpers.go`, `cmd/movie_scan_helpers_print.go`)
- In `printScanFooter`, append a dedicated, high-visibility guidance card:
  - `💡 Next Steps & Useful Commands:`
  - `  🎬 Web Dashboard:   movie ui           (launches interactive browser UI)`
  - `  🌐 REST API Server: movie rest --open  (starts API server on :8086)`
  - `  ⚡ Force Re-scan:    movie scan --force (re-fetches metadata and bypasses cache)`
  - `  🔍 Instant Search:  movie search <q>   (find titles directly in terminal)`
  - `  📋 List Library:    movie ls           (tabular catalog view)`

---

## 3. Verification Criteria
- Run `movie scan` mentally against requirements: no interleaved stdout prints from concurrent workers.
- Relative paths rendered cleanly starting at `.movie-output/...`.
- Clean vertical line separation between scanned entries.
- Guidance card rendered prominently at the end of the scan.
