// movie_tmdb.go — TMDb credential helpers for interactive commands.
//
// SHARED: resolveScanTmdbCredentials, readTmdbCredentials, tmdbCredentials, ensureValidTmdbClient.
// Callers: movie scan, movie rescan, movie rescan-failed, search, suggest, discover, info.
// Do NOT re-read TMDb config keys directly in command files — go through
// these helpers so credential resolution (DB → env → prompt) stays
// consistent and the prompt is only shown once per command lifecycle.
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alimtvnetwork/movie-cli-v8/db"
	"github.com/alimtvnetwork/movie-cli-v8/errlog"
	"github.com/alimtvnetwork/movie-cli-v8/tmdb"
	"github.com/mattn/go-isatty"
)

type tmdbCredentials struct {
	ApiKey string
	Token  string
}

func (c tmdbCredentials) HasAuth() bool {
	return c.ApiKey != "" || c.Token != ""
}

// resolveScanTmdbCredentials loads saved/env credentials or prompts before scan.
// It verifies credentials against TMDb and prompts again if invalid.
func resolveScanTmdbCredentials(database *db.DB) tmdbCredentials {
	creds := readTmdbCredentials(database)

	if creds.HasAuth() {
		client := tmdb.NewClientWithToken(creds.ApiKey, creds.Token)

		authErr := client.VerifyAuth()
		if authErr == nil {
			return creds
		}

		if !errors.Is(authErr, tmdb.ErrAuthInvalid) {
			errlog.Warn("Could not verify TMDb credentials online (%v); continuing with saved credentials", authErr)

			return creds
		}

		fmt.Println("❌ Configured TMDb API key or token is invalid (authentication rejected by TMDb).")
	}

	isStdinInteractive := isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
	if !isStdinInteractive {
		errlog.Error("❌ TMDb API key is invalid or unset. In non-interactive mode, continuing without metadata.")

		return tmdbCredentials{}
	}

	return promptForValidTmdbCredentials(database)
}

// promptForValidTmdbCredentials prompts the user for TMDb credentials and
// loops until valid credentials are provided or the user skips.
func promptForValidTmdbCredentials(database *db.DB) tmdbCredentials {
	fmt.Println("⚠️  TMDb is not configured or current credentials are invalid.")
	fmt.Println("   Enter a valid TMDb API key or Bearer token.")
	fmt.Println("   (Leave blank and press Enter to continue without metadata):")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("   TMDb API key or token: ")

		if !scanner.Scan() {
			break
		}

		rawInput := strings.Trim(strings.TrimSpace(scanner.Text()), "\"'`")

		if rawInput == "" {
			fmt.Println("⚠️  No TMDb credentials provided. Scanning will continue without metadata.")
			fmt.Println()

			return tmdbCredentials{}
		}

		var inputKey string
		var inputToken string

		if strings.HasPrefix(rawInput, "eyJ") {
			inputToken = rawInput
		}

		if !strings.HasPrefix(rawInput, "eyJ") {
			inputKey = rawInput
		}

		testClient := tmdb.NewClientWithToken(inputKey, inputToken)

		testErr := testClient.VerifyAuth()
		if testErr == nil {
			saveTmdbCredentialsToDB(database, inputKey, inputToken)
			fmt.Println("✅ TMDb credentials verified and saved.")
			fmt.Println()

			return tmdbCredentials{ApiKey: inputKey, Token: inputToken}
		}

		if errors.Is(testErr, tmdb.ErrAuthInvalid) {
			fmt.Println("❌ That TMDb API key or token was rejected as invalid. Please try again:")
			fmt.Println()

			continue
		}

		fmt.Printf("⚠️  Could not reach TMDb to verify (%v). Saving credentials anyway.\n\n", testErr)
		saveTmdbCredentialsToDB(database, inputKey, inputToken)

		return tmdbCredentials{ApiKey: inputKey, Token: inputToken}
	}

	return tmdbCredentials{}
}

func saveTmdbCredentialsToDB(database *db.DB, apiKey, token string) {
	if apiKey != "" {
		if err := database.SetConfig("TmdbApiKey", apiKey); err != nil {
			errlog.Warn("Could not save TmdbApiKey: %v", err)
		}

		if err := database.SetConfig("tmdb_api_key", apiKey); err != nil {
			errlog.Warn("Could not save tmdb_api_key: %v", err)
		}
	}

	if token != "" {
		if err := database.SetConfig("TmdbToken", token); err != nil {
			errlog.Warn("Could not save TmdbToken: %v", err)
		}

		if err := database.SetConfig("tmdb_token", token); err != nil {
			errlog.Warn("Could not save tmdb_token: %v", err)
		}
	}
}

// ensureValidTmdbClient resolves and verifies TMDb credentials from DB/env,
// prompting for a valid key if missing or rejected.
func ensureValidTmdbClient(database *db.DB) *tmdb.Client {
	creds := resolveScanTmdbCredentials(database)

	if !creds.HasAuth() {
		return nil
	}

	client := tmdb.NewClientWithToken(creds.ApiKey, creds.Token)
	client.SetImdbCache(newImdbCacheAdapter(database))

	return client
}

// readTmdbCredentials reads TMDb credentials from config first, then env.
func readTmdbCredentials(database *db.DB) tmdbCredentials {
	apiKey := strings.TrimSpace(readTmdbConfigValue(database, "TmdbApiKey"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(readTmdbConfigValue(database, "tmdb_api_key"))
	}

	token := strings.TrimSpace(readTmdbConfigValue(database, "TmdbToken"))
	if token == "" {
		token = strings.TrimSpace(readTmdbConfigValue(database, "tmdb_token"))
	}

	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("TMDB_API_KEY"))
	}

	if token == "" {
		token = strings.TrimSpace(os.Getenv("TMDB_TOKEN"))
	}

	return tmdbCredentials{
		ApiKey: apiKey,
		Token:  token,
	}
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
