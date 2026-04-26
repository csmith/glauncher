package search

import (
	"image"
	"testing"
)

func TestScore(t *testing.T) {
	tests := []struct {
		name     string
		nameLow  string
		descLow  string
		query    string
		expected int
	}{
		{"exact prefix match", "test", "", "test", 100},
		{"prefix match", "testing", "", "test", 100},
		{"substring match", "retest", "", "test", 50},
		{"description match", "foo", "test thing", "test", 25},
		{"no match", "other", "other thing", "test", 0},
		{"empty query", "test", "", "", 100},
		{"case insensitive", "test", "", "test", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := score(tt.nameLow, tt.descLow, tt.query)
			if got != tt.expected {
				t.Errorf("score(%q, %q, %q) = %d, want %d", tt.nameLow, tt.descLow, tt.query, got, tt.expected)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s, t     string
		expected int
	}{
		{"test", "test", 0},
		{"test", "tent", 1},
		{"test", "retest", 2},
		{"test", "othertest", 5},
		{"", "test", 4},
		{"test", "", 4},
		{"", "", 0},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.t, func(t *testing.T) {
			got := LevenshteinDistance(tt.s, tt.t)
			if got != tt.expected {
				t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d", tt.s, tt.t, got, tt.expected)
			}
		})
	}
}

func result(name, desc string) Result {
	return Result{
		Name:        name,
		Description: desc,
		Icon:        image.NewRGBA(image.Rect(0, 0, 1, 1)),
		Exec:        func() error { return nil },
	}
}

func TestSortResults_ExactPrefixComesFirst(t *testing.T) {
	results := []Result{
		result("retest", ""),
		result("othertest", ""),
		result("test", ""),
	}

	SortResults(results, "test")

	if results[0].Name != "test" {
		t.Errorf("first result = %q, want %q", results[0].Name, "test")
	}
	if results[1].Name != "retest" {
		t.Errorf("second result = %q, want %q", results[1].Name, "retest")
	}
	if results[2].Name != "othertest" {
		t.Errorf("third result = %q, want %q", results[2].Name, "othertest")
	}
}

func TestSortResults_ScoreTakesPrecedenceOverDistance(t *testing.T) {
	results := []Result{
		result("retesting", ""),
		result("atestthing", ""),
		result("test", ""),
	}

	SortResults(results, "test")

	if results[0].Name != "test" {
		t.Errorf("first result = %q, want %q", results[0].Name, "test")
	}
}

func TestSortResults_SameScoreUsesLevenshtein(t *testing.T) {
	results := []Result{
		result("othertest", ""),
		result("retest", ""),
	}

	SortResults(results, "test")

	if results[0].Name != "retest" {
		t.Errorf("first result = %q, want %q (closer levenshtein distance)", results[0].Name, "retest")
	}
}

func TestSortResults_SameScoreSameDistanceUsesName(t *testing.T) {
	results := []Result{
		result("ztest", ""),
		result("atest", ""),
	}

	SortResults(results, "test")

	if results[0].Name != "atest" {
		t.Errorf("first result = %q, want %q (alphabetical)", results[0].Name, "atest")
	}
}

func TestSortResults_DescriptionMatchScoresLower(t *testing.T) {
	results := []Result{
		result("something", "a test utility"),
		result("retest", ""),
	}

	SortResults(results, "test")

	if results[0].Name != "retest" {
		t.Errorf("first result = %q, want %q (name match scores higher than description)", results[0].Name, "retest")
	}
}

func TestSortResults_PerResultQuery(t *testing.T) {
	r := result
	results := []Result{
		{Name: "repacman", Description: "", Icon: r("repacman", "").Icon, Exec: func() error { return nil }, Query: "pacman"},
		{Name: "arch-manwarn", Description: "", Icon: r("arch-manwarn", "").Icon, Exec: func() error { return nil }, Query: "pacman"},
		{Name: "pacman", Description: "", Icon: r("pacman", "").Icon, Exec: func() error { return nil }, Query: "pacman"},
	}

	SortResults(results, "arch pacman")

	if results[0].Name != "pacman" {
		t.Errorf("first result = %q, want %q (prefix match on provider query)", results[0].Name, "pacman")
	}
}

func TestSortResults_FallbackToGlobalQuery(t *testing.T) {
	results := []Result{
		result("retest", ""),
		result("test", ""),
	}

	SortResults(results, "test")

	if results[0].Name != "test" {
		t.Errorf("first result = %q, want %q (empty Query falls back to global)", results[0].Name, "test")
	}
}
