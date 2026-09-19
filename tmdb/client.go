// Package tmdb provides a client for The Movie Database (TMDb) API.
package tmdb

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

const baseURL = "https://api.themoviedb.org/3"
const imageBaseURL = "https://image.tmdb.org/t/p/w500"

// Sentinel errors for callers to classify TMDb failures.
var (
	ErrAuthInvalid  = errors.New("TMDb API key is invalid")
	ErrAuthMissing  = errors.New("no TMDb API key configured")
	ErrRateLimited  = errors.New("TMDb rate limit exceeded")
	ErrServerError  = errors.New("TMDb is temporarily unavailable")
	ErrNetworkError = errors.New("network error reaching TMDb")
	ErrTimeout      = errors.New("TMDb request timed out")
)

// ImdbCache caches DuckDuckGo→IMDb id lookups (and the resolved TMDb id +
// media type) for the search fallback chain. It is optional; when nil the
// fallback always hits the web AND TMDb /find.
//
// Look returns (imdbID, tmdbID, mediaType, isHit, found):
//   - (id,  >0, "movie"|"tv", true,  true)  fully warm hit — skip both web AND /find.
//   - (id,  0,  "",           true,  true)  IMDb id cached but /find never resolved
//     (or returned nothing) — caller still needs /find.
//   - ("",  0,  "",           false, true)  cached "no match" — skip the web entirely.
//   - ("",  0,  "",           _,     false) no cache entry — caller must hit the web.
//
// Store records a result. Pass tmdbID=0 / mediaType="" when only the IMDb id
// is known. Pass empty imdbID to record a miss.
type ImdbCache interface {
	Look(cleanTitle string, year int) (imdbID string, tmdbID int, mediaType string, isHit, found bool)
	Store(cleanTitle string, year int, imdbID string, tmdbID int, mediaType string) error
}

// Client interacts with the TMDb API.
// Field order is tuned for govet fieldalignment: pointer-heavy fields first,
// then strings (whose trailing length word is non-pointer), so the GC pointer
// scan stops at offset 40 instead of 48 (48 pointer bytes vs 56).
type Client struct {
	HttpClient  *http.Client
	ImdbCache   ImdbCache // optional; persisted lookup cache to skip the web
	ApiKey      string
	AccessToken string
}

// SetImdbCache attaches a persistent cache for DuckDuckGo→IMDb lookups.
// Safe to call with nil to detach.
func (c *Client) SetImdbCache(cache ImdbCache) {
	c.ImdbCache = cache
}

// NewClient creates a new TMDb client from an API key or env vars.
func NewClient(apiKey string) *Client {
	return NewClientWithToken(apiKey, "")
}

// NewClientWithToken creates a TMDb client using either an API key or bearer token.
func NewClientWithToken(apiKey, accessToken string) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("TMDB_API_KEY")
	}
	if accessToken == "" {
		accessToken = os.Getenv("TMDB_TOKEN")
	}
	return &Client{
		ApiKey:      apiKey,
		AccessToken: accessToken,
		HttpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// HasAuth returns true if the client has either an API key or access token.
func (c *Client) HasAuth() bool {
	return c.ApiKey != "" || c.AccessToken != ""
}

// SearchMulti searches for movies and TV shows.
func (c *Client) SearchMulti(query string) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("page", "1")

	var resp searchResponse
	if err := c.get(c.buildURL("/search/multi", params), &resp); err != nil {
		return nil, err
	}

	var filtered []SearchResult
	for i := range resp.Results {
		if resp.Results[i].MediaType == "movie" || resp.Results[i].MediaType == "tv" {
			filtered = append(filtered, resp.Results[i])
		}
	}
	return filtered, nil
}

