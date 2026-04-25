package search

import "image"

type Result struct {
	Name        string
	Description string
	Icon        image.Image
	Exec        func() error
}

type Provider interface {
	Search(query string) []Result
}

type AsyncInitializer interface {
	Ready() <-chan struct{}
}
