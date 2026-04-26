package main

import (
	"log"

	"chameth.com/glauncher/internal/code"
	"chameth.com/glauncher/internal/config"
	"chameth.com/glauncher/internal/desktop"
	"chameth.com/glauncher/internal/folders"
	"chameth.com/glauncher/internal/search"
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

	app := ui.New(cfg.Theme, providers...)
	app.Run()
}
