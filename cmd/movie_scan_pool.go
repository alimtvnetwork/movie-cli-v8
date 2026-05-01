// movie_scan_pool.go — parallel worker-pool dispatcher for new-file scanning.
//
// Workers do the slow I/O (TMDb search + details + thumbnail download) in
// parallel. The main goroutine serializes DB inserts, sidecar writes, and
// stdout prints — SQLite is single-writer and serialized output reads
// cleanly. The TMDb rate limiter (tmdb.DefaultLimiter) prevents bursts
// from tripping API caps.
package cmd

import (
	"runtime"
	"sync"

	"github.com/alimtvnetwork/movie-cli-v7/cleaner"
	"github.com/alimtvnetwork/movie-cli-v7/db"
)

// MaxScanWorkers is the absolute cap on parallel scan workers.
const MaxScanWorkers = 32

// MinScanWorkers ensures at least one worker even on weird systems.
const MinScanWorkers = 1

// scanWorkers is the user-configurable worker count flag (0 = auto).
var scanWorkers int

// resolveWorkerCount picks the worker count: flag > config > auto (NumCPU*2).
func resolveWorkerCount(database *db.DB) int {
	if scanWorkers > 0 {
		return clampWorkers(scanWorkers)
	}
	if cfg, _ := database.GetConfig("scan_workers"); cfg != "" {
		if n := atoiSafe(cfg); n > 0 {
			return clampWorkers(n)
		}
	}
	return clampWorkers(runtime.NumCPU() * 2)
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
	VF     videoFile
	Result cleaner.Result
	Media  *db.Media // nil if enrichment skipped (e.g. file stat failed)
}

// runParallelNewFileScan dispatches new files to N workers and serializes
// results on the calling goroutine. Returns when all files are processed.
func runParallelNewFileScan(ctx *ScanContext, newFiles []videoFile, workers int) {
	if len(newFiles) == 0 {
		return
	}
	jobs := make(chan videoFile, len(newFiles))
	results := make(chan enrichedFile, workers*2)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go scanWorker(ctx, jobs, results, &wg)
	}
	go feedScanJobs(jobs, newFiles, ctx, workers)
	go closeResultsAfter(&wg, results)

	drainScanResults(ctx, results, len(newFiles))
}

// feedScanJobs streams jobs onto the channel and emits batch-progress
// announcements (initial batch + mid-batch top-ups) for the user.
func feedScanJobs(jobs chan<- videoFile, files []videoFile, ctx *ScanContext, workers int) {
	defer close(jobs)
	announceBatchStart(ctx, files, workers)
	for i, vf := range files {
		jobs <- vf
		if i == workers-1 && len(files) > workers {
			announceMidBatchTopUp(ctx, files[workers:], workers)
		}
	}
}

func closeResultsAfter(wg *sync.WaitGroup, results chan<- enrichedFile) {
	wg.Wait()
	close(results)
}

// scanWorker pulls jobs and runs the slow enrichment phase. DB writes and
// printing happen in the serializer to avoid SQLite contention and
// interleaved stdout.
func scanWorker(ctx *ScanContext, jobs <-chan videoFile, results chan<- enrichedFile, wg *sync.WaitGroup) {
	defer wg.Done()
	for vf := range jobs {
		results <- enrichOneFile(ctx, vf)
	}
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
	mediaID := insertScanMedia(ctx, ef.Media)
	trackScanAction(ctx, TrackScanResult{
		Media: ef.Media, FullPath: ef.VF.FullPath, MediaID: mediaID,
	})
	writeScanJSON(ctx, ef.Media)
	ctx.ScannedItems = append(ctx.ScannedItems, *ef.Media)
	incrementTypeCount(ctx, ef.Media.Type)
}
