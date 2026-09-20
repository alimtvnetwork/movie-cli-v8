package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func TestNormalizeExistingThumb(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"https://image.tmdb.org/t/p/w500/test.jpg", "https://image.tmdb.org/t/p/w500/test.jpg"},
		{"http://example.com/poster.jpg", "http://example.com/poster.jpg"},
		{"/1uRkvEDPvZUIGtGnJz895ajvb7Y.jpg", "https://image.tmdb.org/t/p/w342/1uRkvEDPvZUIGtGnJz895ajvb7Y.jpg"},
		{"/thumbnails/red-eye-2005-11460.jpg", "thumbnails/red-eye-2005-11460.jpg"},
		{"thumbnails/red-eye-2005-11460.jpg", "thumbnails/red-eye-2005-11460.jpg"},
		{`C:\Users\AppData\Local\thumbnails\red-eye.jpg`, "thumbnails/red-eye.jpg"},
		{"/home/user/.local/share/movie-cli/thumbnails/red-eye.jpg", "thumbnails/red-eye.jpg"},
		{`C:\Users\AppData/Local\thumbnails/red-eye.jpg`, "thumbnails/red-eye.jpg"},
		{"red-eye.jpg", "thumbnails/red-eye.jpg"},
	}

	for _, tc := range tests {
		got := normalizeExistingThumb(tc.input)
		if got != tc.want {
			t.Errorf("normalizeExistingThumb(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestExtractTmdbIDFromFilename(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"red-eye-2005-11460.jpg", 11460},
		{"pressure-1651941.jpg", 1651941},
		{"some-movie.jpg", 0},
		{"", 0},
	}

	for _, tc := range tests {
		got := extractTmdbIDFromFilename(tc.input)
		if got != tc.want {
			t.Errorf("extractTmdbIDFromFilename(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestIsReportPath(t *testing.T) {
	valid := []string{"/", "/report", "/report.html", "/dashboard", "/ui", "/index.html"}
	for _, p := range valid {
		if !isReportPath(p) {
			t.Errorf("isReportPath(%q) should be true", p)
		}
	}

	invalid := []string{"/api", "/other", "/unknown/path"}
	for _, p := range invalid {
		if isReportPath(p) {
			t.Errorf("isReportPath(%q) should be false", p)
		}
	}
}

func TestHandleThumbnailsServed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "thumb_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	thumbDir := filepath.Join(tmpDir, "thumbnails")
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(thumbDir, "sample-123.jpg")
	if err := os.WriteFile(testFile, []byte("fake image content"), 0644); err != nil {
		t.Fatal(err)
	}

	database := &db.DB{BasePath: tmpDir}

	req := httptest.NewRequest(http.MethodGet, "/thumbnails/sample-123.jpg", nil)
	rr := httptest.NewRecorder()

	handleThumbnails(rr, req, database)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if rr.Body.String() != "fake image content" {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}

	// Test with a slug prefix in the request path
	slugReq := httptest.NewRequest(http.MethodGet, "/dashboard/thumbnails/sample-123.jpg", nil)
	slugRr := httptest.NewRecorder()

	handleThumbnails(slugRr, slugReq, database)

	if slugRr.Code != http.StatusOK {
		t.Errorf("expected status 200 for slug path, got %d", slugRr.Code)
	}
}
