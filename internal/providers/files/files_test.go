package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQuery(t *testing.T) {
	launcher := Launcher{Name: "Notes", Aliases: []string{"n"}}

	tests := []struct {
		query string
		terms []string
		ok    bool
	}{
		{"notes", []string{}, true},
		{"NOTES ", []string{}, true},
		{"n 029", []string{"029"}, true},
		{"n Proj  029", []string{"proj", "029"}, true},
		{"no", nil, false},
		{"nx 029", nil, false},
		{"", nil, false},
	}
	for _, tt := range tests {
		t.Run(
			tt.query,
			func(t *testing.T) {
				terms, ok := parseQuery(launcher, tt.query)
				assert.Equal(t, tt.ok, ok)
				if tt.ok {
					assert.Equal(t, tt.terms, terms)
				}
			},
		)
	}
}

func TestMatchesUsesPathRelativeToDir(t *testing.T) {
	path := "/home/user/code/Proj/specs/029-thing.md"
	assert.True(t, matches(path, "/home/user/code", []string{"proj", "029"}))
	assert.True(t, matches(path, "/home/user/code", nil))
	assert.False(t, matches(path, "/home/user/code", []string{"029", "other"}))
	assert.False(t, matches(path, "/home/user/code", []string{"user"}))
}

func TestCommandLine(t *testing.T) {
	assert.Equal(t, []string{"ed", "--open=/a b.md", "-x"}, commandLine("ed --open=%s -x", "/a b.md"))
	assert.Equal(t, []string{"ed", "-x", "/a b.md"}, commandLine("ed -x", "/a b.md"))
	assert.Nil(t, commandLine("  ", "/a.md"))
}

func TestScan(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{
		"a.md",
		"sub/b.md",
		"sub/c.txt",
		".git/d.md",
		"node_modules/e.md",
		"sub/node_modules/f.md",
	} {
		path := filepath.Join(dir, f)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, nil, 0o644))
	}

	files := scan(dir, []string{"*.md"}, []string{"node_modules"})
	assert.ElementsMatch(
		t,
		[]string{
			filepath.Join(dir, "a.md"),
			filepath.Join(dir, "sub/b.md"),
		},
		files,
	)
}

func TestSearchFiltersScannedFiles(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"proj/specs/029-spec.md", "proj/specs/030-spec.md", "other/029.md"} {
		path := filepath.Join(dir, f)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, nil, 0o644))
	}

	p := NewProvider(Launcher{Name: "rl", Dir: dir, Patterns: []string{"*.md"}, Command: "true"})
	<-p.Ready()

	assert.Len(t, p.Search("rl"), 3)
	assert.Nil(t, p.Search("rlx"))

	results := p.Search("rl proj 029")
	require.Len(t, results, 1)
	assert.Equal(t, "029-spec.md", results[0].Name)
	assert.Equal(t, filepath.Join(dir, "proj/specs/029-spec.md"), results[0].Description)
}
