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

	"blacklistupdater/internal/config"
	"blacklistupdater/internal/logger"
	"blacklistupdater/internal/validator"
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
	log       *logger.Logger
}

func New(client *http.Client, outputDir string, log *logger.Logger) *Fetcher {
	return &Fetcher{
		client:    client,
		outputDir: outputDir,
		stateMap:  make(map[string]*State),
		log:       log,
	}
}

func (f *Fetcher) FetchAll(sources []config.Source) error {
	f.log.Debug("Starting fetch cycle for %d sources", len(sources))
	for _, source := range sources {
		if err := f.Fetch(source); err != nil {
			f.log.Error("Error fetching %s: %v", source.URL, err)
			continue
		}
	}
	f.log.Debug("Fetch cycle completed")
	return nil
}

func (f *Fetcher) Fetch(source config.Source) error {
	f.log.Debug("Fetching %s -> %s", source.URL, source.Filename)

	req, err := http.NewRequest("GET", source.URL, nil)
	if err != nil {
		f.log.Debug("Failed to create request: %v", err)
		return err
	}

	state, exists := f.stateMap[source.URL]
	if exists && state.ETag != "" {
		req.Header.Set("If-None-Match", state.ETag)
		f.log.Debug("Using ETag: %s", state.ETag)
	}

	f.log.Debug("Sending HTTP GET request to %s", source.URL)
	resp, err := f.client.Do(req)
	if err != nil {
		f.log.Debug("HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	f.log.Debug("HTTP response: status=%d, content-length=%d, etag=%s", resp.StatusCode, resp.ContentLength, resp.Header.Get("ETag"))

	if resp.StatusCode == http.StatusNotModified {
		f.log.Info("No changes for %s (not modified)", source.Filename)
		f.log.Debug("Source %s not modified (304)", source.URL)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		f.log.Debug("Unexpected status code: %d", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		f.log.Debug("Failed to read response body: %v", err)
		return err
	}

	f.log.Debug("Read %d bytes from %s", len(body), source.URL)

	content := string(body)
	if !validator.ValidateHostsFile(content) {
		f.log.Debug("Validation failed for %s", source.URL)
		return fmt.Errorf("validation failed: content is not a valid hosts file")
	}

	f.log.Debug("Content validation passed")
	hash := calculateHash(content)
	f.log.Debug("Content hash: %s", hash)

	if state != nil && state.LastHash == hash {
		f.log.Info("No changes for %s (content unchanged)", source.Filename)
		f.log.Debug("Content unchanged for %s (hash match)", source.URL)
		return nil
	}

	outputPath := filepath.Join(f.outputDir, source.Filename)
	if err := os.WriteFile(outputPath, body, 0644); err != nil {
		f.log.Debug("Failed to write file: %v", err)
		return err
	}

	if state == nil {
		state = &State{}
		f.stateMap[source.URL] = state
	}

	state.ETag = resp.Header.Get("ETag")
	state.LastHash = hash
	state.LastCheck = time.Now()

	f.log.Info("Updated %s from %s", source.Filename, source.URL)
	f.log.Debug("State updated: ETag=%s, LastCheck=%s", state.ETag, state.LastCheck.Format(time.RFC3339))
	return nil
}

func calculateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
