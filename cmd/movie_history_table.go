// movie_history_table.go — table-formatted output for unified history
package cmd

import (
	"fmt"
	"strings"
)

const (
	historyColID     = 6
	historyColType   = 14
	historyColStatus = 8
	historyColDate   = 19
	historyColDetail = 40
)

func printHistoryTableUnified(records []unifiedRecord) {
	isColor := isColorEnabled()

	fmt.Println()
	printHistoryTableHeader(isColor)
	printHistoryTableRows(records, isColor)
	printHistoryTableDivider("┴")
	fmt.Printf("\n  Total: %d records\n\n", len(records))
}

func printHistoryTableHeader(isColor bool) {
	fmt.Printf("  %s │ %s │ %s │ %s │ %s\n",
		formatHistoryHeaderCell("ID", historyColID, isColor),
		formatHistoryHeaderCell("Type", historyColType, isColor),
		formatHistoryHeaderCell("Status", historyColStatus, isColor),
		formatHistoryHeaderCell("Date", historyColDate, isColor),
		formatHistoryHeaderCell("Detail", historyColDetail, isColor))

	printHistoryTableDivider("┼")
}

func formatHistoryHeaderCell(text string, width int, isColor bool) string {
	padded := fmt.Sprintf("%-*s", width, text)

	return colorText(padded, ansiCyan, isColor)
}

func printHistoryTableDivider(mid string) {
	fmt.Printf("  %s─%s─%s─%s─%s─%s─%s─%s─%s\n",
		strings.Repeat("─", historyColID), mid,
		strings.Repeat("─", historyColType), mid,
		strings.Repeat("─", historyColStatus), mid,
		strings.Repeat("─", historyColDate), mid,
		strings.Repeat("─", historyColDetail))
}

func printHistoryTableRows(records []unifiedRecord, isColor bool) {
	for i := range records {
		prefix := records[i].Source[0:1]
		idStr := fmt.Sprintf("%s-%d", prefix, records[i].ID)
		statusStr := formatHistoryStatus(records[i].IsReverted, isColor)

		fmt.Printf("  %-*s │ %-*s │ %s │ %-*s │ %-*s\n",
			historyColID, idStr,
			historyColType, truncate(records[i].Type, historyColType),
			statusStr,
			historyColDate, truncate(records[i].Timestamp, historyColDate),
			historyColDetail, truncate(records[i].Detail, historyColDetail))
	}
}

func formatHistoryStatus(isReverted, isColor bool) string {
	if isReverted {
		chip := colorText("[rev]", ansiYellow, isColor)

		return chip + strings.Repeat(" ", historyColStatus-5)
	}

	chip := colorText("[ok]", ansiGreen, isColor)

	return chip + strings.Repeat(" ", historyColStatus-4)
}
