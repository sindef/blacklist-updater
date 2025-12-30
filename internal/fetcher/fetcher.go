package fetcher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dnsblacklist/internal/config"
	"dnsblacklist/internal/validator"
)

type State struct {
	ETag      string
	LastHash  string
	LastCheck time.Time
}

type Fetcher struct {
	client    *http.Client
	outputDir string
	stateMap  map[string]*State
}

func New(client *http.Client, outputDir string) *Fetcher {
	return &Fetcher{
		client:    client,
		outputDir: outputDir,
		stateMap:  make(map[string]*State),
	}
}

func (f *Fetcher) FetchAll(sources []config.Source) error {
	for _, source := range sources {
		if err := f.Fetch(source); err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching %s: %v\n", source.URL, err)
			continue
		}
	}
	return nil
}

func (f *Fetcher) Fetch(source config.Source) error {
	req, err := http.NewRequest("GET", source.URL, nil)
	if err != nil {
		return err
	}

	state, exists := f.stateMap[source.URL]
	if exists && state.ETag != "" {
		req.Header.Set("If-None-Match", state.ETag)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	content := string(body)
	if !validator.ValidateHostsFile(content) {
		return fmt.Errorf("validation failed: content is not a valid hosts file")
	}

	hash := calculateHash(content)
	if state != nil && state.LastHash == hash {
		return nil
	}

	outputPath := filepath.Join(f.outputDir, source.Filename)
	if err := os.WriteFile(outputPath, body, 0644); err != nil {
		return err
	}

	if state == nil {
		state = &State{}
		f.stateMap[source.URL] = state
	}

	state.ETag = resp.Header.Get("ETag")
	state.LastHash = hash
	state.LastCheck = time.Now()

	fmt.Printf("Updated %s from %s\n", source.Filename, source.URL)
	return nil
}

func calculateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
