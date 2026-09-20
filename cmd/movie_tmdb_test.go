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
