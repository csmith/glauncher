package calc

import (
	"strings"

	"chameth.com/glauncher/internal/assets"
	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/system"
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
				return system.CopyToClipboard(text)
			}
		}(formatted),
	}}
}
