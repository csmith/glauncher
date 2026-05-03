package desktop

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/system"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

type Provider struct {
	mu        sync.RWMutex
	entries   []entry
	iconCache map[string]image.Image
	ready     chan struct{}
}

type entry struct {
	name          string
	comment       string
	iconName      string
	exec          string
	terminal      bool
	nameLower     string
	commentLower  string
	keywordsLower string
}

func NewProvider() *Provider {
	p := &Provider{
		iconCache: make(map[string]image.Image),
		ready:     make(chan struct{}),
	}
	go func() {
		p.load()
		close(p.ready)
	}()
	return p
}

func (p *Provider) Ready() <-chan struct{} {
	return p.ready
}

func (p *Provider) Search(query string) []search.Result {
	if query == "" {
		return nil
	}

	q := strings.ToLower(query)

	p.mu.RLock()
	defer p.mu.RUnlock()

	var results []search.Result
	for _, e := range p.entries {
		if !strings.HasPrefix(e.nameLower, q) && !strings.Contains(e.nameLower, q) && !strings.Contains(e.commentLower, q) && !strings.Contains(e.keywordsLower, q) {
			continue
		}
		results = append(results, search.Result{
			Name:        e.name,
			Description: e.comment,
			Icon:        p.lookupIcon(e.iconName),
			Exec: func(exec string) func() error {
				return func() error {
					return launch(exec)
				}
			}(e.exec),
		})
	}

	return results
}

func (p *Provider) load() {
	dirs := xdgDataDirs()
	seen := make(map[string]bool)
	var entries []entry

	for _, dir := range dirs {
		appDir := filepath.Join(dir, "applications")
		scanDirectory(appDir, seen, &entries)
	}

	p.mu.Lock()
	p.entries = entries
	p.mu.Unlock()
}

func scanDirectory(dir string, seen map[string]bool, entries *[]entry) {
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".desktop") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		if seen[rel] {
			return nil
		}
		seen[rel] = true

		e := parseDesktopFile(path)
		if e != nil {
			*entries = append(*entries, *e)
		}
		return nil
	})
}

func xdgDataDirs() []string {
	var dirs []string

	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "share"))
	}

	if xdg := os.Getenv("XDG_DATA_DIRS"); xdg != "" {
		for d := range strings.SplitSeq(xdg, ":") {
			d = strings.TrimSpace(d)
			if d != "" {
				dirs = append(dirs, d)
			}
		}
	} else {
		dirs = append(dirs, "/usr/local/share", "/usr/share")
	}

	return dirs
}

func parseDesktopFile(path string) *entry {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	e := &entry{}
	inDesktopEntry := false
	isApplication := false
	var keywords string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if line == "[Desktop Entry]" {
			inDesktopEntry = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDesktopEntry = false
			continue
		}
		if !inDesktopEntry {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "Type":
			isApplication = value == "Application"
		case "Name":
			if e.name == "" {
				e.name = value
			}
		case "Comment":
			if e.comment == "" {
				e.comment = value
			}
		case "Icon":
			if e.iconName == "" {
				e.iconName = value
			}
		case "Exec":
			if e.exec == "" {
				e.exec = value
			}
		case "Terminal":
			e.terminal = value == "true"
		case "NoDisplay":
			if value == "true" {
				return nil
			}
		case "Hidden":
			if value == "true" {
				return nil
			}
		case "Keywords":
			if keywords == "" {
				keywords = value
			}
		}
	}

	if !isApplication || e.name == "" || e.exec == "" {
		return nil
	}

	e.nameLower = strings.ToLower(e.name)
	e.commentLower = strings.ToLower(e.comment)
	e.keywordsLower = strings.ToLower(strings.ReplaceAll(keywords, ";", " "))
	return e
}

func (p *Provider) lookupIcon(name string) image.Image {
	if name == "" {
		return placeholderIcon()
	}

	if img, ok := p.iconCache[name]; ok {
		return img
	}

	if img := p.findIconFile(name); img != nil {
		p.iconCache[name] = img
		return img
	}

	p.iconCache[name] = nil
	return placeholderIcon()
}

func (p *Provider) findIconFile(name string) image.Image {
	extensions := []string{".png", ".svg"}
	sizeDirs := []string{"48x48", "64x64", "128x128", "scalable", ""}

	searchDirs := iconSearchDirs()

	for _, dir := range searchDirs {
		for _, sizeDir := range sizeDirs {
			var searchPath string
			if sizeDir != "" {
				searchPath = filepath.Join(dir, sizeDir, "apps")
			} else {
				searchPath = dir
			}

			for _, ext := range extensions {
				candidate := filepath.Join(searchPath, name+ext)
				if f, err := os.Open(candidate); err == nil {
					var img image.Image
					if ext == ".png" {
						img, err = png.Decode(f)
					} else if ext == ".svg" {
						img, err = decodeSVG(f, 48)
					}
					f.Close()
					if img != nil && err == nil {
						return resizeIcon(img, 48)
					}
				}
			}
		}
	}

	if filepath.IsAbs(name) {
		if f, err := os.Open(name); err == nil {
			var img image.Image
			lower := strings.ToLower(name)
			if strings.HasSuffix(lower, ".png") {
				img, err = png.Decode(f)
			} else if strings.HasSuffix(lower, ".svg") {
				img, err = decodeSVG(f, 48)
			}
			f.Close()
			if img != nil && err == nil {
				return resizeIcon(img, 48)
			}
		}
	}

	return nil
}

func iconSearchDirs() []string {
	var dirs []string

	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs,
			filepath.Join(home, ".local", "share", "icons", "hicolor"),
			filepath.Join(home, ".icons"),
		)
	}

	dirs = append(dirs,
		"/usr/share/icons/hicolor",
		"/usr/share/icons",
		"/usr/share/pixmaps",
		"/usr/local/share/icons",
		"/usr/local/share/pixmaps",
	)

	return dirs
}

func decodeSVG(f *os.File, size int) (image.Image, error) {
	icon, err := oksvg.ReadIconStream(f)
	if err != nil {
		return nil, err
	}
	icon.SetTarget(0, 0, float64(size), float64(size))
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	dasher := rasterx.NewDasher(size, size, scanner)
	icon.Draw(dasher, 1)
	return img, nil
}

func resizeIcon(img image.Image, size int) image.Image {
	b := img.Bounds()
	if b.Dx() == size && b.Dy() == size {
		return img
	}
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	scaleX := float64(b.Dx()) / float64(size)
	scaleY := float64(b.Dy()) / float64(size)
	for y := range size {
		for x := range size {
			sx := int(float64(x) * scaleX)
			sy := int(float64(y) * scaleY)
			dst.Set(x, y, img.At(b.Min.X+sx, b.Min.Y+sy))
		}
	}
	return dst
}

func placeholderIcon() image.Image {
	const s = 48
	img := image.NewRGBA(image.Rect(0, 0, s, s))
	bg := color.NRGBA{R: 80, G: 80, B: 100, A: 255}
	for y := range s {
		for x := range s {
			img.Set(x, y, bg)
		}
	}
	return img
}

func launch(execLine string) error {
	cmd := cleanExec(execLine)
	parts := splitCommand(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	return system.Launch(parts[0], parts[1:]...)
}

func cleanExec(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '%' && i+1 < len(s) {
			switch s[i+1] {
			case '%':
				b.WriteByte('%')
			}
			i += 2
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return strings.TrimSpace(b.String())
}

func splitCommand(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for _, c := range s {
		switch {
		case c == '"':
			inQuotes = !inQuotes
		case c == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(c)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}
