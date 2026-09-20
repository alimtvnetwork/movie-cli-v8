package cmd

import (
	"testing"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func TestReadTmdbCredentials_FromConfig(t *testing.T) {
	d, err := db.OpenAtPathForTest(t.TempDir())
	if err != nil {
		t.Fatalf("OpenAtPathForTest failed: %v", err)
	}
	defer d.Close()

	if setErr := d.SetConfig("TmdbApiKey", "test_key_123"); setErr != nil {
		t.Fatalf("SetConfig TmdbApiKey failed: %v", setErr)
	}

	if setErr := d.SetConfig("TmdbToken", "test_token_456"); setErr != nil {
		t.Fatalf("SetConfig TmdbToken failed: %v", setErr)
	}

	creds := readTmdbCredentials(d)

	if creds.ApiKey != "test_key_123" {
		t.Fatalf("expected ApiKey test_key_123, got %s", creds.ApiKey)
	}

	if creds.Token != "test_token_456" {
		t.Fatalf("expected Token test_token_456, got %s", creds.Token)
	}

	if !creds.HasAuth() {
		t.Fatal("expected HasAuth() to be true")
	}
}

func TestEnsureValidTmdbClient_NoAuth(t *testing.T) {
	d, err := db.OpenAtPathForTest(t.TempDir())
	if err != nil {
		t.Fatalf("OpenAtPathForTest failed: %v", err)
	}
	defer d.Close()

	client := ensureValidTmdbClient(d)
	if client != nil {
		t.Fatalf("expected nil client when auth is missing/non-interactive, got: %v", client)
	}
}

func TestReadTmdbCredentials_FromSnakeCaseConfig(t *testing.T) {
	d, err := db.OpenAtPathForTest(t.TempDir())
	if err != nil {
		t.Fatalf("OpenAtPathForTest failed: %v", err)
	}
	defer d.Close()

	if setErr := d.SetConfig("tmdb_api_key", "snake_key_999"); setErr != nil {
		t.Fatalf("SetConfig tmdb_api_key failed: %v", setErr)
	}

	if setErr := d.SetConfig("tmdb_token", "snake_token_888"); setErr != nil {
		t.Fatalf("SetConfig tmdb_token failed: %v", setErr)
	}

	creds := readTmdbCredentials(d)

	if creds.ApiKey != "snake_key_999" {
		t.Fatalf("expected ApiKey snake_key_999, got %s", creds.ApiKey)
	}

	if creds.Token != "snake_token_888" {
		t.Fatalf("expected Token snake_token_888, got %s", creds.Token)
	}
}

func TestSaveTmdbCredentialsToDB_SavesBothCases(t *testing.T) {
	d, err := db.OpenAtPathForTest(t.TempDir())
	if err != nil {
		t.Fatalf("OpenAtPathForTest failed: %v", err)
	}
	defer d.Close()

	saveTmdbCredentialsToDB(d, "saved_key", "saved_token")

	val1, _ := d.GetConfig("TmdbApiKey")
	val2, _ := d.GetConfig("tmdb_api_key")
	if val1 != "saved_key" || val2 != "saved_key" {
		t.Fatalf("expected saved_key in both cases, got %s / %s", val1, val2)
	}
}
