// folder_alias_test.go — Unit tests for folder alias management and auto-derivation.
package db

import (
	"testing"
)

func TestDeriveCleanAlias(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{input: `Z:\DownloadRelated\DownloadCompletedVm\# Movies`, expected: "movies"},
		{input: `Z:\DownloadRelated\DownloadCompletedVm\# Tvshows`, expected: "tvshows"},
		{input: `/var/media/4K Movies`, expected: "4kmovies"},
		{input: `D:\Films`, expected: "films"},
		{input: `/media/Sci-Fi_Classics`, expected: "scificlassics"},
		{input: ``, expected: "folder"},
	}

	for _, tc := range cases {
		actual := DeriveCleanAlias(tc.input)

		if actual != tc.expected {
			t.Errorf("DeriveCleanAlias(%q) = %q; want %q", tc.input, actual, tc.expected)
		}
	}
}

func TestFolderAliasCRUD(t *testing.T) {
	d := openTestDB(t)

	err := d.SetFolderAlias("movies", `Z:\DownloadRelated\# Movies`)

	if err != nil {
		t.Fatalf("SetFolderAlias failed: %v", err)
	}

	path, err := d.GetFolderAlias("movies")

	if err != nil {
		t.Fatalf("GetFolderAlias failed: %v", err)
	}

	if path != `Z:\DownloadRelated\# Movies` {
		t.Errorf("unexpected path: %q", path)
	}

	aliases, err := d.ListFolderAliases()

	if err != nil {
		t.Fatalf("ListFolderAliases failed: %v", err)
	}

	if len(aliases) != 1 {
		t.Fatalf("expected 1 alias, got %d", len(aliases))
	}

	err = d.DeleteFolderAlias("movies")

	if err != nil {
		t.Fatalf("DeleteFolderAlias failed: %v", err)
	}

	deletedPath, _ := d.GetFolderAlias("movies")

	if deletedPath != "" {
		t.Errorf("expected empty path after deletion, got %q", deletedPath)
	}
}

func TestListScanFoldersWithStats(t *testing.T) {
	d := openTestDB(t)

	_, err := d.Exec("INSERT INTO ScanFolder (FolderPath) VALUES (?)", `Z:\Movies`)

	if err != nil {
		t.Fatalf("insert ScanFolder failed: %v", err)
	}

	_, err = d.InsertMedia(&Media{
		Title:            "Inception",
		CleanTitle:       "Inception",
		Year:             2010,
		Type:             "movie",
		OriginalFileName: "Inception.2010.mkv",
		OriginalFilePath: `Z:\Movies\Inception.2010.mkv`,
		CurrentFilePath:  `Z:\Movies\Inception.2010.mkv`,
		FileExtension:    ".mkv",
	})

	if err != nil {
		t.Fatalf("insert media 1 failed: %v", err)
	}

	_, err = d.InsertMedia(&Media{
		Title:            "Interstellar",
		CleanTitle:       "Interstellar",
		Year:             2014,
		Type:             "movie",
		OriginalFileName: "Interstellar.2014.mkv",
		OriginalFilePath: `Z:\Movies\Interstellar.2014.mkv`,
		CurrentFilePath:  `Z:\Movies\Interstellar.2014.mkv`,
		FileExtension:    ".mkv",
	})

	if err != nil {
		t.Fatalf("insert media 2 failed: %v", err)
	}

	stats, err := d.ListScanFoldersWithStats()

	if err != nil {
		t.Fatalf("ListScanFoldersWithStats failed: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("expected 1 scan folder stat, got %d", len(stats))
	}

	if stats[0].ItemCount != 2 {
		t.Errorf("expected 2 items, got %d", stats[0].ItemCount)
	}

	if stats[0].Alias != "movies" {
		t.Errorf("expected alias 'movies', got %q", stats[0].Alias)
	}
}
