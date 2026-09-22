// errorlog.go — ErrorLog table helpers.
package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

// ErrorLogEntry holds all fields for an error log row.
type ErrorLogEntry struct {
	Timestamp  string
	Level      string
	Source     string
	Function   string
	Command    string
	WorkDir    string
	Message    string
	StackTrace string
}

// InsertErrorLog writes an error entry to the ErrorLog table with retry on busy.
func (d *DB) InsertErrorLog(entry ErrorLogEntry) error {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		_, err := d.Exec(`
			INSERT INTO ErrorLog (Timestamp, Level, Source, Function, Command, WorkDir, Message, StackTrace)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			entry.Timestamp, entry.Level, entry.Source, entry.Function,
			entry.Command, entry.WorkDir, entry.Message, entry.StackTrace,
		)
		if err == nil {
			return nil
		}

		lastErr = err
		errStr := strings.ToLower(err.Error())
		if !strings.Contains(errStr, "locked") && !strings.Contains(errStr, "busy") {
			break
		}

		time.Sleep(time.Duration(25*(attempt+1)) * time.Millisecond)
	}

	return appfault.Wrap("insert error log", lastErr)
}

// RecentErrorLogs returns the most recent N error log entries.
func (d *DB) RecentErrorLogs(limit int) ([]map[string]string, error) {
	rows, err := d.Query(`
		SELECT ErrorLogId, Timestamp, Level, Source, Function, Command, WorkDir, Message, StackTrace
		FROM ErrorLog ORDER BY ErrorLogId DESC LIMIT ?`, limit)
	if err != nil {
		return nil, appfault.Wrap("query error logs", err)
	}
	defer rows.Close()

	var results []map[string]string
	for rows.Next() {
		var id int
		var ts, lvl, src, fn, cmd, wd, msg, st string
		if scanErr := rows.Scan(&id, &ts, &lvl, &src, &fn, &cmd, &wd, &msg, &st); scanErr != nil {
			return nil, appfault.Wrap("scan error log", scanErr)
		}
		results = append(results, map[string]string{
			"id":          fmt.Sprintf("%d", id),
			"timestamp":   ts,
			"level":       lvl,
			"source":      src,
			"function":    fn,
			"command":     cmd,
			"work_dir":    wd,
			"message":     msg,
			"stack_trace": st,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, appfault.Wrap("rows iteration", err)
	}
	return results, nil
}
