package main

import (
	"log"

	"chameth.com/glauncher/internal/code"
	"chameth.com/glauncher/internal/desktop"
	"chameth.com/glauncher/internal/folders"
	"chameth.com/glauncher/internal/ui"
)

func main() {
	dp := desktop.NewProvider()

	cp, err := code.NewProvider()
	if err != nil {
		log.Fatal(err)
	}

	fp := folders.NewProvider()

	app := ui.New(dp, cp, fp)
	app.Run()
}
