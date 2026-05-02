// checks_env.go — environmental checks added to `movie doctor`:
// required config keys, source folder existence, and REST port availability.
//
// These extend the original updater-focused checks with day-to-day
// "is the app actually usable" diagnostics so users no longer need to
// copy-paste a bash verifier block.
package doctor

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

const (
	idConfigKeys   = "config-keys"
	idSourceFolder = "source-folder"
	idRestPort     = "rest-port"

	doctorRestPort       = 7777
	requiredKeyTmdb      = "TmdbApiKey"
	requiredKeyScanDir   = "ScanDir"
	portDialTimeout      = 300 * time.Millisecond
	defaultRestPortLabel = "default REST port 7777"
)

// runEnvChecks appends config/source/port findings to the report.
func runEnvChecks(report *Report) {
	report.Findings = append(report.Findings, checkConfigKeys())
	report.Findings = append(report.Findings, checkSourceFolder())
	report.Findings = append(report.Findings, checkRestPort())
}

func checkConfigKeys() Finding {
	val, err := readConfigKey(requiredKeyTmdb)
	if err != nil || val == "" {
		return finding(idConfigKeys, "Required config key TmdbApiKey",
			SeverityErr, "TmdbApiKey is not set",
			"Run `movie config set tmdb_api_key <YOUR_KEY>`", false)
	}
	return finding(idConfigKeys, "Required config keys present",
		SeverityOK, "TmdbApiKey set", "", false)
}

func checkSourceFolder() Finding {
	dir, err := readConfigKey(requiredKeyScanDir)
	if err != nil || dir == "" {
		return finding(idSourceFolder, "Source folder (ScanDir)",
			SeverityWarn, "ScanDir is not configured",
			"Run `movie config set scan_dir <PATH>`", false)
	}
	info, statErr := os.Stat(dir)
	if statErr != nil {
		return finding(idSourceFolder, "Source folder (ScanDir)",
			SeverityErr, fmt.Sprintf("%s: %v", dir, statErr),
			"Create the folder or update scan_dir", false)
	}
	if !info.IsDir() {
		return finding(idSourceFolder, "Source folder (ScanDir)",
			SeverityErr, fmt.Sprintf("%s is not a directory", dir),
			"Point scan_dir at a directory", false)
	}
	return finding(idSourceFolder, "Source folder exists",
		SeverityOK, dir, "", false)
}

func checkRestPort() Finding {
	addr := fmt.Sprintf("127.0.0.1:%d", doctorRestPort)
	conn, err := net.DialTimeout("tcp", addr, portDialTimeout)
	if err != nil {
		return finding(idRestPort, defaultRestPortLabel,
			SeverityOK, fmt.Sprintf("port %d is free", doctorRestPort), "", false)
	}
	_ = conn.Close()
	return finding(idRestPort, defaultRestPortLabel,
		SeverityWarn, fmt.Sprintf("port %d is already in use", doctorRestPort),
		"Stop the other process or pass --port to `movie rest`", false)
}

func readConfigKey(key string) (string, error) {
	database, err := db.Open()
	if err != nil {
		return "", err
	}
	defer database.Close()
	return database.GetConfig(key)
}
