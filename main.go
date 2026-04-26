package main

import (
	"log"

	"chameth.com/glauncher/internal/arch"
	"chameth.com/glauncher/internal/calc"
	"chameth.com/glauncher/internal/code"
	"chameth.com/glauncher/internal/config"
	"chameth.com/glauncher/internal/desktop"
	"chameth.com/glauncher/internal/folders"
	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/searchweb"
	"chameth.com/glauncher/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	var providers []search.Provider

	if cfg.Desktop.Enabled {
		providers = append(providers, desktop.NewProvider())
	}

	if cfg.Code.Enabled {
		cp, err := code.NewProvider(cfg.Code.Dir, cfg.Code.Command)
		if err != nil {
			log.Fatal(err)
		}
		providers = append(providers, cp)
	}

	if cfg.Folders.Enabled {
		providers = append(providers, folders.NewProvider())
	}

	if cfg.Arch.Enabled {
		providers = append(providers, arch.NewProvider())
	}

	if cfg.Calc.Enabled {
		providers = append(providers, calc.NewProvider())
	}

	if cfg.SearchWeb.Enabled {
		var entries []searchweb.Entry
		for _, p := range cfg.SearchWeb.Providers {
			entries = append(entries, searchweb.Entry{
				Name:          p.Name,
				Aliases:       p.Aliases,
				URL:           p.URL,
				AlwaysInclude: p.AlwaysInclude,
			})
		}
		providers = append(providers, searchweb.NewProvider(entries))
	}

	app := ui.New(cfg.Theme, providers...)
	app.Run()
}
