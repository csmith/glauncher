package folders

import (
	"image"
	"image/color"
	"os"
	execcmd "os/exec"
	"strings"

	"chameth.com/glauncher/internal/search"
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
		cleanDir := strings.TrimRight(dir, "/")
		if cleanDir == "" && dir == "/" {
			cleanDir = "/"
		}
		if cleanDir != "" {
			name := cleanDir[strings.LastIndex(cleanDir, "/")+1:]
			if name == "" {
				name = "/"
			}
			results = append(results, search.Result{
				Name:        name,
				Description: dir,
				Icon:        folderIcon(),
				Query:       expanded,
				Exec: func(path string) func() error {
					return func() error {
						return launch(path)
					}
				}(cleanDir),
			})
		}
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
			Icon:        folderIcon(),
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
	c := execcmd.Command(openCommand, path)
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
