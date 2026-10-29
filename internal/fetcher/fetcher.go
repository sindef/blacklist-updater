package fetcher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"blacklistupdater/internal/config"
	"blacklistupdater/internal/formatter"
	"blacklistupdater/internal/logger"
	"blacklistupdater/internal/validator"
)

type State struct {
	ETag      string
	Content   string
	LastCheck time.Time
}

type Fetcher struct {
	client    *http.Client
	outputDir string
	stateMap  map[string]*State
	fileHash  map[string]string
	log       *logger.Logger
	whitelist []string
}

func New(client *http.Client, outputDir string, log *logger.Logger, whitelist []string) *Fetcher {
	return &Fetcher{
		client:    client,
		outputDir: outputDir,
		stateMap:  make(map[string]*State),
		fileHash:  make(map[string]string),
		log:       log,
		whitelist: whitelist,
	}
}

func (f *Fetcher) FetchAll(sources []config.Source) error {
	f.log.Debug("Starting fetch cycle for %d sources", len(sources))

	var filenames []string
	groups := make(map[string][]config.Source)
	for _, source := range sources {
		if _, ok := groups[source.Filename]; !ok {
			filenames = append(filenames, source.Filename)
		}
		groups[source.Filename] = append(groups[source.Filename], source)
	}

	for _, filename := range filenames {
		if err := f.update(filename, groups[filename]); err != nil {
			f.log.Error("Error updating %s: %v", filename, err)
		}
	}

	f.log.Debug("Fetch cycle completed")
	return nil
}

func (f *Fetcher) update(filename string, sources []config.Source) error {
	format := sources[0].OutputFormat
	for _, source := range sources[1:] {
		if source.OutputFormat != format {
			return fmt.Errorf("sources for %s use different output formats", filename)
		}
	}

	var bodies []string
	for _, source := range sources {
		body, err := f.Fetch(source)
		if err != nil {
			return fmt.Errorf("%s: %w", source.URL, err)
		}
		bodies = append(bodies, body)
	}

	merged := bodies[0]
	if len(bodies) > 1 {
		merged = mergeBodies(bodies)
	}

	hash := calculateHash(merged)
	if f.fileHash[filename] == hash {
		f.log.Info("No changes for %s (content unchanged)", filename)
		return nil
	}

	outputPath := filepath.Join(f.outputDir, filename)
	tmpPath := outputPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(formatter.Header(format)+merged), 0644); err != nil {
		f.log.Debug("Failed to write file: %v", err)
		return err
	}
	if err := os.Rename(tmpPath, outputPath); err != nil {
		f.log.Debug("Failed to replace file: %v", err)
		return err
	}

	f.fileHash[filename] = hash
	f.log.Info("Updated %s from %d source(s)", filename, len(sources))
	return nil
}

func (f *Fetcher) Fetch(source config.Source) (string, error) {
	f.log.Debug("Fetching %s -> %s", source.URL, source.Filename)

	req, err := http.NewRequest("GET", source.URL, nil)
	if err != nil {
		f.log.Debug("Failed to create request: %v", err)
		return "", err
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
		return "", err
	}
	defer resp.Body.Close()

	f.log.Debug("HTTP response: status=%d, content-length=%d, etag=%s", resp.StatusCode, resp.ContentLength, resp.Header.Get("ETag"))

	if resp.StatusCode == http.StatusNotModified && exists {
		f.log.Debug("Source %s not modified (304)", source.URL)
		return state.Content, nil
	}

	if resp.StatusCode != http.StatusOK {
		f.log.Debug("Unexpected status code: %d", resp.StatusCode)
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		f.log.Debug("Failed to read response body: %v", err)
		return "", err
	}

	f.log.Debug("Read %d bytes from %s", len(body), source.URL)

	content := string(body)
	if !validator.ValidateHostsFile(content) {
		f.log.Debug("Validation failed for %s", source.URL)
		return "", fmt.Errorf("validation failed: content is not a valid hosts file")
	}

	f.log.Debug("Content validation passed")

	var converted string
	switch source.OutputFormat {
	case "hosts":
		converted, err = formatter.ConvertToHosts(content, f.whitelist)
	case "dnsmasq":
		converted, err = formatter.ConvertToDNSmasq(content, f.whitelist)
	case "rfc1035":
		converted, err = formatter.ConvertToRFC1035(content, f.whitelist, source.Wildcards)
	case "rpz":
		converted, err = formatter.ConvertToRPZ(content, f.whitelist)
	default:
		converted = formatter.FilterWhitelist(content, f.whitelist)
	}
	if err != nil {
		f.log.Debug("Format conversion failed: %v", err)
		return "", fmt.Errorf("format conversion failed: %w", err)
	}

	if state == nil {
		state = &State{}
		f.stateMap[source.URL] = state
	}
	state.ETag = resp.Header.Get("ETag")
	state.Content = converted
	state.LastCheck = time.Now()

	f.log.Debug("State updated: ETag=%s, LastCheck=%s", state.ETag, state.LastCheck.Format(time.RFC3339))
	return converted, nil
}

func mergeBodies(bodies []string) string {
	seen := make(map[string]bool)
	var result []string
	for _, body := range bodies {
		for _, line := range strings.Split(body, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "!") {
				continue
			}
			if seen[trimmed] {
				continue
			}
			seen[trimmed] = true
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n") + "\n"
}

func calculateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
