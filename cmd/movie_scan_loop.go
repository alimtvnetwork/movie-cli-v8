// movie_scan_loop.go — main scan processing loop extracted from movie_scan.go.
package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
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
		BatchID: cfg.BatchID, Opts: ScanOutputOpts{UseJson: cfg.UseJson, UseTable: cfg.UseTable},
	})

	existingPaths := make(map[string]*db.Media, len(existingMedia))
	for i := range existingMedia {
		existingPaths[existingMedia[i].OriginalFilePath] = &existingMedia[i]
	}

	newFiles := splitNewFromExisting(ctx, videoFiles, existingPaths, cfg)
	dispatchNewFilesParallel(ctx, newFiles)
	emitJsonItemsIfNeeded(ctx, existingPaths, cfg)
	return removed
}

// splitNewFromExisting handles existing-media rescans serially and returns
// the slice of files that need fresh enrichment via the worker pool.
func splitNewFromExisting(ctx *ScanContext, videoFiles []videoFile,
	existingPaths map[string]*db.Media, cfg ScanLoopConfig) []videoFile {
	newFiles := make([]videoFile, 0, len(videoFiles))
	for _, vf := range videoFiles {
		em, found := existingPaths[vf.FullPath]
		if !found {
			newFiles = append(newFiles, vf)
			continue
		}
		processExistingMedia(ctx, ProcessExistingInput{
			EM: em, VF: vf, Client: cfg.Client, Database: ctx.Database,
			Opts:    ScanOutputOpts{UseTable: cfg.UseTable, UseJson: cfg.UseJson},
			BatchID: cfg.BatchID, HasTMDb: cfg.HasTMDb,
		})
	}
	return newFiles
}

// dispatchNewFilesParallel runs new-file enrichment through the worker pool.
func dispatchNewFilesParallel(ctx *ScanContext, newFiles []videoFile) {
	if len(newFiles) == 0 {
		return
	}
	workers := resolveWorkerCount(ctx.Database)
	runParallelNewFileScan(ctx, newFiles, workers)
}

// emitJsonItemsIfNeeded appends per-file JSON entries after all processing.
func emitJsonItemsIfNeeded(ctx *ScanContext, existingPaths map[string]*db.Media, cfg ScanLoopConfig) {
	if !cfg.UseJson {
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

	if !input.Opts.UseJson && !input.Opts.UseTable {
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

func processExistingMedia(ctx *ScanContext, input ProcessExistingInput) {
	ctx.TotalFiles++

	needsRescan := input.HasTMDb && mediaNeedsRescan(input.EM)
	if needsRescan {
		handleRescan(ctx, HandleRescanInput{
			EM: input.EM, Client: input.Client, Database: input.Database,
			Opts: input.Opts, BatchID: input.BatchID,
		})
	}
	if !needsRescan {
		handleSkippedMedia(ctx, input.EM, input.Opts)
	}

	ctx.ScannedItems = append(ctx.ScannedItems, *input.EM)
	if input.EM.Type == string(db.MediaTypeMovie) {
		ctx.MovieCount++
		return
	}
	ctx.TVCount++
}

func handleRescan(ctx *ScanContext, input HandleRescanInput) {
	preSnapshot, _ := db.MediaToJSON(input.EM)
	if !rescanMediaEntry(input.Database, input.Client, input.EM) {
		ctx.Skipped++
		if !input.Opts.UseTable && !input.Opts.UseJson {
			printRescanFailed(ctx.TotalFiles, input.EM)
		}
		return
	}
	detail := fmt.Sprintf("Rescan updated: %s", input.EM.CleanTitle)
	_, _ = input.Database.InsertActionSimple(db.ActionSimpleInput{
		FileAction: db.FileActionRescanUpdate, MediaID: input.EM.ID,
		Snapshot: preSnapshot, Detail: detail, BatchID: input.BatchID,
	})
	if input.Opts.UseTable {
		printScanTableRow(buildMediaTableRow(ctx.TotalFiles, input.EM, "rescanned"))
		return
	}
	if !input.Opts.UseJson {
		printRescanSuccess(ctx.TotalFiles, input.EM)
	}
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

func handleSkippedMedia(ctx *ScanContext, em *db.Media, opts ScanOutputOpts) {
	ctx.Skipped++
	if opts.UseTable {
		printScanTableRow(buildMediaTableRow(ctx.TotalFiles, em, "existing"))
	} else if !opts.UseJson {
		printSkippedText(ctx.TotalFiles, em)
	}
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
