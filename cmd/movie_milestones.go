// movie_milestones.go — implements `movie milestones`: list and filter
// entries from MILESTONES.md by date or keyword.
//
// Filtering:
//
//	--date YYYY-MM-DD   keep entries whose timestamp matches the given day
//	--since YYYY-MM-DD  keep entries on or after the given day
//	--keyword TEXT      case-insensitive substring match against the entry
//	--limit N           cap the number of printed entries (0 = no cap)
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

const (
	milestonesFile     = "MILESTONES.md"
	milestoneTSLayout  = "02-Jan-2006 03:04 PM"
	milestoneDayLayout = "2006-01-02"
)

// tsRegex matches "dd-MMM-YYYY hh:mm AM/PM" inside a milestone bullet.
var tsRegex = regexp.MustCompile(`\b(\d{2}-[A-Za-z]{3}-\d{4} \d{2}:\d{2} (?:AM|PM))\b`)

type milestoneEntry struct {
	Time    time.Time
	Raw     string
	HasTime bool
}

type milestoneFilter struct {
	Date    string
	Since   string
	Keyword string
	Limit   int
}

var milestoneFlags milestoneFilter

var movieMilestonesCmd = &cobra.Command{
	Use:   "milestones",
	Short: "List milestone entries from MILESTONES.md",
	Long: `List entries from MILESTONES.md with optional date or keyword filters.

Examples:
  movie milestones
  movie milestones --keyword scan
  movie milestones --date 2026-04-24
  movie milestones --since 2026-04-01 --keyword run --limit 20`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMilestones(milestonesFile, milestoneFlags)
	},
}

func init() {
	f := movieMilestonesCmd.Flags()
	f.StringVar(&milestoneFlags.Date, "date", "", "exact day to match (YYYY-MM-DD)")
	f.StringVar(&milestoneFlags.Since, "since", "", "earliest day to include (YYYY-MM-DD)")
	f.StringVarP(&milestoneFlags.Keyword, "keyword", "k", "", "case-insensitive substring filter")
	f.IntVarP(&milestoneFlags.Limit, "limit", "n", 0, "max entries to print (0 = no cap)")
}

func runMilestones(path string, f milestoneFilter) error {
	entries, err := readMilestones(path)
	if err != nil {
		return err
	}
	matched, err := filterMilestones(entries, f)
	if err != nil {
		return err
	}
	printMilestones(matched, f.Limit)
	return nil
}

func readMilestones(path string) ([]milestoneEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, apperror.Wrap("open milestones file", err)
	}
	defer file.Close()

	var out []milestoneEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		entry, isMilestone := parseMilestoneLine(line)
		if isMilestone {
			out = append(out, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, apperror.Wrap("read milestones file", err)
	}
	return out, nil
}

func parseMilestoneLine(line string) (milestoneEntry, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "- ") {
		return milestoneEntry{}, false
	}
	entry := milestoneEntry{Raw: trimmed}
	match := tsRegex.FindString(trimmed)
	if match == "" {
		return entry, true
	}
	parsed, err := time.Parse(milestoneTSLayout, match)
	if err == nil {
		entry.Time = parsed
		entry.HasTime = true
	}
	return entry, true
}

func filterMilestones(entries []milestoneEntry, f milestoneFilter) ([]milestoneEntry, error) {
	dayExact, err := parseFilterDay(f.Date)
	if err != nil {
		return nil, err
	}
	daySince, err := parseFilterDay(f.Since)
	if err != nil {
		return nil, err
	}
	keyword := strings.ToLower(strings.TrimSpace(f.Keyword))
	out := make([]milestoneEntry, 0, len(entries))
	for _, e := range entries {
		if entryMatches(e, dayExact, daySince, keyword) {
			out = append(out, e)
		}
	}
	return out, nil
}

func parseFilterDay(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(milestoneDayLayout, raw)
	if err != nil {
		return time.Time{}, apperror.Wrap("parse date filter (want YYYY-MM-DD)", err)
	}
	return t, nil
}

func entryMatches(e milestoneEntry, dayExact, daySince time.Time, keyword string) bool {
	if !matchesKeyword(e, keyword) {
		return false
	}
	if !matchesExactDay(e, dayExact) {
		return false
	}
	return matchesSince(e, daySince)
}

func matchesKeyword(e milestoneEntry, keyword string) bool {
	if keyword == "" {
		return true
	}
	return strings.Contains(strings.ToLower(e.Raw), keyword)
}

func matchesExactDay(e milestoneEntry, day time.Time) bool {
	if day.IsZero() {
		return true
	}
	if !e.HasTime {
		return false
	}
	return sameDay(e.Time, day)
}

func matchesSince(e milestoneEntry, since time.Time) bool {
	if since.IsZero() {
		return true
	}
	if !e.HasTime {
		return false
	}
	return !e.Time.Before(since)
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func printMilestones(entries []milestoneEntry, limit int) {
	total := len(entries)
	if total == 0 {
		fmt.Println("No milestones matched the filters.")
		return
	}
	end := total
	if limit > 0 && limit < total {
		end = limit
	}
	fmt.Printf("📍 Milestones — showing %d of %d\n", end, total)
	fmt.Println(strings.Repeat("─", 60))
	for i := 0; i < end; i++ {
		fmt.Println(entries[i].Raw)
	}
}
