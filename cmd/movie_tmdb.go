// movie_tmdb.go — TMDb credential helpers for interactive commands.
//
// SHARED: resolveScanTmdbCredentials, readTmdbCredentials, tmdbCredentials.
// Callers: movie scan, movie rescan, movie rescan-failed.
// Do NOT re-read TMDb config keys directly in command files — go through
// these helpers so credential resolution (DB → env → prompt) stays
// consistent and the prompt is only shown once per command lifecycle.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
)

type tmdbCredentials struct {
	ApiKey string
	Token  string
}

func (c tmdbCredentials) HasAuth() bool {
	return c.ApiKey != "" || c.Token != ""
}

// resolveScanTmdbCredentials loads saved/env credentials or prompts before scan.
func resolveScanTmdbCredentials(database *db.DB) tmdbCredentials {
	creds := readTmdbCredentials(database)
	if creds.HasAuth() {
		return creds
	}

	fmt.Println("⚠️  TMDb is not configured yet.")
	fmt.Println("   Enter your TMDb API key and/or TMDb access token before scanning.")
	fmt.Println("   Leave both blank to continue without metadata.")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("   TMDb API key: ")
	if scanner.Scan() {
		creds.ApiKey = strings.TrimSpace(scanner.Text())
	}

	fmt.Print("   TMDb access token: ")
	if scanner.Scan() {
		creds.Token = strings.TrimSpace(scanner.Text())
	}
	fmt.Println()

	if creds.ApiKey != "" {
		if err := database.SetConfig("TmdbApiKey", creds.ApiKey); err != nil {
			errlog.Warn("Could not save tmdb_api_key: %v", err)
		}
	}
	if creds.Token != "" {
		if err := database.SetConfig("TmdbToken", creds.Token); err != nil {
			errlog.Warn("Could not save tmdb_token: %v", err)
		}
	}

	if creds.HasAuth() {
		fmt.Println("✅ TMDb credentials saved.")
		fmt.Println()
		return creds
	}
	fmt.Println("⚠️  No TMDb credentials provided.")
	fmt.Println("   Scanning will continue without metadata fetching.")
	fmt.Println()

	return creds
}

// readTmdbCredentials reads TMDb credentials from config first, then env.
func readTmdbCredentials(database *db.DB) tmdbCredentials {
	creds := tmdbCredentials{
		ApiKey: strings.TrimSpace(readTmdbConfigValue(database, "TmdbApiKey")),
		Token:  strings.TrimSpace(readTmdbConfigValue(database, "TmdbToken")),
	}
	if creds.ApiKey == "" {
		creds.ApiKey = strings.TrimSpace(os.Getenv("TMDB_API_KEY"))
	}
	if creds.Token == "" {
		creds.Token = strings.TrimSpace(os.Getenv("TMDB_TOKEN"))
	}
	return creds
}

func readTmdbConfigValue(database *db.DB, key string) string {
	val, err := database.GetConfig(key)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			errlog.Warn("Config read error for %s: %v", key, err)
		}
		return ""
	}
	return val
}
