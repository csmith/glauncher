package calc

import (
	"fmt"
	"image"
	"image/color"
	"os/exec"
	"strings"

	"chameth.com/glauncher/internal/search"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) Search(query string) []search.Result {
	if !looksLikeExpression(query) {
		return nil
	}

	result, err := evaluate(query)
	if err != nil {
		return nil
	}

	formatted := formatResult(result)
	if formatted == "" {
		return nil
	}

	display := strings.TrimSpace(query) + " = " + formatted

	return []search.Result{{
		Name:        formatted,
		Description: display,
		Icon:        calcIcon(),
		Query:       query,
		Exec: func(text string) func() error {
			return func() error {
				return copyToClipboard(text)
			}
		}(formatted),
	}}
}

func calcIcon() image.Image {
	const s = 48
	img := image.NewRGBA(image.Rect(0, 0, s, s))
	bg := color.NRGBA{R: 80, G: 160, B: 100, A: 255}
	fg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	for y := range s {
		for x := range s {
			img.Set(x, y, bg)
		}
	}
	equals := [][2]int{
		{14, 16}, {15, 16}, {16, 16}, {17, 16}, {18, 16}, {19, 16}, {20, 16}, {21, 16}, {22, 16}, {23, 16},
		{14, 20}, {15, 20}, {16, 20}, {17, 20}, {18, 20}, {19, 20}, {20, 20}, {21, 20}, {22, 20}, {23, 20},
		{14, 28}, {15, 28}, {16, 28}, {17, 28}, {18, 28}, {19, 28}, {20, 28}, {21, 28}, {22, 28}, {23, 28},
		{14, 32}, {15, 32}, {16, 32}, {17, 32}, {18, 32}, {19, 32}, {20, 32}, {21, 32}, {22, 32}, {23, 32},
	}
	for _, p := range equals {
		if p[0] < s && p[1] < s {
			img.Set(p[0], p[1], fg)
		}
	}
	return img
}

func copyToClipboard(text string) error {
	var name string
	var args []string

	if p, _ := exec.LookPath("wl-copy"); p != "" {
		name = "wl-copy"
	} else if p, _ := exec.LookPath("xclip"); p != "" {
		name = "xclip"
		args = []string{"-selection", "clipboard"}
	} else if p, _ := exec.LookPath("xsel"); p != "" {
		name = "xsel"
		args = []string{"--clipboard", "--input"}
	} else {
		return fmt.Errorf("no clipboard utility found (install wl-copy, xclip, or xsel)")
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
