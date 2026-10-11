# AGENTS.md

## Project

Gio UI-based Linux application launcher. Reads `.desktop` files from XDG data directories, provides a search-driven GUI to find and launch apps, with additional providers for code projects and folder navigation.

Module: `chameth.com/glauncher` | Go 1.26.2 | License: MIT

## Makefile targets

- `make build` — compile and produce `glauncher` binary (requires CGO + Wayland/X11 dev headers)
- `make verify` — build, vet, fix, staticcheck, and fmt

## Verification

Always run `make verify` after making changes, and `go test ./...`.

CI uses Forgejo Actions with reusable workflows (`meta/workflows`) that run `go build` and `go test` on every PR.

## Architecture

- `main.go` — entrypoint; loads unified config, conditionally creates enabled providers
- `internal/config/` — unified config loading from `~/.config/glauncher/config.yml` (`config.example.yml` in repo root shows all options)
- `internal/search/` — `search.Result` struct and result sorting
- `internal/providers/desktop/` — implements `ui.Provider`: parses `.desktop` files, loads icons, launches apps
- `internal/providers/code/` — implements `ui.Provider`: scans a directory for version-controlled projects, opens them in an editor
- `internal/providers/folders/` — implements `ui.Provider`: opens directories by path
- `internal/providers/arch/` — implements `ui.Provider`: searches Arch Linux and AUR packages
- `internal/providers/calc/` — implements `ui.Provider`: evaluates math expressions
- `internal/providers/files/` — implements `ui.Provider`: one instance per configured launcher; lists files matching glob patterns under a directory and opens them with a command
- `internal/providers/searchweb/` — implements `ui.Provider`: opens web searches
- `internal/ui/` — Gio UI window: search input, arrow-key navigation, result list with icons; defines the `ui.Provider` interface

## Configuration

Single `config.yml` loaded from `~/.config/glauncher/config.yml`. Each provider has its own top-level section with an `enabled` bool (defaults to `false`). The `theme` section controls colours and font. See `config.example.yml` for all options.

Adding a new search source means implementing `ui.Provider`, adding a config section with an `enabled` field to `internal/config`, and wiring it in `main.go`.

## Key details

- Gio UI (`gioui.org`) is the GUI toolkit — requires CGO and platform graphics libraries (Wayland/X11 headers) to build
- Config uses `github.com/csmith/config` for YAML loading from XDG config dir
- Tests use `github.com/stretchr/testify` for assertions, in `_test.go` files alongside the code they cover
- CI is on Forgejo (not GitHub), defined in `.forgejo/workflows/`
