package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"dnsblacklist/internal/config"
	"dnsblacklist/internal/fetcher"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config.yaml>\n", os.Args[0])
		os.Exit(1)
	}

	configPath := os.Args[1]
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.HTTPClient.Timeout) * time.Second,
	}

	f := fetcher.New(client, cfg.OutputDir)

	if err := f.FetchAll(cfg.Sources); err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching sources: %v\n", err)
		os.Exit(1)
	}

	if cfg.Interval > 0 {
		ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := f.FetchAll(cfg.Sources); err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching sources: %v\n", err)
			}
		}
	}
}
