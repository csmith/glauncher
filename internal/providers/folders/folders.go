package folders

import (
	"os"
	"strings"

	"chameth.com/glauncher/internal/assets"
	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/system"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) Search(query string) []search.Result {
	if query == "" || (query[0] != '/' && query[0] != '~') {
		return nil
	}

	expanded := expandPath(query)
	dir, prefix := splitPath(expanded)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var results []search.Result

	if prefix == "" {
		name := strings.TrimRight(dir, "/")
		if name != "" {
			name = name[strings.LastIndex(name, "/")+1:]
		} else {
			name = "/"
		}
		results = append(results, search.Result{
			Name:        name,
			Description: dir,
			Icon:        assets.Folder(48),
			Query:       expanded,
			Exec: func(path string) func() error {
				return func() error {
					return launch(path)
				}
			}(strings.TrimRight(dir, "/")),
		})
	}

	prefixLower := strings.ToLower(prefix)
	var matches []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if prefixLower != "" && !strings.HasPrefix(strings.ToLower(name), prefixLower) {
			continue
		}
		matches = append(matches, name)
	}

	for _, name := range matches {
		fullPath := dir + name
		results = append(results, search.Result{
			Name:        name,
			Description: fullPath + "/",
			Icon:        assets.Folder(48),
			Query:       expanded,
			Exec: func(path string) func() error {
				return func() error {
					return launch(path)
				}
			}(fullPath),
		})
	}

	return results
}

func expandPath(query string) string {
	if strings.HasPrefix(query, "~") {
		home, _ := os.UserHomeDir()
		if len(query) == 1 {
			return home + "/"
		}
		return home + query[1:]
	}
	return query
}

func splitPath(query string) (dir string, prefix string) {
	idx := strings.LastIndex(query, "/")
	if idx < 0 {
		return "", query
	}
	return query[:idx+1], query[idx+1:]
}

func launch(path string) error {
	return system.OpenURL(path)
}
