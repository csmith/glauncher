package searchweb

import (
	"fmt"
	"net/url"
	"strings"

	"chameth.com/glauncher/internal/assets"
	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/system"
)

type Entry struct {
	Name          string
	Aliases       []string
	URL           string
	AlwaysInclude bool
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

	var results []search.Result

	for _, sp := range p.entries {
		if matchPrefix(sp.Name, sp.Aliases, query) {
			searchTerm := trimPrefix(sp.Name, sp.Aliases, query)
			if searchTerm == "" {
				continue
			}
			results = append(results, makeResult(sp, searchTerm, 0))
			return results
		}
	}

	for _, sp := range p.entries {
		if sp.AlwaysInclude {
			results = append(results, makeResult(sp, query, -100))
		}
	}

	return results
}

func matchPrefix(name string, aliases []string, query string) bool {
	lq := strings.ToLower(query)
	if strings.HasPrefix(lq, strings.ToLower(name)+" ") {
		return true
	}
	for _, a := range aliases {
		if strings.HasPrefix(lq, strings.ToLower(a)+" ") {
			return true
		}
	}
	return false
}

func trimPrefix(name string, aliases []string, query string) string {
	lq := strings.ToLower(query)
	for _, prefix := range append(aliases, name) {
		p := strings.ToLower(prefix) + " "
		if strings.HasPrefix(lq, p) {
			return query[len(p):]
		}
	}
	return query
}

func makeResult(sp Entry, searchTerm string, priority int) search.Result {
	return search.Result{
		Name:        fmt.Sprintf("Search on %s", sp.Name),
		Description: searchTerm,
		Icon:        assets.Search(48),
		Exec: func(u string) func() error {
			return func() error {
				return openURL(u)
			}
		}(fmt.Sprintf(sp.URL, url.QueryEscape(searchTerm))),
		Priority: priority,
	}
}

func openURL(u string) error {
	return system.OpenURL(u)
}
