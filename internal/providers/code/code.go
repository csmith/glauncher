package code

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"chameth.com/glauncher/internal/assets"
	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/system"
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

func NewProvider(dir string, command string) (*Provider, error) {
	p := &Provider{
		dir:     dir,
		command: command,
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

	var results []search.Result
	for _, proj := range p.projects {
		nameLower := strings.ToLower(proj.name)
		if !strings.HasPrefix(nameLower, searchLower) && !strings.Contains(nameLower, searchLower) {
			continue
		}
		results = append(results, search.Result{
			Name:        proj.name,
			Description: proj.path,
			Icon:        assets.Code(48),
			Query:       searchStr,
			Exec: func(path string, cmd string) func() error {
				return func() error {
					return launch(cmd, path)
				}
			}(proj.path, p.command),
		})
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

	return system.Launch(parts[0], parts[1:]...)
}
