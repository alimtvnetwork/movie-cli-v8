// movie_cd_test.go — Unit tests for movie cd target resolution and output formatting.
package cmd

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/alimtvnetwork/movie-cli-v8/db"
)

func setupTestDBForCd(t *testing.T) *db.DB {
	t.Helper()
	conn, err := sql.Open("sqlite", ":memory:")

	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	d := &db.DB{DB: conn, BasePath: t.TempDir()}

	// Migrate schema
	schemaSql := `
		CREATE TABLE IF NOT EXISTS ScanFolder (
			ScanFolderId INTEGER PRIMARY KEY AUTOINCREMENT,
			FolderPath TEXT UNIQUE NOT NULL,
			CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS FolderAlias (
			AliasId INTEGER PRIMARY KEY AUTOINCREMENT,
			AliasName TEXT NOT NULL UNIQUE,
			FolderPath TEXT NOT NULL,
			CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS Media (
			MediaId INTEGER PRIMARY KEY AUTOINCREMENT,
			Title TEXT NOT NULL,
			CleanTitle TEXT,
			Year INTEGER,
			Type TEXT DEFAULT 'movie',
			OriginalFileName TEXT,
			OriginalFilePath TEXT,
			CurrentFilePath TEXT,
			IsDeleted INTEGER DEFAULT 0,
			TmdbRating REAL DEFAULT 0.0
		);
	`
	_, execErr := conn.Exec(schemaSql)

	if execErr != nil {
		t.Fatalf("init test schema: %v", execErr)
	}

	t.Cleanup(func() { conn.Close() })

	return d
}

func TestResolveCdTargetByAlias(t *testing.T) {
	d := setupTestDBForCd(t)

	_, err := d.Exec("INSERT INTO FolderAlias (AliasName, FolderPath) VALUES (?, ?)", "downloads", `D:\Downloads`)

	if err != nil {
		t.Fatalf("insert alias failed: %v", err)
	}

	res, _, err := resolveCdTarget(d, "downloads", "")

	if err != nil {
		t.Fatalf("resolveCdTarget failed: %v", err)
	}

	if res == nil {
		t.Fatal("expected non-nil CdTargetResult")
	}

	if res.TargetDirectory != `D:\Downloads` {
		t.Errorf("expected target D:\\Downloads, got %q", res.TargetDirectory)
	}

	if res.MatchType != "alias" {
		t.Errorf("expected MatchType 'alias', got %q", res.MatchType)
	}
}

func TestResolveCdTargetByNumber(t *testing.T) {
	d := setupTestDBForCd(t)

	_, err := d.Exec("INSERT INTO ScanFolder (FolderPath) VALUES (?)", `D:\Media\Movies`)

	if err != nil {
		t.Fatalf("insert scan folder: %v", err)
	}

	res, _, err := resolveCdTarget(d, "1", "")

	if err != nil {
		t.Fatalf("resolveCdTarget by number: %v", err)
	}

	if res == nil {
		t.Fatal("expected non-nil result for index 1")
	}

	if res.TargetDirectory != `D:\Media\Movies` {
		t.Errorf("expected D:\\Media\\Movies, got %q", res.TargetDirectory)
	}
}

func TestResolveCdTargetByMovieTitle(t *testing.T) {
	d := setupTestDBForCd(t)

	_, err := d.Exec(`
		INSERT INTO Media (Title, CleanTitle, Year, CurrentFilePath, OriginalFilePath, IsDeleted)
		VALUES (?, ?, ?, ?, ?, 0)`,
		"The Matrix", "The Matrix", 1999, `D:\Films\The Matrix (1999)\matrix.mkv`, `D:\Films\The Matrix (1999)\matrix.mkv`)

	if err != nil {
		t.Fatalf("insert media: %v", err)
	}

	res, _, err := resolveCdTarget(d, "Matrix", "")

	if err != nil {
		t.Fatalf("resolveCdTarget by movie title: %v", err)
	}

	if res == nil {
		t.Fatal("expected non-nil result for movie Matrix")
	}

	if res.MovieTitle != "The Matrix" {
		t.Errorf("expected movie title 'The Matrix', got %q", res.MovieTitle)
	}

	expectedDir := `D:\Films\The Matrix (1999)`

	if res.TargetDirectory != expectedDir {
		t.Errorf("expected dir %q, got %q", expectedDir, res.TargetDirectory)
	}
}

func TestTruncateText(t *testing.T) {
	short := truncateText("hello", 10)

	if short != "hello" {
		t.Errorf("expected 'hello', got %q", short)
	}

	long := truncateText("supercalifragilistic", 8)

	if long != "super..." {
		t.Errorf("expected 'super...', got %q", long)
	}
}
