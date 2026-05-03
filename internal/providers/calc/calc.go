package calc

import (
	"fmt"
	"os/exec"
	"strings"

	"chameth.com/glauncher/internal/assets"
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
		Icon:        assets.Calculator(48),
		Query:       query,
		Exec: func(text string) func() error {
			return func() error {
				return copyToClipboard(text)
			}
		}(formatted),
	}}
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
