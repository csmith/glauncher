package code

import (
	"fmt"
	"image"
	"image/color"
	"os"
	execcmd "os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"chameth.com/glauncher/internal/search"

	"github.com/csmith/config"
)

type Provider struct {
	mu       sync.RWMutex
	projects []project
	dir      string
	command  string
	ready    chan struct{}
}

type project struct {
	name string
	path string
}

type configStruct struct {
	Dir     string `yaml:"dir"`
	Command string `yaml:"command"`
}

func NewProvider() (*Provider, error) {
	var cfg configStruct
	home, _ := os.UserHomeDir()
	cfg.Dir = filepath.Join(home, "code")
	cfg.Command = "code %s"

	_, _ = config.Load(&cfg, config.DirectoryName("glauncher"), config.FileName("code.yml"))

	p := &Provider{
		dir:     cfg.Dir,
		command: cfg.Command,
		ready:   make(chan struct{}),
	}
	go func() {
		p.scan()
		close(p.ready)
	}()
	return p, nil
}

func (p *Provider) Ready() <-chan struct{} {
	return p.ready
}

func (p *Provider) Search(query string) []search.Result {
	if !strings.HasPrefix(query, "code ") {
		return nil
	}

	searchStr := strings.TrimSpace(strings.TrimPrefix(query, "code"))
	if searchStr == "" {
		return nil
	}
	searchLower := strings.ToLower(searchStr)

	p.mu.RLock()
	defer p.mu.RUnlock()

	type scored struct {
		proj  project
		score int
	}

	var matches []scored
	for _, proj := range p.projects {
		nameLower := strings.ToLower(proj.name)
		var s int
		if strings.HasPrefix(nameLower, searchLower) {
			s = 100
		} else if strings.Contains(nameLower, searchLower) {
			s = 50
		}
		if s > 0 {
			matches = append(matches, scored{proj, s})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].proj.name < matches[j].proj.name
	})

	results := make([]search.Result, len(matches))
	for i := range results {
		proj := matches[i].proj
		results[i] = search.Result{
			Name:        proj.name,
			Description: proj.path,
			Icon:        folderIcon(),
			Exec: func(path string, cmd string) func() error {
				return func() error {
					return launch(cmd, path)
				}
			}(proj.path, p.command),
		}
	}

	return results
}

func (p *Provider) scan() {
	var projects []project
	filepath.WalkDir(p.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if path == p.dir {
			return nil
		}

		if isProjectDir(path) {
			projects = append(projects, project{
				name: filepath.Base(path),
				path: path,
			})
			return filepath.SkipDir
		}

		return nil
	})

	p.mu.Lock()
	p.projects = projects
	p.mu.Unlock()
}

func isProjectDir(path string) bool {
	markers := []string{".git", ".idea", ".zed", ".vscode", ".svn", ".hg"}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(path, m)); err == nil {
			return true
		}
	}
	return false
}

func launch(command string, path string) error {
	cmdStr := fmt.Sprintf(command, path)
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	c := execcmd.Command(parts[0], parts[1:]...)
	c.Stdin = nil
	c.Stdout = nil
	c.Stderr = nil
	c.SysProcAttr = &syscallSetProcessGroupID
	return c.Start()
}

func folderIcon() image.Image {
	const s = 48
	img := image.NewRGBA(image.Rect(0, 0, s, s))
	bg := color.NRGBA{R: 100, G: 140, B: 200, A: 255}
	tab := color.NRGBA{R: 80, G: 120, B: 180, A: 255}
	for y := range s {
		for x := range s {
			if y < 10 && x < 16 {
				img.Set(x, y, tab)
			} else if y >= 8 {
				img.Set(x, y, bg)
			}
		}
	}
	return img
}
