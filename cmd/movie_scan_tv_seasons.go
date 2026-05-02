// SHARED HELPER: TV-show season/episode ingestion. Walks a TV title's
// seasons via TMDb /tv/{id}/season/{n} and persists each season + its
// episode list using the FK-safe db.InsertSeason / db.InsertEpisode
// helpers (see spec/09-app-issues/10-season-episode-stale-lastinsertid-audit.md).
//
// Called from cmd/movie_scan_process.go::linkScanMediaRelations when the
// resolved Media is of type TV. Failures are logged via errlog.Warn and
// never abort the parent scan.
package cmd

import (
	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
)

// tvIngestInput bundles the params for ingestTVSeasons so we stay under
// the 3-param guideline limit.
type tvIngestInput struct {
	Client  *tmdb.Client
	DB      *db.DB
	Title   string
	MediaID int64
	TmdbID  int
}

// ingestTVSeasons fetches every season for a TV title and persists its
// season + episode rows. Safe to call multiple times for the same Media:
// db.InsertSeason / db.InsertEpisode are idempotent upserts keyed on
// natural unique keys.
func ingestTVSeasons(in tvIngestInput) {
	count := resolveSeasonCount(in.Client, in.TmdbID, in.Title)
	if count <= 0 {
		return
	}
	for n := 1; n <= count; n++ {
		ingestOneSeason(in, n)
	}
}

// resolveSeasonCount fetches /tv/{id} purely to learn number_of_seasons.
// We don't reuse the value from fetchTVDetails because Media has no
// season-count column to carry it across.
func resolveSeasonCount(client *tmdb.Client, tmdbID int, title string) int {
	details, err := client.GetTVDetails(tmdbID)
	if err != nil {
		errlog.Warn("TV season-count lookup failed for '%s' (TMDb %d): %v",
			title, tmdbID, err)
		return 0
	}
	return details.Seasons
}

// ingestOneSeason fetches one season + its episode list and persists
// both. Per-season errors are isolated so a single bad season does not
// abort the rest.
func ingestOneSeason(in tvIngestInput, seasonNumber int) {
	tvSeason, err := in.Client.GetTVSeason(in.TmdbID, seasonNumber)
	if err != nil {
		errlog.Warn("TV season %d fetch failed for '%s': %v",
			seasonNumber, in.Title, err)
		return
	}
	seasonID := persistSeason(in, tvSeason, seasonNumber)
	if seasonID <= 0 {
		return
	}
	persistEpisodes(in.DB, seasonID, tvSeason.Episodes, in.Title)
}

// persistSeason upserts the Season row and returns its canonical PK.
func persistSeason(in tvIngestInput, tvSeason *tmdb.TVSeason, seasonNumber int) int64 {
	row := buildSeasonRow(in.MediaID, tvSeason, seasonNumber)
	id, err := in.DB.InsertSeason(row)
	if err != nil {
		errlog.Warn("Season upsert failed for '%s' season %d: %v",
			in.Title, seasonNumber, err)
		return 0
	}
	return id
}

// buildSeasonRow maps a TMDb TVSeason onto a db.Season ready to upsert.
func buildSeasonRow(mediaID int64, s *tmdb.TVSeason, seasonNumber int) *db.Season {
	return &db.Season{
		MediaID:      mediaID,
		SeasonNumber: seasonNumber,
		TmdbSeasonID: s.ID,
		Name:         s.Name,
		Overview:     s.Overview,
		PosterPath:   s.PosterPath,
		AirDate:      s.AirDate,
		EpisodeCount: len(s.Episodes),
	}
}

// persistEpisodes upserts every episode in a season. Per-episode errors
// are logged but never abort the loop.
func persistEpisodes(database *db.DB, seasonID int64, eps []tmdb.TVEpisode, title string) {
	for i := range eps {
		row := buildEpisodeRow(seasonID, &eps[i])
		if _, err := database.InsertEpisode(row); err != nil {
			errlog.Warn("Episode upsert failed for '%s' S%dE%d: %v",
				title, row.SeasonID, row.EpisodeNumber, err)
		}
	}
}

// buildEpisodeRow maps a TMDb TVEpisode onto a db.Episode ready to upsert.
func buildEpisodeRow(seasonID int64, e *tmdb.TVEpisode) *db.Episode {
	return &db.Episode{
		SeasonID:      seasonID,
		EpisodeNumber: e.EpisodeNumber,
		TmdbEpisodeID: e.ID,
		Name:          e.Name,
		Overview:      e.Overview,
		AirDate:       e.AirDate,
		Runtime:       e.Runtime,
		StillPath:     e.StillPath,
		VoteAvg:       e.VoteAvg,
	}
}