// GetMovieDetails returns detailed info for a movie.
func (c *Client) GetMovieDetails(tmdbID int) (*MovieDetails, error) {
	var d MovieDetails
	if err := c.get(c.buildURL(fmt.Sprintf("/movie/%d", tmdbID), nil), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// GetTVDetails returns detailed info for a TV show.
func (c *Client) GetTVDetails(tmdbID int) (*TVDetails, error) {
	var d TVDetails
	if err := c.get(c.buildURL(fmt.Sprintf("/tv/%d", tmdbID), nil), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// GetTVSeason returns a TV season with its episode list. Mirrors TMDb's
// /tv/{id}/season/{season_number} endpoint.
func (c *Client) GetTVSeason(tmdbID, seasonNumber int) (*TVSeason, error) {
	var s TVSeason
	url := c.buildURL(fmt.Sprintf("/tv/%d/season/%d", tmdbID, seasonNumber), nil)
	if err := c.get(url, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetMovieCredits returns cast and crew for a movie.
func (c *Client) GetMovieCredits(tmdbID int) (*Credits, error) {
	var cr Credits
	if err := c.get(c.buildURL(fmt.Sprintf("/movie/%d/credits", tmdbID), nil), &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

// GetTVCredits returns cast and crew for a TV show.
func (c *Client) GetTVCredits(tmdbID int) (*Credits, error) {
	var cr Credits
	if err := c.get(c.buildURL(fmt.Sprintf("/tv/%d/credits", tmdbID), nil), &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

// GetMovieVideos returns videos (trailers, teasers) for a movie.
func (c *Client) GetMovieVideos(tmdbID int) ([]VideoResult, error) {
	var resp videosResponse
	if err := c.get(c.buildURL(fmt.Sprintf("/movie/%d/videos", tmdbID), nil), &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// GetTVVideos returns videos (trailers, teasers) for a TV show.
func (c *Client) GetTVVideos(tmdbID int) ([]VideoResult, error) {
	var resp videosResponse
	if err := c.get(c.buildURL(fmt.Sprintf("/tv/%d/videos", tmdbID), nil), &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// TrailerURL finds the best YouTube trailer URL from a list of videos.
func TrailerURL(videos []VideoResult) string {
	for _, v := range videos {
		if v.Site == "YouTube" && v.Type == "Trailer" {
			return "https://www.youtube.com/watch?v=" + v.Key
		}
	}
	for _, v := range videos {
		if v.Site == "YouTube" {
			return "https://www.youtube.com/watch?v=" + v.Key
		}
	}
	return ""
}

// DownloadPoster downloads a poster image and saves it to dst.
func (c *Client) DownloadPoster(posterPath, dst string) error {
	return c.DownloadImage(posterPath, dst, "w500")
}

// DownloadImage downloads an image with the specified size (w500, w780, w1280, original) and saves it to dst.
func (c *Client) DownloadImage(imagePath, dst, size string) error {
	if imagePath == "" {
		return appfault.New("no image available")
	}

	if size == "" {
		size = "w500"
	}

	imgURL := fmt.Sprintf("https://image.tmdb.org/t/p/%s%s", size, imagePath)
	DefaultLimiter().Wait()
	resp, err := c.HttpClient.Get(imgURL)
	if err != nil {
		if IsNetworkError(err) {
			return ErrNetworkError
		}
		return appfault.Wrapf(err, "fetch image %s", imgURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return appfault.New("download image returned status %d", resp.StatusCode)
	}

	if mkErr := os.MkdirAll(filepath.Dir(dst), 0755); mkErr != nil {
		return appfault.Wrapf(mkErr, "create destination directory for %s", dst)
	}

	f, err := os.Create(dst)
	if err != nil {
		return appfault.Wrapf(err, "create image file %s", dst)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return appfault.Wrapf(err, "write image bytes to %s", dst)
	}

	return nil
}

// GetMovieImages queries /movie/{id}/images with multi-language support.
func (c *Client) GetMovieImages(tmdbID int) (*MediaImagesResponse, error) {
	params := url.Values{}
	params.Set("include_image_language", "en,null")

	var resp MediaImagesResponse
	urlPath := fmt.Sprintf("/movie/%d/images", tmdbID)
	if err := c.get(c.buildURL(urlPath, params), &resp); err != nil {
		return nil, appfault.Wrapf(err, "get movie %d images", tmdbID)
	}

	return &resp, nil
}

// GetTVImages queries /tv/{id}/images with multi-language support.
func (c *Client) GetTVImages(tmdbID int) (*MediaImagesResponse, error) {
	params := url.Values{}
	params.Set("include_image_language", "en,null")

	var resp MediaImagesResponse
	urlPath := fmt.Sprintf("/tv/%d/images", tmdbID)
	if err := c.get(c.buildURL(urlPath, params), &resp); err != nil {
		return nil, appfault.Wrapf(err, "get tv %d images", tmdbID)
	}

	return &resp, nil
}

// GetRecommendations returns recommended movies or TV shows.
func (c *Client) GetRecommendations(
	tmdbID int,
	mediaType string,
	page int,
) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))

	var resp searchResponse
	urlPath := fmt.Sprintf("/%s/%d/recommendations", mediaType, tmdbID)
	if err := c.get(c.buildURL(urlPath, params), &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// DiscoverByGenre discovers content by genre ID.
func (c *Client) DiscoverByGenre(mediaType string, genreID int, page int) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("with_genres", fmt.Sprintf("%d", genreID))
	params.Set("sort_by", "popularity.desc")
	params.Set("page", fmt.Sprintf("%d", page))

	var resp searchResponse
	if err := c.get(c.buildURL(fmt.Sprintf("/discover/%s", mediaType), params), &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// Trending returns trending content.
func (c *Client) Trending(mediaType string) ([]SearchResult, error) {
	var resp searchResponse
	if err := c.get(c.buildURL(fmt.Sprintf("/trending/%s/week", mediaType), nil), &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}
