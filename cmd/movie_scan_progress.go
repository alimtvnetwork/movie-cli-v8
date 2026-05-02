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

// shouldSuppressProgress returns true for machine-readable output formats.
func shouldSuppressProgress(ctx *ScanContext) bool {
	if ctx == nil {
		return true
	}
	if ctx.UseTable {
		return true
	}
	return scanFormat == "json"
}

// announceBatchStart prints the initial batch dispatch line.
func announceBatchStart(ctx *ScanContext, files []videoFile, workers int) {
	if shouldSuppressProgress(ctx) {
		return
	}
	n := len(files)
	if n > workers {
		n = workers
	}
	fmt.Printf("\n🚀 Processing batch of %d (%d worker%s): %s\n",
		n, workers, pluralS(workers), previewVideoTitles(files[:n]))
}

// announceMidBatchTopUp prints when more files enter the queue mid-flight.
func announceMidBatchTopUp(ctx *ScanContext, remaining []videoFile, workers int) {
	if shouldSuppressProgress(ctx) {
		return
	}
	if len(remaining) == 0 {
		return
	}
	fmt.Printf("➕ %d more queued: %s\n",
		len(remaining), previewVideoTitles(remaining))
	_ = workers
}

// announceWorkerCompletion prints a single completion line.
func announceWorkerCompletion(ctx *ScanContext, idx, total int, ef enrichedFile) {
	if shouldSuppressProgress(ctx) {
		return
	}
	rating := ef.Media.TmdbRating
	title := ef.Media.CleanTitle
	year := ""
	if ef.Media.Year > 0 {
		year = fmt.Sprintf(" (%d)", ef.Media.Year)
	}
	fmt.Printf("  [%d/%d] ⭐ %.1f  %s%s\n", idx, total, rating, title, year)
}

// announceBatchSummary prints the parallel-batch wrap-up line.
func announceBatchSummary(ctx *ScanContext, completed int) {
	if shouldSuppressProgress(ctx) {
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
	const max = 42
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
