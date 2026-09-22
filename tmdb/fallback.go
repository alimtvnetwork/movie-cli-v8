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

const desktopUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// SearchWithFallback tries year-specific search first (exact and year ±1),
// clean title search, progressive trimming, OMDB, and finally web search
// fallback (DuckDuckGo + Google Search) → IMDb id → TMDb /find lookup.
func (c *Client) SearchWithFallback(title string, year int) ([]SearchResult, error) {
	// Tier 1: Year-specific searches with exact year and year ± 1
	if year > 0 {
		if results, err := c.SearchMovie(title, year); err == nil && len(results) > 0 {
			return results, nil
		}
		if results, err := c.SearchTV(title, year); err == nil && len(results) > 0 {
			return results, nil
		}

		// Relax year ± 1 and ± 2 (festival vs theatrical release year discrepancies)
		for _, offset := range []int{1, -1, 2, -2} {
			if results, err := c.SearchMovie(title, year+offset); err == nil && len(results) > 0 {
				return results, nil
			}

			if results, err := c.SearchTV(title, year+offset); err == nil && len(results) > 0 {
				return results, nil
			}
		}
	}

	// Tier 2: SearchMovie with clean title alone (no year parameter)
	if results, err := c.SearchMovie(title, 0); err == nil && len(results) > 0 {
		if year > 0 {
			ranked := rankResultsByYear(results, year)

			return ranked, nil
		}

		return results, nil
	}

	// Tier 2b: SearchMulti with clean title alone (no year in query text string)
	if results, err := c.SearchMulti(title); err == nil && len(results) > 0 {
		if year > 0 {
			ranked := rankResultsByYear(results, year)

			return ranked, nil
		}

		return results, nil
	} else if err != nil && !isEmptyResultErr(err) {
		return nil, err
	}

	// Tier 3: If year was provided, try SearchMulti with title + year string
	if year > 0 {
		query := title + " " + strconv.Itoa(year)
		if results, err := c.SearchMulti(query); err == nil && len(results) > 0 {
			return results, nil
		}
	}

	// Tier 4: Progressive word trimming for multi-word titles
	if results := c.tryProgressiveTrim(title, year); len(results) > 0 {
		return results, nil
	}

	// Tier 5: OMDB fallback
	if results := c.tryOmdbFallback(title, year); len(results) > 0 {
		return results, nil
	}

	// Tier 6: Multi-engine web search (DuckDuckGo + Google Search fallback)
	if results := c.tryImdbViaWeb(title, year); len(results) > 0 {
		return results, nil
	}

	return nil, nil
}

func rankResultsByYear(results []SearchResult, targetYear int) []SearchResult {
	if len(results) <= 1 || targetYear <= 0 {
		return results
	}

	var exact []SearchResult
	var closeMatch []SearchResult
	var other []SearchResult

	for i := range results {
		yStr := results[i].GetYear()
		if len(yStr) >= 4 {
			if y, err := strconv.Atoi(yStr[:4]); err == nil {
				diff := y - targetYear
				if diff == 0 {
					exact = append(exact, results[i])
					continue
				}
				if diff == 1 || diff == -1 {
					closeMatch = append(closeMatch, results[i])
					continue
				}
			}
		}
		other = append(other, results[i])
	}

	if len(exact) > 0 {
		exact = append(exact, closeMatch...)
		return append(exact, other...)
	}
	if len(closeMatch) > 0 {
		return append(closeMatch, other...)
	}

	return results
}

func isEmptyResultErr(err error) bool {
	// network / auth errors should bubble up; only "no results" is treated as empty.
	return false
}

// tryProgressiveTrim drops trailing tokens from the title repeatedly until a
// match is found or the title is too short.
func (c *Client) tryProgressiveTrim(title string, year int) []SearchResult {
	words := strings.Fields(title)
	if len(words) <= 1 {
		return nil
	}

	for n := len(words) - 1; n >= 1; n-- {
		shorter := strings.Join(words[:n], " ")
		if year > 0 {
			if results, err := c.SearchMovie(shorter, year); err == nil && len(results) > 0 {
				return results
			}
			if results, err := c.SearchTV(shorter, year); err == nil && len(results) > 0 {
				return results
			}
		}
		if results, err := c.SearchMulti(shorter); err == nil && len(results) > 0 {
			return results
		}
	}

	return nil
}

var imdbIdPattern = regexp.MustCompile(`tt\d{7,10}`)

// tryImdbViaWeb resolves a title via the IMDb-cache-aware fallback chain.
func (c *Client) tryImdbViaWeb(title string, year int) []SearchResult {
	imdbID, cachedTmdbID, cachedMediaType, found := c.lookupImdbCache(title, year)
	if found && imdbID == "" {
		return nil // cached miss — do not hit the web
	}

	if cachedTmdbID > 0 && cachedMediaType != "" {
		return []SearchResult{{ID: cachedTmdbID, MediaType: cachedMediaType}}
	}

	if imdbID == "" {
		// Tier 6A: DuckDuckGo scraper
		imdbID = c.fetchImdbIdFromDuckDuckGo(title, year)

		// Tier 6B: Google Search scraper fallback
		if imdbID == "" {
			imdbID = c.fetchImdbIdFromGoogle(title, year)
		}

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

// fetchImdbIdFromDuckDuckGo performs the actual HTTP scrape.
func (c *Client) fetchImdbIdFromDuckDuckGo(title string, year int) string {
	query := title + " imdb"
	if year > 0 {
		query = title + " " + strconv.Itoa(year) + " imdb"
	}

	urls := []string{
		"https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query),
		"https://lite.duckduckgo.com/lite/?q=" + url.QueryEscape(query),
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	for _, searchURL := range urls {
		req, reqErr := http.NewRequest(http.MethodGet, searchURL, nil)
		if reqErr != nil {
			continue
		}
		req.Header.Set("User-Agent", desktopUserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")

		resp, getErr := httpClient.Do(req)
		if getErr != nil {
			continue
		}

		if resp.StatusCode == 200 {
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
			resp.Body.Close()
			if readErr == nil {
				if id := imdbIdPattern.FindString(string(body)); id != "" {
					return id
				}
			}
		} else {
			resp.Body.Close()
		}
	}

	return ""
}

// fetchImdbIdFromGoogle scrapes Google Search to extract IMDb IDs if DuckDuckGo fails.
func (c *Client) fetchImdbIdFromGoogle(title string, year int) string {
	query := title + " imdb"
	if year > 0 {
		query = title + " " + strconv.Itoa(year) + " imdb"
	}
	searchURL := "https://www.google.com/search?q=" + url.QueryEscape(query)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	req, reqErr := http.NewRequest(http.MethodGet, searchURL, nil)
	if reqErr != nil {
		return ""
	}
	req.Header.Set("User-Agent", desktopUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

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
