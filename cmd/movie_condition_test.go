package cmd

import (
	"strings"
	"testing"
)

func TestBuildConditionSQL_Simple(t *testing.T) {
	where, args, err := BuildConditionSQL("rating < 5")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(where, "TmdbRating < ?") {
		t.Fatalf("where missing rating clause: %s", where)
	}
	if !strings.Contains(where, "IsDeleted = 0") {
		t.Fatalf("missing IsDeleted guard: %s", where)
	}
	if len(args) != 1 || args[0].(float64) != 5 {
		t.Fatalf("bad args: %v", args)
	}
}

func TestBuildConditionSQL_AndOr(t *testing.T) {
	where, args, err := BuildConditionSQL("rating < 5 AND year >= 2010")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(where, "AND") {
		t.Fatalf("missing AND: %s", where)
	}
	if len(args) != 2 {
		t.Fatalf("want 2 args, got %d", len(args))
	}
}

func TestBuildConditionSQL_Aliases(t *testing.T) {
	where, _, err := BuildConditionSQL("r > 7 OR y = 2024")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(where, "TmdbRating > ?") || !strings.Contains(where, "Year = ?") {
		t.Fatalf("alias resolution failed: %s", where)
	}
}

func TestBuildConditionSQL_Genre(t *testing.T) {
	where, args, err := BuildConditionSQL(`g = "Horror"`)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(where, "EXISTS") || !strings.Contains(where, "MediaGenre") {
		t.Fatalf("genre subquery missing: %s", where)
	}
	if args[0] != "Horror" {
		t.Fatalf("bad genre arg: %v", args)
	}
}

func TestBuildConditionSQL_SizeSuffix(t *testing.T) {
	_, args, err := BuildConditionSQL("size > 2GB")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if args[0].(float64) != 2048 {
		t.Fatalf("expected 2048 MB, got %v", args[0])
	}
}

func TestBuildConditionSQL_Errors(t *testing.T) {
	cases := []string{"", "foo < 5", "rating ?? 5", "rating <"}
	for _, c := range cases {
		if _, _, err := BuildConditionSQL(c); err == nil {
			t.Fatalf("expected error for %q", c)
		}
	}
}
