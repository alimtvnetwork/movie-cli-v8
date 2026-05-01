// movie_tv_run.go — Run handlers for `movie tv` subcommands. Kept
// separate from movie_tv.go to respect the ≤200-line file rule.
package cmd

import (
	"fmt"
	"strconv"

	"github.com/alimtvnetwork/movie-cli-v7/apperror"
	"github.com/alimtvnetwork/movie-cli-v7/db"
	"github.com/spf13/cobra"
)

func runTvSeasons(cmd *cobra.Command, args []string) {
	database, media, err := resolveTvMedia(args[0])
	if err != nil {
		reportTvError(err)
		return
	}
	defer database.Close()
	seasons, listErr := database.SeasonsByMediaID(media.ID)
	if listErr != nil {
		reportTvError(listErr)
		return
	}
	printSeasonList(media, seasons)
}

func printSeasonList(media *db.Media, seasons []db.Season) {
	header := fmt.Sprintf("📺 %s%s", media.Title, formatYearOrDash(media.Year))
	fmt.Println(header)
	if len(seasons) == 0 {
		fmt.Println("   (no seasons cached yet — run `movie scan` on the source folder)")
		return
	}
	for _, s := range seasons {
		fmt.Printf("  S%02d  %-30s  %d episodes  %s\n",
			s.SeasonNumber, truncateLine(s.Name, 30), s.EpisodeCount, s.AirDate)
	}
}

func runTvEpisodes(cmd *cobra.Command, args []string) {
	seasonNumber, parseErr := strconv.Atoi(args[1])
	if parseErr != nil {
		reportTvError(parseErr)
		return
	}
	database, media, err := resolveTvMedia(args[0])
	if err != nil {
		reportTvError(err)
		return
	}
	defer database.Close()
	printEpisodesForSeason(database, media, seasonNumber)
}

func printEpisodesForSeason(database *db.DB, media *db.Media, seasonNumber int) {
	season := findSeasonByNumber(database, media.ID, seasonNumber)
	if season == nil {
		reportTvError(apperror.New("S%02d not found for '%s'", seasonNumber, media.Title))
		return
	}
	eps, err := database.EpisodesBySeasonID(season.ID)
	if err != nil {
		reportTvError(err)
		return
	}
	fmt.Printf("📺 %s%s — S%02d %s\n",
		media.Title, formatYearOrDash(media.Year), season.SeasonNumber, season.Name)
	for _, e := range eps {
		fmt.Printf("  %s S%02dE%02d  ⭐ %.1f  %-40s  %s\n",
			watchMarker(e.IsWatched), season.SeasonNumber, e.EpisodeNumber,
			e.VoteAvg, truncateLine(e.Name, 40), e.AirDate)
	}
}

func findSeasonByNumber(database *db.DB, mediaID int64, seasonNumber int) *db.Season {
	all, err := database.SeasonsByMediaID(mediaID)
	if err != nil {
		return nil
	}
	for i := range all {
		if all[i].SeasonNumber == seasonNumber {
			return &all[i]
		}
	}
	return nil
}

func runTvSetWatched(args []string, watched bool) {
	seasonNumber, episodeNumber, parseErr := parseEpisodeCode(args[1])
	if parseErr != nil {
		reportTvError(parseErr)
		return
	}
	database, media, err := resolveTvMedia(args[0])
	if err != nil {
		reportTvError(err)
		return
	}
	defer database.Close()
	applyWatchedState(database, media, seasonNumber, episodeNumber, watched)
}

func applyWatchedState(database *db.DB, media *db.Media,
	seasonNumber, episodeNumber int, watched bool) {
	episodeID, lookupErr := database.FindEpisodeByMediaAndCode(
		media.ID, seasonNumber, episodeNumber)
	if lookupErr != nil || episodeID <= 0 {
		reportTvError(apperror.New("S%02dE%02d not found for '%s'",
			seasonNumber, episodeNumber, media.Title))
		return
	}
	if watched {
		if err := database.MarkEpisodeWatched(episodeID); err != nil {
			reportTvError(err)
			return
		}
		fmt.Printf("✅ Marked %s S%02dE%02d as watched\n",
			media.Title, seasonNumber, episodeNumber)
		return
	}
	if err := database.MarkEpisodePending(episodeID); err != nil {
		reportTvError(err)
		return
	}
	fmt.Printf("↩️  Marked %s S%02dE%02d as pending\n",
		media.Title, seasonNumber, episodeNumber)
}

func watchMarker(isWatched bool) string {
	if isWatched {
		return "✅"
	}
	return "  "
}

func truncateLine(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
