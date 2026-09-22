// movie_rescan_pool.go — parallel worker pool for media rescan and TMDb fallback searches.
package cmd

import (
	"fmt"
	"sync"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

// rescanFileJob represents a media item that needs TMDb rescan during folder scan.
type rescanFileJob struct {
	Media       *db.Media
	VF          videoFile
	PreSnapshot string
}

// rescanFileResult holds the worker output after TMDb resolution during folder scan.
type rescanFileResult struct {
	Job       rescanFileJob
	IsSuccess bool
}

// runParallelRescanScan dispatches rescan jobs to N workers, each with M threads,
// and serializes DB updates, action history, and output on the calling goroutine.
func runParallelRescanScan(ctx *ScanContext, rescanJobs []rescanFileJob, workers, threads int, cfg ScanLoopConfig) {
	if len(rescanJobs) == 0 {
		return
	}

	totalRoutines := workers * threads
	jobs := make(chan rescanFileJob, len(rescanJobs))
	results := make(chan rescanFileResult, totalRoutines*2)

	var wg sync.WaitGroup

	for w := 1; w <= workers; w++ {
		startRescanWorker(cfg.Client, w, threads, jobs, results, &wg)
	}

	go feedRescanJobs(jobs, rescanJobs, ctx, workers, threads)
	go closeRescanResultsAfter(&wg, results)

	drainRescanResults(ctx, results, len(rescanJobs), cfg)
}

func startRescanWorker(client *tmdb.Client, workerID, threads int, jobs <-chan rescanFileJob, results chan<- rescanFileResult, wg *sync.WaitGroup) {
	for t := 1; t <= threads; t++ {
		wg.Add(1)

		go func(threadID int) {
			defer wg.Done()

			for job := range jobs {
				isSuccess := resolveRescanFromTMDb(client, job.Media)

				results <- rescanFileResult{
					Job:       job,
					IsSuccess: isSuccess,
				}
			}
		}(t)
	}
}

func feedRescanJobs(jobs chan<- rescanFileJob, rescanList []rescanFileJob, ctx *ScanContext, workers, threads int) {
	defer close(jobs)

	announceRescanBatchStart(ctx, rescanList, workers, threads)

	batchSize := workers * threads
	for i, job := range rescanList {
		jobs <- job

		if i == batchSize-1 && len(rescanList) > batchSize {
			announceMidBatchRescanTopUp(ctx, rescanList[batchSize:], workers)
		}
	}
}

func closeRescanResultsAfter(wg *sync.WaitGroup, results chan<- rescanFileResult) {
	wg.Wait()
	close(results)
}

func drainRescanResults(ctx *ScanContext, results <-chan rescanFileResult, total int, cfg ScanLoopConfig) {
	completed := 0
	for res := range results {
		completed++
		commitRescanResult(ctx, res, completed, total, cfg)
	}

	if completed > 0 && !isProgressSuppressed(ctx) {
		fmt.Printf("\n🔄 Rescan batch complete: %d file%s updated in parallel\n\n",
			completed, pluralS(completed))
	}
}

func commitRescanResult(ctx *ScanContext, res rescanFileResult, idx, total int, cfg ScanLoopConfig) {
	ctx.TotalFiles++

	job := res.Job
	if !res.IsSuccess {
		ctx.Skipped++

		if !cfg.IsTableOutput && !cfg.IsJsonOutput {
			printRescanFailed(ctx.TotalFiles, job.Media)
		}

		appendScannedItem(ctx, job.Media)

		return
	}

	persistRescanEntry(ctx.Database, job.Media)

	detail := fmt.Sprintf("Rescan updated: %s", job.Media.CleanTitle)
	_, _ = ctx.Database.InsertActionSimple(db.ActionSimpleInput{
		FileAction: db.FileActionRescanUpdate,
		MediaID:    job.Media.ID,
		Snapshot:   job.PreSnapshot,
		Detail:     detail,
		BatchID:    cfg.BatchID,
	})

	ensureThumbnailInOutputDir(ctx.OutputDir, ctx.Database.BasePath, job.Media.ThumbnailPath)

	if cfg.IsTableOutput {
		printScanTableRow(buildMediaTableRow(ctx.TotalFiles, job.Media, "rescanned"))
	} else if !cfg.IsJsonOutput {
		printRescanSuccess(ctx.TotalFiles, job.Media)
	}

	appendScannedItem(ctx, job.Media)
}

func appendScannedItem(ctx *ScanContext, m *db.Media) {
	ctx.ScannedItems = append(ctx.ScannedItems, *m)
	if m.Type == string(db.MediaTypeMovie) {
		ctx.MovieCount++

		return
	}

	ctx.TVCount++
}

// rescanEntryJob represents a media row for movie rescan / rescan-failed.
type rescanEntryJob struct {
	Media *db.Media
	Index int
}

// rescanEntryResult holds the worker output for movie rescan / rescan-failed.
type rescanEntryResult struct {
	Media     *db.Media
	Index     int
	IsSuccess bool
}

// processRescanEntries processes database media entries using N workers × M threads.
func processRescanEntries(database *db.DB, client *tmdb.Client, entries []db.Media, workers, threads int) (int, int) {
	total := len(entries)
	totalConcurrent := workers * threads
	if totalConcurrent > total {
		totalConcurrent = total
	}

	fmt.Printf("\n🔄 Rescanning %d entries (%d worker%s × %d thread%s = %d concurrent)...\n\n",
		total, workers, pluralS(workers), threads, pluralS(threads), totalConcurrent)

	jobs := make(chan rescanEntryJob, total)
	results := make(chan rescanEntryResult, workers*threads*2)

	var wg sync.WaitGroup

	for w := 1; w <= workers; w++ {
		startRescanEntryWorker(client, w, threads, jobs, results, &wg)
	}

	for i := range entries {
		jobs <- rescanEntryJob{
			Media: &entries[i],
			Index: i + 1,
		}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	updated, failed := 0, 0
	for res := range results {
		yearStr := ""
		if res.Media.Year > 0 {
			yearStr = fmt.Sprintf(" (%d)", res.Media.Year)
		}

		isPersisted := false
		if res.IsSuccess {
			isPersisted = persistRescanEntry(database, res.Media)
		}

		if isPersisted {
			fmt.Printf("  [%d/%d]  %s%s  ✅ ⭐%.1f %s\n",
				res.Index, total, res.Media.CleanTitle, yearStr, res.Media.TmdbRating, res.Media.Genre)
			updated++

			continue
		}

		fmt.Printf("  [%d/%d]  %s%s  ❌ failed\n", res.Index, total, res.Media.CleanTitle, yearStr)
		failed++
	}

	return updated, failed
}

func startRescanEntryWorker(client *tmdb.Client, workerID, threads int, jobs <-chan rescanEntryJob, results chan<- rescanEntryResult, wg *sync.WaitGroup) {
	for t := 1; t <= threads; t++ {
		wg.Add(1)

		go func(threadID int) {
			defer wg.Done()

			for job := range jobs {
				isSuccess := resolveRescanFromTMDb(client, job.Media)

				results <- rescanEntryResult{
					Media:     job.Media,
					Index:     job.Index,
					IsSuccess: isSuccess,
				}
			}
		}(t)
	}
}
