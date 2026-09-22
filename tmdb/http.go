package tmdb

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

func (c *Client) buildURL(path string, params url.Values) string {
	if params == nil {
		params = url.Values{}
	}

	if c.AccessToken == "" && c.ApiKey != "" {
		params.Set("api_key", c.ApiKey)
	}

	base := baseURL
	if c.BaseURL != "" {
		base = c.BaseURL
	}

	encoded := params.Encode()
	if encoded == "" {
		return base + path
	}

	return base + path + "?" + encoded
}

// MaxRetries is the number of retry attempts for rate-limited requests.
const MaxRetries = 3

func (c *Client) get(reqURL string, target interface{}) error {
	var lastErr error
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		lastErr = c.doGet(reqURL, target, attempt)
		if lastErr == nil {
			return nil
		}

		if errors.Is(lastErr, ErrRateLimited) || errors.Is(lastErr, ErrAuthInvalid) {
			if c.RotateCredential() {
				updatedURL := c.rebuildURLWithCurrentKey(reqURL)
				return c.get(updatedURL, target)
			}
			if errors.Is(lastErr, ErrRateLimited) {
				continue
			}
			return lastErr
		}

		isFatal := errors.Is(lastErr, ErrTimeout) ||
			errors.Is(lastErr, ErrNetworkError)
		if isFatal {
			return lastErr
		}
	}

	return appfault.Wrapf(lastErr, "TMDb request failed after %d retries", MaxRetries)
}

func (c *Client) rebuildURLWithCurrentKey(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := parsed.Query()
	if q.Has("api_key") {
		c.credMu.Lock()
		key := c.ApiKey
		c.credMu.Unlock()

		if key != "" {
			q.Set("api_key", key)
		} else {
			q.Del("api_key")
		}
		parsed.RawQuery = q.Encode()
	}

	return parsed.String()
}

func (c *Client) doGet(reqURL string, target interface{}, attempt int) error {
	DefaultLimiter().Wait()
	req, reqErr := http.NewRequest(http.MethodGet, reqURL, nil)
	if reqErr != nil {
		backoff(attempt)
		return appfault.Wrap("build request failed", reqErr)
	}

	c.credMu.Lock()
	token := c.AccessToken
	c.credMu.Unlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return classifyHTTPError(err)
	}

	return handleResponse(resp, target, attempt)
}

func classifyHTTPError(err error) error {
	if IsTimeoutError(err) {
		return appfault.Wrap("check your internet connection", ErrTimeout)
	}

	if IsNetworkError(err) {
		return ErrNetworkError
	}

	return appfault.Wrap("HTTP request failed", err)
}

func handleResponse(resp *http.Response, target interface{}, attempt int) error {
	switch {
	case resp.StatusCode == 401 || resp.StatusCode == 403:
		resp.Body.Close()

		return appfault.Wrap("Run: movie config set tmdb_api_key YOUR_KEY", ErrAuthInvalid)

	case resp.StatusCode == 429:
		resp.Body.Close()
		retryAfter := resp.Header.Get("Retry-After")
		delay := 2 * time.Second

		if secs, parseErr := time.ParseDuration(retryAfter + "s"); parseErr == nil && secs > 0 {
			delay = secs
		}

		time.Sleep(delay)

		return ErrRateLimited

	case resp.StatusCode >= 500:
		resp.Body.Close()
		lastErr := appfault.Wrapf(ErrServerError, "HTTP %d", resp.StatusCode)

		if attempt == 0 {
			delay := serverRetryDelay(resp.StatusCode)
			time.Sleep(delay)
		}

		return lastErr

	case resp.StatusCode != 200:
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return appfault.New("TMDb API error %d: %s", resp.StatusCode, string(body))
	}

	err := json.NewDecoder(resp.Body).Decode(target)
	resp.Body.Close()
	return err
}

// backoff sleeps for exponential duration: 1s, 2s, 4s, ...
func backoff(attempt int) {
	if attempt >= MaxRetries {
		return
	}
	d := time.Duration(1<<uint(attempt)) * time.Second
	time.Sleep(d)
}

// serverRetryDelay returns the retry delay for server errors based on status code.
func serverRetryDelay(statusCode int) time.Duration {
	if statusCode == 502 || statusCode == 503 || statusCode == 504 {
		return 5 * time.Second
	}
	return 3 * time.Second
}
