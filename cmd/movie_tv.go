// movie_tv.go — `movie tv [seasons|episodes|mark|unmark]` parent command.
// Surfaces the Season/Episode tables populated by the scan path
// (see mem://features/tv-season-ingestion).
package cmd

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
	"github.com/spf13/cobra"
)

// episodeCodePattern matches `S<n>E<m>` (case-insensitive). Captures
// season and episode numbers as decimal strings.
var episodeCodePattern = regexp.MustCompile(`(?i)^S(\d{1,3})E(\d{1,3})$`)

// movieTvCmd is the `movie tv` parent.
var movieTvCmd = &cobra.Command{
	Use:   "tv",
	Short: "Browse and track TV show seasons and episodes",
	Long: `Inspect and manage the season/episode data populated by 'movie scan'
for TV titles. Episodes can be marked watched/pending to track progress.`,
	Run: func(cmd *cobra.Command, args []string) { _ = cmd.Help() },
}

// movie tv seasons <id-or-title>
var movieTvSeasonsCmd = &cobra.Command{
	Use:   "seasons <id-or-title>",
	Short: "List all seasons of a TV show",
	Args:  cobra.ExactArgs(1),
	Run:   runTvSeasons,
}

// movie tv episodes <id-or-title> <seasonNumber>
var movieTvEpisodesCmd = &cobra.Command{
	Use:   "episodes <id-or-title> <seasonNumber>",
	Short: "List episodes for one season of a TV show",
	Args:  cobra.ExactArgs(2),
	Run:   runTvEpisodes,
}

// movie tv mark <id-or-title> <S?E?>
var movieTvMarkCmd = &cobra.Command{
	Use:   "mark <id-or-title> <SxxExx>",
	Short: "Mark an episode as watched (e.g. S01E03)",
	Args:  cobra.ExactArgs(2),
	Run:   func(cmd *cobra.Command, args []string) { runTvSetWatched(args, true) },
}

// movie tv unmark <id-or-title> <S?E?>
var movieTvUnmarkCmd = &cobra.Command{
	Use:   "unmark <id-or-title> <SxxExx>",
	Short: "Mark an episode as pending again (e.g. S01E03)",
	Args:  cobra.ExactArgs(2),
	Run:   func(cmd *cobra.Command, args []string) { runTvSetWatched(args, false) },
}

func init() {
	rootCmd.AddCommand(movieTvCmd)
	movieTvCmd.AddCommand(movieTvSeasonsCmd)
	movieTvCmd.AddCommand(movieTvEpisodesCmd)
	movieTvCmd.AddCommand(movieTvMarkCmd)
	movieTvCmd.AddCommand(movieTvUnmarkCmd)
}

// resolveTvMedia opens the DB and resolves the query into a TV media
// row. Returns nil if the resolved media is not a TV show.
func resolveTvMedia(query string) (*db.DB, *db.Media, error) {
	database, err := db.Open()
	if err != nil {
		return nil, nil, apperror.Wrap("open database", err)
	}
	media, resolveErr := resolveMediaByQuery(database, query)
	if resolveErr != nil {
		database.Close()
		return nil, nil, resolveErr
	}
	if media.Type != string(db.MediaTypeTV) {
		database.Close()
		return nil, nil, apperror.New("'%s' is not a TV show (type=%s)",
			media.Title, media.Type)
	}
	return database, media, nil
}

// parseEpisodeCode parses `S01E03` into (1, 3).
func parseEpisodeCode(code string) (int, int, error) {
	match := episodeCodePattern.FindStringSubmatch(code)
	if match == nil {
		return 0, 0, apperror.New("invalid episode code %q (expected SxxExx, e.g. S01E03)", code)
	}
	season, _ := strconv.Atoi(match[1])
	episode, _ := strconv.Atoi(match[2])
	return season, episode, nil
}

// reportTvError prints a friendly error and bails. Centralized so each
// Run wrapper stays tiny.
func reportTvError(err error) {
	errlog.Error("%v", err)
}

// formatYearOrDash returns " (YYYY)" or "" depending on the year.
func formatYearOrDash(year int) string {
	if year <= 0 {
		return ""
	}
	return fmt.Sprintf(" (%d)", year)
}
