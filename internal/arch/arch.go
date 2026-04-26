package arch

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"strings"
	"sync"
	"time"

	execcmd "os/exec"

	"chameth.com/glauncher/internal/search"
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
			return []search.Result{statusResult(fmt.Sprintf("No results found for \"%s\"", searchTerm), statusIcon())}
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

	return []search.Result{statusResult(fmt.Sprintf("Searching for \"%s\"...", searchTerm), statusIcon())}
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
			Icon:        packageIcon(),
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
			Icon:        aurIcon(),
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
	c := execcmd.Command(openCommand, url)
	c.Stdin = nil
	c.Stdout = nil
	c.Stderr = nil
	c.SysProcAttr = &syscallSetProcessGroupID
	return c.Start()
}

func packageIcon() image.Image {
	const s = 48
	img := image.NewRGBA(image.Rect(0, 0, s, s))
	bg := color.NRGBA{R: 23, G: 147, B: 209, A: 255}
	darker := color.NRGBA{R: 15, G: 110, B: 170, A: 255}
	for y := range s {
		for x := range s {
			if y < 6 || y >= s-4 {
				img.Set(x, y, darker)
			} else {
				img.Set(x, y, bg)
			}
		}
	}
	mid := s / 2
	for y := 8; y < s-6; y++ {
		img.Set(mid, y, color.NRGBA{R: 255, G: 255, B: 255, A: 200})
	}
	return img
}

func aurIcon() image.Image {
	const s = 48
	img := image.NewRGBA(image.Rect(0, 0, s, s))
	bg := color.NRGBA{R: 24, G: 120, B: 180, A: 255}
	darker := color.NRGBA{R: 16, G: 90, B: 145, A: 255}
	for y := range s {
		for x := range s {
			if y < 6 || y >= s-4 {
				img.Set(x, y, darker)
			} else {
				img.Set(x, y, bg)
			}
		}
	}
	mid := s / 2
	for y := 8; y < s-6; y++ {
		img.Set(mid, y, color.NRGBA{R: 255, G: 200, B: 50, A: 200})
	}
	return img
}

func statusResult(text string, icon image.Image) search.Result {
	return search.Result{
		Name:        text,
		Description: "",
		Icon:        icon,
		Exec:        func() error { return nil },
	}
}

func statusIcon() image.Image {
	const s = 48
	img := image.NewRGBA(image.Rect(0, 0, s, s))
	bg := color.NRGBA{R: 100, G: 100, B: 120, A: 255}
	for y := range s {
		for x := range s {
			img.Set(x, y, bg)
		}
	}
	return img
}
