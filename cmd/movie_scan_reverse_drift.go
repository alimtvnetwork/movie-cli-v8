// movie_scan_reverse_drift.go — Per-media drift collector for the
// reverse-sync printer. Keeps the orchestrator slim by isolating the
// reason enum, label table, and grouped renderer.
package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
)

// driftReason enumerates why a media row / sidecar required action.
type driftReason int

const (
	driftMissingSidecar  driftReason = iota // DB row + disk file present, sidecar absent → rewrite
	driftStaleMtime                         // sidecar mtime < DB UpdatedAt → rewrite
	driftSoftDeletedRow                     // IsDeleted=1 with sidecar still present → purge
	driftDiskOrphanFile                     // sidecar present, no DB row + no disk file → purge
	driftMissingDiskFile                    // active DB row, file gone from disk → mark missing
)

const driftMaxLines = 20 // cap per-section output to stay concise

// driftReasonLabel returns the short human label printed in the summary.
func driftReasonLabel(r driftReason) string {
	switch r {
	case driftMissingSidecar:
		return "missing-sidecar"
	case driftStaleMtime:
		return "stale-mtime"
	case driftSoftDeletedRow:
		return "soft-deleted-row"
	case driftDiskOrphanFile:
		return "disk-orphan"
	case driftMissingDiskFile:
		return "file-missing"
	}
	return "unknown"
}

// driftEntry records one per-media drift event. MediaID == 0 for
// orphan-sidecar entries (no DB row).
type driftEntry struct {
	Subject string // sidecar path or media title
	Reason  driftReason
	MediaID int64
}

// driftLog is the collector passed through every reverse-sync step.
type driftLog struct {
	entries []driftEntry
}

func (d *driftLog) add(reason driftReason, mediaID int64, subject string) {
	d.entries = append(d.entries, driftEntry{
		Reason: reason, MediaID: mediaID, Subject: subject,
	})
}

// groupByReason returns a stable-ordered list of (reason, entries).
func (d *driftLog) groupByReason() []driftReasonGroup {
	buckets := make(map[driftReason][]driftEntry)
	for _, e := range d.entries {
		buckets[e.Reason] = append(buckets[e.Reason], e)
	}
	out := make([]driftReasonGroup, 0, len(buckets))
	for r, list := range buckets {
		out = append(out, driftReasonGroup{Reason: r, Entries: list})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Reason < out[j].Reason })
	return out
}

type driftReasonGroup struct {
	Entries []driftEntry
	Reason  driftReason
}

// printDriftSummary prints a per-reason block with up to driftMaxLines
// rows; remaining entries are summarized as "...N more".
func printDriftSummary(d *driftLog) {
	if len(d.entries) == 0 {
		return
	}
	fmt.Println("   drift detail:")
	for _, g := range d.groupByReason() {
		printDriftGroup(g)
	}
}

func printDriftGroup(g driftReasonGroup) {
	fmt.Printf("   ├─ %-16s ×%d\n", driftReasonLabel(g.Reason), len(g.Entries))
	limit := len(g.Entries)
	if limit > driftMaxLines {
		limit = driftMaxLines
	}
	for i := 0; i < limit; i++ {
		fmt.Printf("   │   %s\n", formatDriftEntry(g.Entries[i]))
	}
	if remaining := len(g.Entries) - limit; remaining > 0 {
		fmt.Printf("   │   ...%d more\n", remaining)
	}
}

func formatDriftEntry(e driftEntry) string {
	subject := e.Subject
	if filepath.IsAbs(subject) {
		subject = filepath.Base(subject)
	}
	if e.MediaID > 0 {
		return fmt.Sprintf("#%d %s", e.MediaID, subject)
	}
	return subject
}

// (minInt removed: defined in cmd/movie_history.go)
