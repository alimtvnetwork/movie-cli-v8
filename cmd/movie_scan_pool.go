// movie_scan_pool.go — parallel worker-pool dispatcher for new-file scanning.
//
// SHARED: runParallelNewFileScan, resolveWorkerCount, MaxScanWorkers.
// Callers today: movie scan (via movie_scan_loop.go). Future callers
// (movie rescan parallel mode, movie cache backfill) MUST reuse these
// helpers instead of spinning up their own goroutines so the global
// TMDb rate limiter and the NumCPU*2/cap-32 worker policy stay consistent.
//
// Workers do the slow I/O (TMDb search + details + thumbnail download) in
// parallel. The main goroutine serializes DB inserts, sidecar writes, and
// stdout prints — SQLite is single-writer and serialized output reads
// cleanly. The TMDb rate limiter (tmdb.DefaultLimiter) prevents bursts
// from tripping API caps.
package cmd

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/alimtvnetwork/movie-cli-v8/cleaner"
	"github.com/alimtvnetwork/movie-cli-v8/db"
)

// MaxScanWorkers is the absolute cap on parallel scan workers.
const MaxScanWorkers = 32

// MinScanWorkers ensures at least one worker even on weird systems.
const MinScanWorkers = 1

// DefaultThreadsPerWorker is the default number of threads per worker process (4).
const DefaultThreadsPerWorker = 4

// MaxThreadsPerWorker is the maximum allowed threads per worker process.
const MaxThreadsPerWorker = 16

// MinThreadsPerWorker is the minimum allowed threads per worker process.
const MinThreadsPerWorker = 1

// scanWorkers is the user-configurable worker count flag (0 = auto).
var scanWorkers int

// scanThreads is the user-configurable threads per worker flag (0 = auto: 4).
var scanThreads int

// resolveWorkerCount picks the worker count: flag > config > auto (NumCPU*2).
func resolveWorkerCount(database *db.DB) int {
	return resolveWorkers(scanWorkers, database)
}

// resolveThreadsPerWorker picks the threads per worker: flag > config > DefaultThreadsPerWorker (4).
func resolveThreadsPerWorker(database *db.DB) int {
	return resolveThreads(scanThreads, database)
}

func resolveWorkers(flagVal int, database *db.DB) int {
	if flagVal > 0 {
		return clampWorkers(flagVal)
	}

	if cfg, _ := database.GetConfig("scan_workers"); cfg != "" {
		if n := atoiSafe(cfg); n > 0 {
			return clampWorkers(n)
		}
	}

	return clampWorkers(runtime.NumCPU() * 2)
}

func resolveThreads(flagVal int, database *db.DB) int {
	if flagVal > 0 {
		return clampThreads(flagVal)
	}

	if cfg, _ := database.GetConfig("scan_threads"); cfg != "" {
		if n := atoiSafe(cfg); n > 0 {
			return clampThreads(n)
		}
	}

	return DefaultThreadsPerWorker
}

func clampWorkers(n int) int {
	if n > MaxScanWorkers {
		return MaxScanWorkers
	}

	if n < MinScanWorkers {
		return MinScanWorkers
	}

	return n
}

func clampThreads(n int) int {
	if n > MaxThreadsPerWorker {
		return MaxThreadsPerWorker
	}

	if n < MinThreadsPerWorker {
		return MinThreadsPerWorker
	}

	return n
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}

		n = n*10 + int(c-'0')
	}

	return n
}

// enrichedFile is a videoFile that has finished worker-side enrichment and
// is ready for the serializer (DB insert + sidecar + print).
type enrichedFile struct {
	Media  *db.Media // nil if enrichment skipped (e.g. file stat failed)
	VF     videoFile
	Result cleaner.Result
}

// runParallelNewFileScan dispatches new files to N workers, each with M threads,
// and serializes results on the calling goroutine. Returns when all files are processed.
func runParallelNewFileScan(ctx *ScanContext, newFiles []videoFile, workers, threads int) {
	if len(newFiles) == 0 {
		return
	}

	totalRoutines := workers * threads
	jobs := make(chan videoFile, len(newFiles))
	results := make(chan enrichedFile, totalRoutines*2)

	var wg sync.WaitGroup

	for w := 1; w <= workers; w++ {
		startScanWorker(ctx, w, threads, jobs, results, &wg)
	}

	go feedScanJobs(jobs, newFiles, ctx, workers, threads)
	go closeResultsAfter(&wg, results)

	drainScanResults(ctx, results, len(newFiles))
}

// startScanWorker spawns a worker process unit that launches threads concurrent goroutines.
func startScanWorker(ctx *ScanContext, workerID, threads int, jobs <-chan videoFile, results chan<- enrichedFile, wg *sync.WaitGroup) {
	for t := 1; t <= threads; t++ {
		wg.Add(1)

		go func(threadID int) {
			defer wg.Done()

			for vf := range jobs {
				results <- enrichOneFile(ctx, vf)
			}
		}(t)
	}
}

// feedScanJobs streams jobs onto the channel and emits batch-progress
// announcements (initial batch + mid-batch top-ups) for the user.
func feedScanJobs(jobs chan<- videoFile, files []videoFile, ctx *ScanContext, workers, threads int) {
	defer close(jobs)

	announceBatchStart(ctx, files, workers, threads)

	batchSize := workers * threads
	for i, vf := range files {
		jobs <- vf

		if i == batchSize-1 && len(files) > batchSize {
			announceMidBatchTopUp(ctx, files[batchSize:], workers)
		}
	}
}

func closeResultsAfter(wg *sync.WaitGroup, results chan<- enrichedFile) {
	wg.Wait()
	close(results)
}

// enrichOneFile runs the read-only / network-heavy portion of the per-file
// pipeline: clean filename, stat file, TMDb search + details, thumbnail
// download. No DB writes, no stdout prints.
func enrichOneFile(ctx *ScanContext, vf videoFile) enrichedFile {
	result := cleaner.Clean(vf.Name)
	media := buildScanMedia(vf, result)
	if media == nil {
		return enrichedFile{VF: vf, Result: result, Media: nil}
	}

	if ctx.HasTMDb {
		enrichFromTMDb(ctx, media, result)
	}

	return enrichedFile{VF: vf, Result: result, Media: media}
}

// drainScanResults receives enriched files and applies them serially:
// DB insert, JSON sidecar, stdout print, counter update.
func drainScanResults(ctx *ScanContext, results <-chan enrichedFile, total int) {
	completed := 0
	for ef := range results {
		completed++
		commitEnrichedFile(ctx, ef, completed, total)
	}

	announceBatchSummary(ctx, completed)
}

// commitEnrichedFile performs the serialized "tail" of the per-file
// pipeline for one enriched result.
func commitEnrichedFile(ctx *ScanContext, ef enrichedFile, idx, total int) {
	if ef.Media == nil {
		return
	}

	ctx.TotalFiles++
	announceWorkerCompletion(ctx, idx, total, ef)

	if ef.Media.ThumbnailPath != "" {
		fmt.Println("     🖼️  Thumbnail saved")
	}

	mediaID := insertScanMedia(ctx, ef.Media)
	trackScanAction(ctx, TrackScanResult{
		Media:    ef.Media,
		FullPath: ef.VF.FullPath,
		MediaID:  mediaID,
	})
	writeScanJSON(ctx, ef.Media)
	ctx.ScannedItems = append(ctx.ScannedItems, *ef.Media)
	incrementTypeCount(ctx, ef.Media.Type)

	if !isProgressSuppressed(ctx) {
		fmt.Println()
	}
}
