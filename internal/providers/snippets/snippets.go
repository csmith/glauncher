package snippets

import (
	"strings"

	"chameth.com/glauncher/internal/assets"
	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/system"
)

type Entry struct {
	Name    string
	Aliases []string
	Content string
}

type Provider struct {
	entries []Entry
}

func NewProvider(entries []Entry) *Provider {
	return &Provider{entries: entries}
}

func (p *Provider) Search(query string) []search.Result {
	if query == "" {
		return nil
	}

	lq := strings.ToLower(query)
	var results []search.Result

	for _, s := range p.entries {
		if matches(s.Name, s.Aliases, lq) {
			results = append(
				results,
				search.Result{
					Name:        s.Name,
					Description: s.Content,
					Icon:        assets.Copy(48),
					Query:       query,
					Exec: func(content string) func() error {
						return func() error {
							return system.CopyToClipboard(content)
						}
					}(s.Content),
				},
			)
		}
	}

	return results
}

func matches(name string, aliases []string, query string) bool {
	if strings.HasPrefix(strings.ToLower(name), query) {
		return true
	}
	for _, a := range aliases {
		if strings.HasPrefix(strings.ToLower(a), query) {
			return true
		}
	}
	return false
}
