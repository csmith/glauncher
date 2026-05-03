package config

import (
	"bytes"
	_ "embed"
	"io"
	"os"
	"path/filepath"

	cconfig "github.com/csmith/config"
)

//go:embed config.example.yml
var defaultConfig []byte

type Config struct {
	Theme     ThemeConfig     `yaml:"theme"`
	Desktop   DesktopConfig   `yaml:"desktop"`
	Code      CodeConfig      `yaml:"code"`
	Folders   FoldersConfig   `yaml:"folders"`
	Arch      ArchConfig      `yaml:"arch"`
	Calc      CalcConfig      `yaml:"calc"`
	SearchWeb SearchWebConfig `yaml:"search"`
}

type ThemeConfig struct {
	Background string `yaml:"background"`
	Divider    string `yaml:"divider"`
	Primary    string `yaml:"primary"`
	Secondary  string `yaml:"secondary"`
	Selection  string `yaml:"selection"`
	Font       string `yaml:"font"`
}

type DesktopConfig struct {
	Enabled bool `yaml:"enabled"`
}

type CodeConfig struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
	Command string `yaml:"command"`
}

type FoldersConfig struct {
	Enabled bool `yaml:"enabled"`
}

type ArchConfig struct {
	Enabled bool `yaml:"enabled"`
}

type CalcConfig struct {
	Enabled bool `yaml:"enabled"`
}

type SearchWebConfig struct {
	Enabled   bool             `yaml:"enabled"`
	Providers []SearchWebEntry `yaml:"providers"`
}

type SearchWebEntry struct {
	Name          string   `yaml:"name"`
	Aliases       []string `yaml:"aliases"`
	URL           string   `yaml:"url"`
	AlwaysInclude bool     `yaml:"always_include"`
}

func Load() (*Config, error) {
	home, _ := os.UserHomeDir()

	cfg := Config{
		Theme: ThemeConfig{
			Background: "#1e1e2ef0",
			Divider:    "#3c3c50",
			Primary:    "#cdd6f4",
			Secondary:  "#a0a0b4",
			Selection:  "#6496e6c8",
		},
		Code: CodeConfig{
			Dir:     filepath.Join(home, "code"),
			Command: "code %s",
		},
	}

	_, err := cconfig.Load(&cfg,
		cconfig.DirectoryName("glauncher"),
		cconfig.FileName("config.yml"),
		cconfig.DefaultConfig(func() io.Reader {
			return bytes.NewReader(defaultConfig)
		}),
	)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
