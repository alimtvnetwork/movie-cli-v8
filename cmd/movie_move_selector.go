// movie_move_selector.go — selector mode for `movie move <selector> <dest>`.
//
// Spec: spec/08-app/10-remove-move-rescan/move-command/01-spec.md
//
// Dispatch contract:
//   - argc < 2 AND no -g flag         → interactive flow (movie_move.go).
//   - argc == 2 OR -g + 1 dest path   → bulk selector flow (this file).
//
// Resolution mirrors movie rm: numeric id, fuzzy title, or condition expr.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v7/apperror"
	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/alimtvnetwork/movie-cli-v7/errlog"
)

var (
	moveGenreSugar string
	moveAssumeYes  bool
)

// isSelectorMoveInvocation returns true when args + flags describe selector mode.
func isSelectorMoveInvocation(args []string) bool {
	if moveGenreSugar != "" && len(args) == 1 {
		return true
	}
	return len(args) == 2
}

func runSelectorMove(args []string) {
	selector, dest, err := buildSelectorAndDest(args)
	if err != nil {
		errlog.Error("move: %v", err)
		return
	}
	database, dbErr := db.Open()
	if dbErr != nil {
		errlog.Error(msgDatabaseError, dbErr)
		return
	}
	defer database.Close()

	ids, resolveErr := resolveMoveTargets(database, selector)
	if resolveErr != nil {
		errlog.Error("move resolve: %v", resolveErr)
		return
	}
	if len(ids) == 0 {
		fmt.Println("📭 No matching media.")
		return
	}
	if !ensureDestDir(dest) {
		return
	}
	previewMoveTargets(database, ids, dest)
	if !confirmSelectorMove(len(ids)) {
		return
	}
	applySelectorMove(database, ids, dest)
}

// buildSelectorAndDest derives (selector, dest) from positional args + -g flag.
func buildSelectorAndDest(args []string) (string, string, error) {
	if moveGenreSugar != "" && len(args) == 1 {
		return "g = " + moveGenreSugar, args[0], nil
	}
	if len(args) == 2 {
		return args[0], args[1], nil
	}
	return "", "", apperrorMoveUsage()
}

func apperrorMoveUsage() error {
	return apperror.New("usage: movie move <selector> <dest>  OR  movie move -g <genre> <dest>")
}

func resolveMoveTargets(database *db.DB, selector string) ([]int64, error) {
	if isConditionExpression(selector) {
		return resolveByCondition(database, selector)
	}
	m, err := resolveMediaByQuery(database, selector)
	if err != nil {
		return nil, err
	}
	return []int64{m.ID}, nil
}

func ensureDestDir(dest string) bool {
	if mkErr := os.MkdirAll(dest, 0755); mkErr != nil {
		errlog.Error("cannot create dest %s: %v", dest, mkErr)
		return false
	}
	return true
}

func previewMoveTargets(database *db.DB, ids []int64, dest string) {
	fmt.Printf("📦 %d media will be moved → %s\n\n", len(ids), dest)
	limit := len(ids)
	if limit > 20 {
		limit = 20
	}
	for i := 0; i < limit; i++ {
		printRmRow(database, ids[i])
	}
	if len(ids) > limit {
		fmt.Printf("  … and %d more\n", len(ids)-limit)
	}
}

const moveConfirmThreshold = 5

func confirmSelectorMove(count int) bool {
	if moveAssumeYes {
		return true
	}
	if count < moveConfirmThreshold {
		return true
	}
	fmt.Printf("\nProceed with move of %d items? [y/N]: ", count)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("❌ Canceled.")
		return false
	}
	return true
}

func applySelectorMove(database *db.DB, ids []int64, dest string) {
	success := 0
	for _, id := range ids {
		if applySingleSelectorMove(database, id, dest) {
			success++
		}
	}
	fmt.Printf("\n✅ Moved %d/%d media.\n", success, len(ids))
}

func applySingleSelectorMove(database *db.DB, id int64, dest string) bool {
	m, err := database.GetMediaByID(id)
	if err != nil {
		errlog.Warn("move: load #%d: %v", id, err)
		return false
	}
	if m.CurrentFilePath == "" {
		errlog.Warn("move: #%d has no current path; skipped", id)
		return false
	}
	newPath := filepath.Join(dest, filepath.Base(m.CurrentFilePath))
	if mvErr := MoveFile(m.CurrentFilePath, newPath); mvErr != nil {
		errlog.Error("move: #%d %v", id, mvErr)
		return false
	}
	updateMoveDB(database, m, newPath)
	fmt.Printf("  ✅ #%d  %s → %s\n", id, m.CurrentFilePath, newPath)
	return true
}

func updateMoveDB(database *db.DB, m *db.Media, newPath string) {
	if updErr := database.UpdateMediaPath(m.ID, newPath); updErr != nil {
		errlog.Warn("move: db path #%d: %v", m.ID, updErr)
	}
	histErr := database.InsertMoveHistory(db.MoveInput{
		MediaID:      m.ID,
		FileActionID: int(db.FileActionMove),
		FromPath:     m.CurrentFilePath,
		ToPath:       newPath,
		OrigName:     filepath.Base(m.CurrentFilePath),
		NewName:      filepath.Base(newPath),
	})
	if histErr != nil {
		errlog.Warn("move: history #%d: %v", m.ID, histErr)
	}
}
