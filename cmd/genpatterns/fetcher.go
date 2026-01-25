// Package main - fetcher.go fetches PHP source files from GitHub
package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher handles HTTP requests to GitHub
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new Fetcher with timeout
func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: timeout},
	}
}

// Fetch retrieves content from a URL
func (f *Fetcher) Fetch(url string) (string, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", url, err)
	}

	return string(body), nil
}

// FetchFrameworkSources fetches all source files for a framework
func (f *Fetcher) FetchFrameworkSources(fw *FrameworkDefinition) (map[string]string, error) {
	sources := make(map[string]string)

	for _, src := range fw.Sources {
		content, err := f.Fetch(src.URL)
		if err != nil {
			return nil, fmt.Errorf("framework %s: %w", fw.Name, err)
		}
		sources[src.ClassName] = content
	}

	return sources, nil
}
