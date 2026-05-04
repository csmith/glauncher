# glauncher, a native application quick launcher for Linux

glauncher is an application launcher written in Go. It's
similar to [ulauncher](https://ulauncher.io/) but doesn't
require webkit or python. ulauncher is much more fully
featured with a large ecosystem of extensions, though,
so it's worth checking out if glauncher doesn't fit your
needs.

glauncher shows a search UI when it opens, and exits
if that UI loses focus. It's designed to be invoked in
response to a keyboard shortcut (e.g. in your desktop
environment's "keyboard" settings). It does _not_ run
in the background.

## Features

glauncher has several results providers which can each
be enabled or disabled in the config.

### Desktop applications

Scans XDG Desktop Entries to find launchable applications.
Select an app to launch it.

![](./screenshots/apps.png?raw=true)

### Folders

Lists folders and opens them in the default app.

![](./screenshots/folders.png?raw=true)

### Calculator

Works out simple calculations. Supports `+`, `-`, `/`, `*`,
`^`, `%` and parentheses. Select an entry to copy the result
to the clipboard.

![](./screenshots/calculator.png?raw=true)

### Code / IDE launcher

Requires configuring. See the [example config](./internal/config/config.example.yml).
Finds project folders in a directory, and allows launching them
directly with your IDE of choice. Requires the prefix "code".

![](./screenshots/code.png?raw=true)

### Arch linux packages

Searches for Arch and AUR packages. Opens the corresponding
webpage if selected.

![](./screenshots/arch.png?raw=true)

### Web search

Launches a configured website with the search terms populated.
Can be configured to always show search suggestions, or use
the site name/alias as a prefix.

![](./screenshots/search.png?raw=true)

## Getting started

Currently you need to manually build and install the project:

```
go install chameth.com/glauncher@latest
```

Then configure a keyboard shortcut to launch the
binary (probably `~/go/bin/glauncher`).

- In XFCE, go to settings -> keyboard -> application shortcuts

The first time it's opened, glauncher will place a default
config file in `~/.config/glauncher/` (or wherever your
`XDG_CONFIG_HOME` directory is).

## Provenance

This project was primarily created with an LLM, but with a strong guiding
hand. It's not "vibe coded", but an LLM was still the primary author of most
lines of code. I believe it meets the same sort of standards I'd aim for with
hand-crafted code, but some slop may slip through. I understand if you
prefer not to use LLM-created software, and welcome human-authored alternatives
(I just don't personally have the time/motivation to do so).

## Feedback / Contributing

Feedback, feature requests, bug reports and pull requests are all welcome!
