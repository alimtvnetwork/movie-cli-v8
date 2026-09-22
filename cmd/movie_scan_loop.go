// movie_scan_loop.go — main scan processing loop extracted from movie_scan.go.
package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

// runMainScanLoop processes all video files: detects removals, rescans existing, processes new.
func runMainScanLoop(ctx *ScanContext, videoFiles []videoFile, cfg ScanLoopConfig) int {
	database := ctx.Database

	existingMedia, _ := database.GetMediaByScanDir(cfg.ScanDir)
	diskPaths := make(map[string]bool, len(videoFiles))
	for _, vf := range videoFiles {
		diskPaths[vf.FullPath] = true
	}

	removed := removeStaleEntries(RemoveStaleInput{
		Database: database, ExistingMedia: existingMedia, DiskPaths: diskPaths,
		BatchID: cfg.BatchID, Opts: ScanOutputOpts{OutputFormatOpts: OutputFormatOpts{IsJsonOutput: cfg.IsJsonOutput, IsTableOutput: cfg.IsTableOutput}},
	})

	existingPaths := make(map[string]*db.Media, len(existingMedia))
	for i := range existingMedia {
		existingPaths[existingMedia[i].OriginalFilePath] = &existingMedia[i]
	}

	workers := resolveWorkerCount(ctx.Database)
	threads := resolveThreadsPerWorker(ctx.Database)

	newFiles, rescanJobs := splitFilesForScan(ctx, videoFiles, existingPaths, cfg)

	if len(rescanJobs) > 0 {
		runParallelRescanScan(ctx, rescanJobs, workers, threads, cfg)
	}

	if len(newFiles) > 0 {
		runParallelNewFileScan(ctx, newFiles, workers, threads)
	}

	emitJsonItemsIfNeeded(ctx, existingPaths, cfg)

	return removed
}

// splitFilesForScan divides video files into new files, rescan jobs, and complete existing media.
func splitFilesForScan(ctx *ScanContext, videoFiles []videoFile,
	existingPaths map[string]*db.Media, cfg ScanLoopConfig) ([]videoFile, []rescanFileJob) {
	newFiles := make([]videoFile, 0, len(videoFiles))
	rescanJobs := make([]rescanFileJob, 0)

	for _, vf := range videoFiles {
		em, found := existingPaths[vf.FullPath]
		if !found || scanForce {
			newFiles = append(newFiles, vf)

			continue
		}

		needsRescan := cfg.HasTMDb && mediaNeedsRescan(em)
		if needsRescan {
			preSnapshot, _ := db.MediaToJSON(em)
			rescanJobs = append(rescanJobs, rescanFileJob{
				Media:       em,
				VF:          vf,
				PreSnapshot: preSnapshot,
			})

			continue
		}

		handleSkippedMediaDirect(ctx, em, ScanOutputOpts{
			OutputFormatOpts: OutputFormatOpts{
				IsTableOutput: cfg.IsTableOutput,
				IsJsonOutput:  cfg.IsJsonOutput,
			},
		})
	}

	return newFiles, rescanJobs
}

// emitJsonItemsIfNeeded appends per-file JSON entries after all processing.
func emitJsonItemsIfNeeded(ctx *ScanContext, existingPaths map[string]*db.Media, cfg ScanLoopConfig) {
	if !cfg.IsJsonOutput {
		return
	}
	for i := range ctx.ScannedItems {
		status := "existing"
		if existingPaths[ctx.ScannedItems[i].OriginalFilePath] == nil {
			status = "new"
		}
		*cfg.JsonItems = append(*cfg.JsonItems, buildMediaJsonItem(&ctx.ScannedItems[i], status))
	}
}

func removeStaleEntries(input RemoveStaleInput) int {
	var removeIDs []int64
	var removeMedia []*db.Media
	for i := range input.ExistingMedia {
		if !input.DiskPaths[input.ExistingMedia[i].OriginalFilePath] {
			removeIDs = append(removeIDs, input.ExistingMedia[i].ID)
			removeMedia = append(removeMedia, &input.ExistingMedia[i])
		}
	}

	if len(removeIDs) == 0 {
		return 0
	}

	snapshotRemovedMedia(input.Database, removeMedia, input.BatchID)

	delCount, delErr := input.Database.DeleteMediaByIDs(removeIDs)
	if delErr != nil {
		errlog.Warn("Could not remove %d stale entries: %v", len(removeIDs), delErr)
		return 0
	}

	if !input.Opts.IsJsonOutput && !input.Opts.IsTableOutput {
		fmt.Printf("  🗑️  Removed %d entries (files no longer on disk)\n\n", delCount)
	}
	return delCount
}

func snapshotRemovedMedia(database *db.DB, media []*db.Media, scanBatchID string) {
	for _, rm := range media {
		snapshot, snapErr := db.MediaToJSON(rm)
		if snapErr != nil {
			errlog.Warn("Could not snapshot media %d for undo: %v", rm.ID, snapErr)
			continue
		}
		detail := fmt.Sprintf("Scan removed: %s (%s)", rm.CleanTitle, rm.OriginalFilePath)
		_, _ = database.InsertActionSimple(db.ActionSimpleInput{
			FileAction: db.FileActionScanRemove, MediaID: rm.ID,
			Snapshot: snapshot, Detail: detail, BatchID: scanBatchID,
		})
	}
}

func handleSkippedMediaDirect(ctx *ScanContext, em *db.Media, opts ScanOutputOpts) {
	ctx.TotalFiles++
	ctx.Skipped++

	if opts.IsTableOutput {
		printScanTableRow(buildMediaTableRow(ctx.TotalFiles, em, "existing"))
	} else if !opts.IsJsonOutput {
		printSkippedText(ctx.TotalFiles, em)
	}

	ensureThumbnailInOutputDir(ctx.OutputDir, ctx.Database.BasePath, em.ThumbnailPath)

	ctx.ScannedItems = append(ctx.ScannedItems, *em)
	if em.Type == string(db.MediaTypeMovie) {
		ctx.MovieCount++

		return
	}

	ctx.TVCount++
}

// printSkippedText prints a plain-text line for a skipped (already-in-db) media item.
func printSkippedText(index int, em *db.Media) {
	typeIcon := db.TypeIcon(em.Type)
	yearSuffix := ""
	if em.Year > 0 {
		yearSuffix = fmt.Sprintf(" (%d)", em.Year)
	}

	fmt.Printf("\n  %d. %s %s%s [%s]\n", index, typeIcon, em.CleanTitle, yearSuffix, em.Type)
	fmt.Println("     ⏩ Already in database")
}

func printRescanSuccess(idx int, em *db.Media) {
	typeIcon := db.TypeIcon(em.Type)
	fmt.Printf("\n  %d. %s %s", idx, typeIcon, em.CleanTitle)
	if em.Year > 0 {
		fmt.Printf(" (%d)", em.Year)
	}

	fmt.Printf(" [%s]\n", em.Type)
	fmt.Printf("     🔄 Rescanned — ⭐%.1f %s\n", em.TmdbRating, em.Genre)
}

func printRescanFailed(idx int, em *db.Media) {
	fmt.Printf("\n  %d. %s", idx, em.CleanTitle)
	fmt.Printf(" [%s]\n", em.Type)
	fmt.Println("     ⚠️  Rescan failed — kept existing data")
}
