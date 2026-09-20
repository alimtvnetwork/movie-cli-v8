package tmdb

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyAuth_MissingAuth(t *testing.T) {
	c := NewClientWithToken("", "")

	err := c.VerifyAuth()
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}

	if !errors.Is(err, ErrAuthMissing) {
		t.Fatalf("expected ErrAuthMissing, got: %v", err)
	}
}

func TestVerifyAuth_ValidAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/configuration" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if r.URL.Query().Get("api_key") != "valid_key" {
			t.Fatalf("unexpected api_key: %s", r.URL.Query().Get("api_key"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"images":{"base_url":"http://image.tmdb.org/t/p/"}}`))
	}))
	defer ts.Close()

	c := NewClient("valid_key")
	c.BaseURL = ts.URL

	err := c.VerifyAuth()
	if err != nil {
		t.Fatalf("expected nil error for valid auth, got: %v", err)
	}
}

func TestVerifyAuth_InvalidAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status_code":7,"status_message":"Invalid API key: You must be granted a valid key."}`))
	}))
	defer ts.Close()

	c := NewClient("bad_key")
	c.BaseURL = ts.URL

	err := c.VerifyAuth()
	if err == nil {
		t.Fatal("expected error for invalid key, got nil")
	}

	if !errors.Is(err, ErrAuthInvalid) {
		t.Fatalf("expected ErrAuthInvalid, got: %v", err)
	}
}

func TestVerifyAuth_ForbiddenAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"status_code":3,"status_message":"Authentication failed: You do not have permissions to access the service."}`))
	}))
	defer ts.Close()

	c := NewClient("forbidden_key")
	c.BaseURL = ts.URL

	err := c.VerifyAuth()
	if err == nil {
		t.Fatal("expected error for forbidden key, got nil")
	}

	if !errors.Is(err, ErrAuthInvalid) {
		t.Fatalf("expected ErrAuthInvalid, got: %v", err)
	}
}
