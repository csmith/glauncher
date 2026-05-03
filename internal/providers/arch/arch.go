package arch

import (
	"encoding/json"
	"fmt"
	"image"
	"net/http"
	"strings"
	"sync"
	"time"

	"chameth.com/glauncher/internal/assets"
	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/system"
)

const debounceDelay = time.Second

type Provider struct {
	mu            sync.RWMutex
	results       []search.Result
	lastSearch    string
	pendingSearch string
	searching     bool
	timer         *time.Timer
	invalidate    func()
	client        *http.Client
}

func NewProvider() *Provider {
	return &Provider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *Provider) SetInvalidate(f func()) {
	p.mu.Lock()
	p.invalidate = f
	p.mu.Unlock()
}

func (p *Provider) Search(query string) []search.Result {
	if !strings.HasPrefix(query, "arch ") {
		return nil
	}

	searchTerm := strings.TrimSpace(strings.TrimPrefix(query, "arch"))
	if searchTerm == "" {
		return nil
	}

	p.mu.Lock()
	if searchTerm == p.lastSearch && !p.searching {
		results := p.results
		p.mu.Unlock()
		if len(results) == 0 {
			return []search.Result{statusResult(fmt.Sprintf("No results found for \"%s\"", searchTerm), assets.Error(48))}
		}
		return results
	}

	p.pendingSearch = searchTerm
	p.searching = true
	if p.timer != nil {
		p.timer.Stop()
	}
	p.timer = time.AfterFunc(debounceDelay, p.doSearch)
	p.mu.Unlock()

	return []search.Result{statusResult(fmt.Sprintf("Searching for \"%s\"...", searchTerm), assets.Loading(48))}
}

func (p *Provider) doSearch() {
	p.mu.RLock()
	searchTerm := p.pendingSearch
	invalidate := p.invalidate
	p.mu.RUnlock()

	if searchTerm == "" || invalidate == nil {
		return
	}

	var official []officialResult
	var aur []aurResult
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		official = p.searchOfficial(searchTerm)
	}()
	go func() {
		defer wg.Done()
		aur = p.searchAUR(searchTerm)
	}()
	wg.Wait()

	var results []search.Result
	for _, pkg := range official {
		url := fmt.Sprintf("https://archlinux.org/packages/%s/%s/%s/", pkg.Repo, pkg.Arch, pkg.Pkgname)
		results = append(results, search.Result{
			Name:        pkg.Pkgname,
			Description: fmt.Sprintf("[%s/%s] %s", pkg.Repo, pkg.Arch, pkg.Pkgdesc),
			Icon:        assets.Arch(48),
			Query:       searchTerm,
			Exec: func(u string) func() error {
				return func() error {
					return openURL(u)
				}
			}(url),
		})
	}
	for _, pkg := range aur {
		url := fmt.Sprintf("https://aur.archlinux.org/packages/%s", pkg.Name)
		results = append(results, search.Result{
			Name:        pkg.Name,
			Description: fmt.Sprintf("[AUR] %s", pkg.Description),
			Icon:        assets.ArchAlt(48),
			Query:       searchTerm,
			Exec: func(u string) func() error {
				return func() error {
					return openURL(u)
				}
			}(url),
		})
	}

	p.mu.Lock()
	p.results = results
	p.lastSearch = searchTerm
	p.searching = false
	p.mu.Unlock()

	invalidate()
}

func (p *Provider) searchOfficial(query string) []officialResult {
	url := fmt.Sprintf("https://archlinux.org/packages/search/json/?q=%s", query)
	resp, err := p.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var data struct {
		Results []officialResult `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}
	return data.Results
}

func (p *Provider) searchAUR(query string) []aurResult {
	url := fmt.Sprintf("https://aur.archlinux.org/rpc/v5/search/%s", query)
	resp, err := p.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var data struct {
		Results []aurResult `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}
	return data.Results
}

type officialResult struct {
	Pkgname string `json:"pkgname"`
	Pkgdesc string `json:"pkgdesc"`
	Repo    string `json:"repo"`
	Arch    string `json:"arch"`
}

type aurResult struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

func openURL(url string) error {
	return system.OpenURL(url)
}

func statusResult(text string, icon image.Image) search.Result {
	return search.Result{
		Name:        text,
		Description: "",
		Icon:        icon,
		Exec:        func() error { return nil },
	}
}
