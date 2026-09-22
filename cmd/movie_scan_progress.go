// movie_scan_progress.go — batch-aware progress printers for parallel scan.
//
// Three event types are surfaced to the user:
//   - Batch start:   "🚀 Processing batch of N: <preview titles>"
//   - Mid-batch:     "➕ Added M more to the queue: <preview titles>"
//   - Per-completion: "  [k/N] ⭐ <rating> <title>"  (in arrival order)
//
// Suppressed entirely when scanFormat is "json" or "table" so machine
// output stays clean.
package cmd

import (
	"fmt"
	"strings"
)

// PreviewTitleLimit is how many titles to list in batch announcements.
const PreviewTitleLimit = 5

// isProgressSuppressed returns true for machine-readable output formats.
func isProgressSuppressed(ctx *ScanContext) bool {
	if ctx == nil {
		return true
	}

	if ctx.IsTableOutput {
		return true
	}

	return scanFormat == "json"
}

// announceBatchStart prints the initial batch dispatch line.
func announceBatchStart(ctx *ScanContext, files []videoFile, workers, threads int) {
	if isProgressSuppressed(ctx) {
		return
	}

	totalConcurrent := workers * threads
	n := len(files)
	if n > totalConcurrent {
		n = totalConcurrent
	}

	fmt.Printf("\n🚀 Processing batch of %d (%d worker%s × %d thread%s = %d concurrent): %s\n",
		n, workers, pluralS(workers), threads, pluralS(threads), totalConcurrent, previewVideoTitles(files[:n]))
}

// announceMidBatchTopUp prints when more files enter the queue mid-flight.
func announceMidBatchTopUp(ctx *ScanContext, remaining []videoFile, workers int) {
	if isProgressSuppressed(ctx) {
		return
	}

	if len(remaining) == 0 {
		return
	}

	fmt.Printf("➕ %d more queued: %s\n",
		len(remaining), previewVideoTitles(remaining))
	_ = workers
}

// announceRescanBatchStart prints the batch dispatch line for rescanning existing items.
func announceRescanBatchStart(ctx *ScanContext, jobs []rescanFileJob, workers, threads int) {
	if isProgressSuppressed(ctx) {
		return
	}

	totalConcurrent := workers * threads
	n := len(jobs)
	if n > totalConcurrent {
		n = totalConcurrent
	}

	fmt.Printf("\n🔄 Rescanning batch of %d (%d worker%s × %d thread%s = %d concurrent): %s\n",
		n, workers, pluralS(workers), threads, pluralS(threads), totalConcurrent, previewRescanTitles(jobs[:n]))
}

// announceMidBatchRescanTopUp prints when more rescan items enter the queue.
func announceMidBatchRescanTopUp(ctx *ScanContext, remaining []rescanFileJob, workers int) {
	if isProgressSuppressed(ctx) {
		return
	}

	if len(remaining) == 0 {
		return
	}

	fmt.Printf("➕ %d more rescan queued: %s\n",
		len(remaining), previewRescanTitles(remaining))
	_ = workers
}

func previewRescanTitles(jobs []rescanFileJob) string {
	limit := PreviewTitleLimit
	if len(jobs) < limit {
		limit = len(jobs)
	}

	names := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		names = append(names, trimForPreview(jobs[i].Media.CleanTitle))
	}

	if len(jobs) > limit {
		names = append(names, fmt.Sprintf("…(+%d more)", len(jobs)-limit))
	}

	return strings.Join(names, ", ")
}

// announceWorkerCompletion prints a single completion line.
func announceWorkerCompletion(ctx *ScanContext, idx, total int, ef enrichedFile) {
	if isProgressSuppressed(ctx) {
		return
	}

	title := ef.Media.CleanTitle
	year := ""
	if ef.Media.Year > 0 {
		year = fmt.Sprintf(" (%d)", ef.Media.Year)
	}

	if ef.Media.TmdbID > 0 {
		fmt.Printf("  [%d/%d] ⭐ %.1f  %s%s\n", idx, total, ef.Media.TmdbRating, title, year)
	} else {
		fmt.Printf("  [%d/%d] ⚠️  0.0  %s%s (local info only)\n", idx, total, title, year)
	}
}

// announceBatchSummary prints the parallel-batch wrap-up line.
func announceBatchSummary(ctx *ScanContext, completed int) {
	if isProgressSuppressed(ctx) {
		return
	}

	if completed == 0 {
		return
	}

	fmt.Printf("\n✅ Batch complete: %d file%s processed in parallel\n",
		completed, pluralS(completed))
}

// previewVideoTitles returns a comma-joined preview of file names.
func previewVideoTitles(files []videoFile) string {
	limit := PreviewTitleLimit
	if len(files) < limit {
		limit = len(files)
	}
	names := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		names = append(names, trimForPreview(files[i].Name))
	}
	if len(files) > limit {
		names = append(names, fmt.Sprintf("…(+%d more)", len(files)-limit))
	}
	return strings.Join(names, ", ")
}

func trimForPreview(s string) string {
	const limit = 42
	if len(s) <= limit {
		return s
	}
	return s[:limit-1] + "…"
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
