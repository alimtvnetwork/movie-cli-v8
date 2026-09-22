// Package db manages the SQLite database for the movie CLI.
package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/alimtvnetwork/movie-cli-v8/pkg/appfault"
)

const dbFile = "movie.db"

// DB wraps the sql.DB connection for primary library storage and split cache storage.
type DB struct {
	*sql.DB
	CacheDB  *sql.DB
	BasePath string // path to data directory
}

// exeDir returns the directory where the running binary is located.
func exeDir() (string, error) {
	exe, err := os.Executable()

	if err != nil {
		return "", appfault.Wrap("cannot locate executable", err)
	}

	exe, err = filepath.EvalSymlinks(exe)

	if err != nil {
		return "", appfault.Wrap("cannot resolve symlinks for executable", err)
	}

	return filepath.Dir(exe), nil
}

// Open opens (or creates) the SQLite database and runs migrations.
// The app version is stored in Config on every startup.
func Open() (*DB, error) {
	binDir, dirErr := exeDir()

	if dirErr != nil {
		return nil, dirErr
	}

	base := filepath.Join(binDir, "data")

	if err := createDataDirs(base); err != nil {
		return nil, err
	}

	conn, err := openAndConfigureDB(base)

	if err != nil {
		return nil, err
	}

	cacheConn, _ := openAndConfigureCacheDB(base)

	d := &DB{DB: conn, CacheDB: cacheConn, BasePath: base}

	if err := d.migrateSchema(); err != nil {
		_ = d.Close()

		return nil, appfault.Wrap("migration failed", err)
	}

	if cacheConn != nil {
		_ = d.initCacheSchema()
	}

	return d, nil
}

// Close closes both the master database and the cache database.
func (d *DB) Close() error {
	var firstErr error

	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			firstErr = err
		}
	}

	if d.CacheDB != nil {
		if err := d.CacheDB.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func createDataDirs(base string) error {
	dirs := []string{
		base,
		filepath.Join(base, "json", string(MediaTypeMovie)),
		filepath.Join(base, "json", string(MediaTypeTV)),
		filepath.Join(base, "json", "history"),
		filepath.Join(base, "thumbnails"),
		filepath.Join(base, "config"),
		filepath.Join(base, "log"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return appfault.Wrapf(err, "cannot create directory %s", d)
		}
	}
	return nil
}

func openAndConfigureDB(base string) (*sql.DB, error) {
	dbPath := filepath.Join(base, dbFile)
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, appfault.Wrap("cannot open database", err)
	}

	conn.SetMaxOpenConns(1)

	pragmas := []struct{ stmt, errMsg string }{
		{"PRAGMA journal_mode=WAL", "cannot set WAL mode"},
		{"PRAGMA busy_timeout = 10000", "cannot set busy_timeout"},
		{"PRAGMA foreign_keys = ON", "cannot enable foreign keys"},
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p.stmt); err != nil {
			conn.Close()
			return nil, appfault.Wrap(p.errMsg, err)
		}
	}

	return conn, nil
}
