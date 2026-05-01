// season.go — Season and Episode DB helpers for TV show tracking.
package db

import "database/sql"

// Season represents a row in the Season table.
type Season struct {
	Name         string
	Overview     string
	PosterPath   string
	AirDate      string
	ID           int64
	MediaID      int64
	SeasonNumber int
	TmdbSeasonID int
	EpisodeCount int
}

// Episode represents a row in the Episode table.
type Episode struct {
	Name          string
	Overview      string
	AirDate       string
	StillPath     string
	WatchedAt     sql.NullString
	ID            int64
	SeasonID      int64
	VoteAvg       float64
	EpisodeNumber int
	TmdbEpisodeID int
	Runtime       int
	IsWatched     bool
}

// InsertSeason upserts a season and returns its canonical PK.
//
// IMPORTANT: Never trust res.LastInsertId() after ON CONFLICT DO UPDATE — on
// the conflict path the driver may return a rowid from a different table on
// the same connection (parallel scans). Always SELECT the canonical PK back.
// See spec/09-app-issues/09-director-fk-stale-lastinsertid.md.
func (d *DB) InsertSeason(s *Season) (int64, error) {
	_, err := d.Exec(`
		INSERT INTO Season (MediaId, SeasonNumber, TmdbSeasonId, Name, Overview, PosterPath, AirDate, EpisodeCount)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (MediaId, SeasonNumber) DO UPDATE SET
			TmdbSeasonId=excluded.TmdbSeasonId, Name=excluded.Name,
			Overview=excluded.Overview, PosterPath=excluded.PosterPath,
			AirDate=excluded.AirDate, EpisodeCount=excluded.EpisodeCount`,
		s.MediaID, s.SeasonNumber, s.TmdbSeasonID, s.Name,
		s.Overview, s.PosterPath, s.AirDate, s.EpisodeCount)
	if err != nil {
		return 0, err
	}
	var seasonID int64
	err = d.QueryRow(
		"SELECT SeasonId FROM Season WHERE MediaId = ? AND SeasonNumber = ?",
		s.MediaID, s.SeasonNumber).Scan(&seasonID)
	return seasonID, err
}

// InsertEpisode upserts an episode and returns its canonical PK.
//
// IMPORTANT: Same rule as InsertSeason — SELECT the canonical PK back instead
// of trusting LastInsertId() after ON CONFLICT DO UPDATE.
func (d *DB) InsertEpisode(e *Episode) (int64, error) {
	_, err := d.Exec(`
		INSERT INTO Episode (SeasonId, EpisodeNumber, TmdbEpisodeId, Name, Overview, AirDate, Runtime, StillPath, VoteAvg)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (SeasonId, EpisodeNumber) DO UPDATE SET
			TmdbEpisodeId=excluded.TmdbEpisodeId, Name=excluded.Name,
			Overview=excluded.Overview, AirDate=excluded.AirDate,
			Runtime=excluded.Runtime, StillPath=excluded.StillPath,
			VoteAvg=excluded.VoteAvg`,
		e.SeasonID, e.EpisodeNumber, e.TmdbEpisodeID, e.Name,
		e.Overview, e.AirDate, e.Runtime, e.StillPath, e.VoteAvg)
	if err != nil {
		return 0, err
	}
	var episodeID int64
	err = d.QueryRow(
		"SELECT EpisodeId FROM Episode WHERE SeasonId = ? AND EpisodeNumber = ?",
		e.SeasonID, e.EpisodeNumber).Scan(&episodeID)
	return episodeID, err
}

// SeasonsByMediaID returns all seasons for a media entry.
func (d *DB) SeasonsByMediaID(mediaID int64) ([]Season, error) {
	rows, err := d.Query(`
		SELECT SeasonId, MediaId, SeasonNumber, COALESCE(TmdbSeasonId, 0),
		       COALESCE(Name, ''), COALESCE(Overview, ''), COALESCE(PosterPath, ''),
		       COALESCE(AirDate, ''), EpisodeCount
		FROM Season WHERE MediaId = ? ORDER BY SeasonNumber`, mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seasons []Season
	for rows.Next() {
		var s Season
		if err := rows.Scan(&s.ID, &s.MediaID, &s.SeasonNumber, &s.TmdbSeasonID,
			&s.Name, &s.Overview, &s.PosterPath, &s.AirDate, &s.EpisodeCount); err != nil {
			return nil, err
		}
		seasons = append(seasons, s)
	}
	return seasons, nil
}

// EpisodesBySeasonID returns all episodes for a season.
func (d *DB) EpisodesBySeasonID(seasonID int64) ([]Episode, error) {
	rows, err := d.Query(`
		SELECT EpisodeId, SeasonId, EpisodeNumber, COALESCE(TmdbEpisodeId, 0),
		       COALESCE(Name, ''), COALESCE(Overview, ''), COALESCE(AirDate, ''),
		       Runtime, COALESCE(StillPath, ''), VoteAvg, IsWatched, WatchedAt
		FROM Episode WHERE SeasonId = ? ORDER BY EpisodeNumber`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var episodes []Episode
	for rows.Next() {
		var e Episode
		if err := rows.Scan(&e.ID, &e.SeasonID, &e.EpisodeNumber, &e.TmdbEpisodeID,
			&e.Name, &e.Overview, &e.AirDate, &e.Runtime, &e.StillPath,
			&e.VoteAvg, &e.IsWatched, &e.WatchedAt); err != nil {
			return nil, err
		}
		episodes = append(episodes, e)
	}
	return episodes, nil
}

// MarkEpisodeWatched marks an episode as watched.
func (d *DB) MarkEpisodeWatched(episodeID int64) error {
	_, err := d.Exec(
		"UPDATE Episode SET IsWatched = 1, WatchedAt = datetime('now') WHERE EpisodeId = ?",
		episodeID)
	return err
}

// MarkEpisodePending reverts an episode back to the unwatched/pending
// state. Name avoids the negative `Unwatched` prefix per
// mem://constraints/boolean-no-negative-words.
func (d *DB) MarkEpisodePending(episodeID int64) error {
	_, err := d.Exec(
		"UPDATE Episode SET IsWatched = 0, WatchedAt = NULL WHERE EpisodeId = ?",
		episodeID)
	return err
}

// FindEpisodeByMediaAndCode resolves an episode from a media id plus a
// season+episode pair. Returns the canonical EpisodeId.
func (d *DB) FindEpisodeByMediaAndCode(mediaID int64, seasonNumber, episodeNumber int) (int64, error) {
	var id int64
	err := d.QueryRow(`
		SELECT e.EpisodeId
		FROM Episode e
		JOIN Season s ON s.SeasonId = e.SeasonId
		WHERE s.MediaId = ? AND s.SeasonNumber = ? AND e.EpisodeNumber = ?`,
		mediaID, seasonNumber, episodeNumber).Scan(&id)
	return id, err
}
