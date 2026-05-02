// open_test_helper.go — public, test-only helper exposed for cross-package
// integration tests that live in OTHER packages (e.g. cmd/) and need a
// real schema-migrated SQLite database without touching the on-disk
// movie.db file used by the running binary.
//
// Why this is in a non-_test.go file:
//
//	Go build tags do not let one package's _test.go expose helpers to a
//	different package. The cmd/ integration tests need a database, so this
//	helper has to be reachable from a regular .go file. The function name
//	is prefixed with "Test" + suffixed with "ForTest" to make grep-ability
//	obvious — production code MUST NOT call OpenInMemoryForTest.
//
// Lint-style guarantees:
//
//   - Schema parity with Open(): runs the same migrateSchema() so all
//     tables, views, indexes, and seed rows are present.
//   - BasePath set to t.TempDir() (passed in) so any code that writes
//     thumbnails/JSON history under d.BasePath stays sandboxed.
//   - Cleanup is the caller's responsibility (defer d.Close()).
package db

import (
	"database/sql"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

// OpenInMemoryForTest opens a fresh in-memory SQLite database, runs the
// full schema migration, and seeds the lookup tables. basePath should be
// a temp directory (typically t.TempDir()) — it's used as DB.BasePath for
// any code that writes auxiliary files under <BasePath>/json or thumbnails.
//
// PRODUCTION CODE MUST NEVER CALL THIS. The "ForTest" suffix and the
// package-doc warning are the only enforcement; relying on naming because
// Go has no test-only export visibility.
func OpenInMemoryForTest(basePath string) (*DB, error) {
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, apperror.Wrap("open in-memory db", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, apperror.Wrap("enable foreign keys", err)
	}
	d := &DB{DB: conn, BasePath: basePath}
	if err := d.migrateSchema(); err != nil {
		conn.Close()
		return nil, apperror.Wrap("migrate schema", err)
	}
	return d, nil
}

// OpenAtPathForTest opens (or creates) the on-disk movie.db located under
// basePath/data/movie.db and runs the standard migration. This mirrors the
// real Open() path layout so integration tests can share one database with
// a separately-built CLI binary placed in basePath.
//
// PRODUCTION CODE MUST NEVER CALL THIS. Used by cmd/ end-to-end tests
// only.
func OpenAtPathForTest(basePath string) (*DB, error) {
	dataDir := filepath.Join(basePath, "data")
	if err := createDataDirs(dataDir); err != nil {
		return nil, err
	}
	conn, err := openAndConfigureDB(dataDir)
	if err != nil {
		return nil, err
	}
	d := &DB{DB: conn, BasePath: dataDir}
	if err := d.migrateSchema(); err != nil {
		conn.Close()
		return nil, apperror.Wrap("migration failed", err)
	}
	return d, nil
}
