package config

import (
	"os"
	"path/filepath"

	cconfig "github.com/csmith/config"
)

type Config struct {
	Theme   ThemeConfig   `yaml:"theme"`
	Desktop DesktopConfig `yaml:"desktop"`
	Code    CodeConfig    `yaml:"code"`
	Folders FoldersConfig `yaml:"folders"`
	Arch    ArchConfig    `yaml:"arch"`
	Calc    CalcConfig    `yaml:"calc"`
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

	_, err := cconfig.Load(&cfg, cconfig.DirectoryName("glauncher"), cconfig.FileName("config.yml"))
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
