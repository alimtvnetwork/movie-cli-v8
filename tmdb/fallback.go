// fallback.go — search fallbacks: progressive query trimming and IMDb-via-web lookup.
package tmdb

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FindResponse mirrors TMDb /find/{external_id} payload.
type FindResponse struct {
	MovieResults []SearchResult `json:"movie_results"`
	TVResults    []SearchResult `json:"tv_results"`
}

// SearchWithFallback tries SearchMulti first, then progressively trims trailing
// tokens from the title, and finally falls back to a web search → IMDb id →
// TMDb /find lookup. Returns the first non-empty result set or nil.
func (c *Client) SearchWithFallback(title string, year int) ([]SearchResult, error) {
	query := title
	if year > 0 {
		query = title + " " + strconv.Itoa(year)
	}
	if results, err := c.SearchMulti(query); err == nil && len(results) > 0 {
		return results, nil
	} else if err != nil && !isEmptyResultErr(err) {
		return nil, err
	}

	if results := c.tryProgressiveTrim(title, year); len(results) > 0 {
		return results, nil
	}

	if results := c.tryOmdbFallback(title, year); len(results) > 0 {
		return results, nil
	}

	if results := c.tryImdbViaWeb(title, year); len(results) > 0 {
		return results, nil
	}

	return nil, nil
}

func isEmptyResultErr(err error) bool {
	// network / auth errors should bubble up; only "no results" is treated as empty.
	return false
}

// tryProgressiveTrim drops the last word from the title repeatedly until a
// match is found or the title is too short.
func (c *Client) tryProgressiveTrim(title string, year int) []SearchResult {
	words := strings.Fields(title)
	for n := len(words) - 1; n >= 2; n-- {
		shorter := strings.Join(words[:n], " ")
		query := shorter
		if year > 0 {
			query = shorter + " " + strconv.Itoa(year)
		}
		if results, err := c.SearchMulti(query); err == nil && len(results) > 0 {
			return results
		}
		if year > 0 {
			if results, err := c.SearchMulti(shorter); err == nil && len(results) > 0 {
				return results
			}
		}
	}
	return nil
}

var imdbIdPattern = regexp.MustCompile(`tt\d{7,10}`)

// tryImdbViaWeb resolves a title via the IMDb-cache-aware fallback chain.
// On a fully warm cache hit it returns a synthetic SearchResult containing
// just the TMDb id + media type — the caller is expected to enrich it via
// the /movie/{id} or /tv/{id} detail endpoints. On a partial hit (IMDb id
// cached but TmdbId not yet resolved) it calls TMDb /find and back-fills the
// cache so the next run is fully warm.
func (c *Client) tryImdbViaWeb(title string, year int) []SearchResult {
	imdbID, cachedTmdbID, cachedMediaType, found := c.lookupImdbCache(title, year)
	if found && imdbID == "" {
		return nil // cached miss — do not hit the web
	}

	if cachedTmdbID > 0 && cachedMediaType != "" {
		return []SearchResult{{ID: cachedTmdbID, MediaType: cachedMediaType}}
	}

	if imdbID == "" {
		imdbID = c.fetchImdbIdFromDuckDuckGo(title, year)
		if imdbID == "" {
			c.storeImdbCache(title, year, "", 0, "")
			return nil
		}
	}

	results := c.lookupByImdbId(imdbID)
	if len(results) == 0 {
		// Store IMDb id but no TmdbId so a future /find retry can succeed.
		c.storeImdbCache(title, year, imdbID, 0, "")
		return nil
	}

	best := results[0]
	c.storeImdbCache(title, year, imdbID, best.ID, best.MediaType)
	return results
}

func (c *Client) lookupImdbCache(title string, year int) (string, int, string, bool) {
	if c.ImdbCache == nil {
		return "", 0, "", false
	}
	imdbID, tmdbID, mediaType, _, found := c.ImdbCache.Look(title, year)
	return imdbID, tmdbID, mediaType, found
}

func (c *Client) storeImdbCache(
	title string, year int, imdbID string, tmdbID int, mediaType string,
) {
	if c.ImdbCache == nil {
		return
	}
	_ = c.ImdbCache.Store(title, year, imdbID, tmdbID, mediaType)
}

// fetchImdbIdFromDuckDuckGo performs the actual HTTP scrape. Always hits the
// network; callers should consult the cache via tryImdbViaWeb instead.
func (c *Client) fetchImdbIdFromDuckDuckGo(title string, year int) string {
	query := title + " imdb"
	if year > 0 {
		query = title + " " + strconv.Itoa(year) + " imdb"
	}
	searchURL := "https://duckduckgo.com/html/?q=" + url.QueryEscape(query)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	req, reqErr := http.NewRequest(http.MethodGet, searchURL, nil)
	if reqErr != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; movie-cli/1.0)")
	resp, getErr := httpClient.Do(req)
	if getErr != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if readErr != nil {
		return ""
	}
	return imdbIdPattern.FindString(string(body))
}

// LookupByImdbId is the exported wrapper around the TMDb /find endpoint
// (external_source=imdb_id). Returns the same shape as a search result so
// callers can treat it identically. Useful for cache backfill tools that
// already have an IMDb id and only need its TMDb counterpart.
func (c *Client) LookupByImdbId(imdbID string) []SearchResult {
	return c.lookupByImdbId(imdbID)
}

func (c *Client) lookupByImdbId(imdbID string) []SearchResult {
	if !c.HasAuth() {
		return nil
	}
	params := url.Values{}
	params.Set("external_source", "imdb_id")
	var resp FindResponse
	if err := c.get(c.buildURL("/find/"+imdbID, params), &resp); err != nil {
		return nil
	}
	out := make([]SearchResult, 0, len(resp.MovieResults)+len(resp.TVResults))
	for i := range resp.MovieResults {
		resp.MovieResults[i].MediaType = "movie"
		out = append(out, resp.MovieResults[i])
	}
	for i := range resp.TVResults {
		resp.TVResults[i].MediaType = "tv"
		out = append(out, resp.TVResults[i])
	}
	return out
}
