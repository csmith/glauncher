package main

import (
	"chameth.com/glauncher/internal/desktop"
	"chameth.com/glauncher/internal/ui"
)

func main() {
	dp := desktop.NewProvider()
	app := ui.New(dp)
	app.Run()
}
