package search

import (
	"image"
	"sort"
	"strings"
)

type Result struct {
	Name        string
	Description string
	Icon        image.Image
	Exec        func() error
	Query       string
	Priority    int
}

type Provider interface {
	Search(query string) []Result
}

type AsyncInitializer interface {
	Ready() <-chan struct{}
}

type AsyncSearchProvider interface {
	Provider
	SetInvalidate(f func())
}

func SortResults(results []Result, query string) {
	sort.Slice(results, func(i, j int) bool {
		pi := results[i].Priority
		pj := results[j].Priority
		if pi != pj {
			return pi > pj
		}

		qi := results[i].Query
		if qi == "" {
			qi = query
		}
		qj := results[j].Query
		if qj == "" {
			qj = query
		}

		si := score(strings.ToLower(results[i].Name), strings.ToLower(results[i].Description), strings.ToLower(qi))
		sj := score(strings.ToLower(results[j].Name), strings.ToLower(results[j].Description), strings.ToLower(qj))
		if si != sj {
			return si > sj
		}
		di := LevenshteinDistance(strings.ToLower(results[i].Name), strings.ToLower(qi))
		dj := LevenshteinDistance(strings.ToLower(results[j].Name), strings.ToLower(qj))
		if di != dj {
			return di < dj
		}
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})
}

func score(nameLower, descLower, q string) int {
	if strings.HasPrefix(nameLower, q) {
		return 100
	}
	if strings.Contains(nameLower, q) {
		return 50
	}
	if strings.Contains(descLower, q) {
		return 25
	}
	return 0
}

func LevenshteinDistance(s, t string) int {
	if s == t {
		return 0
	}
	if len(s) == 0 {
		return len(t)
	}
	if len(t) == 0 {
		return len(s)
	}

	prev := make([]int, len(t)+1)
	curr := make([]int, len(t)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(s); i++ {
		curr[0] = i
		for j := 1; j <= len(t); j++ {
			cost := 1
			if s[i-1] == t[j-1] {
				cost = 0
			}
			curr[j] = min(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(t)]
}
