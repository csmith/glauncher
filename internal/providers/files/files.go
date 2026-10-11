package files

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"chameth.com/glauncher/internal/assets"
	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/system"
)

// Launcher describes a set of files, and a command to open one of them with.
type Launcher struct {
	Name     string
	Aliases  []string
	Dir      string
	Patterns []string
	Exclude  []string
	Command  string
}

type Provider struct {
	launcher Launcher
	ready    chan struct{}

	mu    sync.RWMutex
	files []string
}

// NewProvider returns a provider for the given launcher. The launcher's
// directory is scanned in the background.
func NewProvider(launcher Launcher) *Provider {
	launcher.Dir = expandHome(launcher.Dir)
	p := &Provider{
		launcher: launcher,
		ready:    make(chan struct{}),
	}
	go func() {
		files := scan(launcher.Dir, launcher.Patterns, launcher.Exclude)
		p.mu.Lock()
		p.files = files
		p.mu.Unlock()
		close(p.ready)
	}()
	return p
}

func (p *Provider) Ready() <-chan struct{} {
	return p.ready
}

func (p *Provider) Search(query string) []search.Result {
	terms, ok := parseQuery(p.launcher, query)
	if !ok {
		return nil
	}

	select {
	case <-p.ready:
	default:
		return []search.Result{{
			Name:        p.launcher.Name,
			Description: p.launcher.Dir,
			Icon:        assets.Loading(48),
			Exec:        func() error { return nil },
			Priority:    1,
		}}
	}

	// Score against the last term, as it's the one being typed.
	var scoreQuery string
	if len(terms) > 0 {
		scoreQuery = terms[len(terms)-1]
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	var results []search.Result
	for _, path := range p.files {
		if !matches(path, p.launcher.Dir, terms) {
			continue
		}
		results = append(results, search.Result{
			Name:        filepath.Base(path),
			Description: path,
			Icon:        assets.File(48),
			Query:       scoreQuery,
			Exec: func() error {
				return launch(p.launcher.Command, path)
			},
		})
	}
	return results
}

// parseQuery checks whether the query is addressed to the launcher (by its
// name or an alias, as the first word) and returns the lowercased filter
// terms that follow.
func parseQuery(launcher Launcher, query string) ([]string, bool) {
	fields := strings.Fields(strings.ToLower(query))
	if len(fields) == 0 {
		return nil, false
	}
	if fields[0] != strings.ToLower(launcher.Name) && !slices.ContainsFunc(launcher.Aliases, func(a string) bool {
		return strings.ToLower(a) == fields[0]
	}) {
		return nil, false
	}
	return fields[1:], true
}

// matches reports whether every term appears in the path relative to dir.
func matches(path, dir string, terms []string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		rel = path
	}
	rel = strings.ToLower(rel)
	for _, t := range terms {
		if !strings.Contains(rel, t) {
			return false
		}
	}
	return true
}

// launch runs cmd to open path, substituting it for each %s in the
// command's arguments, or appending it if there are none.
func launch(cmd, path string) error {
	parts := commandLine(cmd, path)
	if len(parts) == 0 {
		return fmt.Errorf("no command configured")
	}
	return system.Launch(parts[0], parts[1:]...)
}

func commandLine(cmd, path string) []string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	substituted := false
	for i := range parts {
		if strings.Contains(parts[i], "%s") {
			parts[i] = strings.ReplaceAll(parts[i], "%s", path)
			substituted = true
		}
	}
	if !substituted {
		parts = append(parts, path)
	}
	return parts
}

func scan(dir string, patterns, exclude []string) []string {
	var files []string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if path != dir && (strings.HasPrefix(name, ".") || slices.Contains(exclude, name)) {
				return filepath.SkipDir
			}
			return nil
		}
		if slices.ContainsFunc(patterns, func(p string) bool {
			ok, _ := filepath.Match(p, name)
			return ok
		}) {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func expandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
