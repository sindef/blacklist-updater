package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"blacklistupdater/internal/config"
	"blacklistupdater/internal/fetcher"
	"blacklistupdater/internal/logger"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "Error: config file path is required\n")
		flag.Usage()
		os.Exit(1)
	}

	log := logger.New(*debug)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("Error loading config: %v", err)
		os.Exit(1)
	}

	log.Info("Starting blacklistupdater at %s", log.Timestamp())
	log.Info("Configuration:")
	log.Info("  Output directory: %s", cfg.OutputDir)
	log.Info("  Update interval: %d seconds", cfg.Interval)
	log.Info("  HTTP timeout: %d seconds", cfg.HTTPClient.Timeout)
	log.Info("  Whitelist entries: %d", len(cfg.Whitelist))
	log.Info("  Sources: %d", len(cfg.Sources))
	for i, source := range cfg.Sources {
		log.Info("    [%d] %s -> %s", i+1, source.URL, source.Filename)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		log.Error("Error creating output directory: %v", err)
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.HTTPClient.Timeout) * time.Second,
	}

	f := fetcher.New(client, cfg.OutputDir, log, cfg.Whitelist)

	if err := f.FetchAll(cfg.Sources); err != nil {
		log.Error("Error fetching sources: %v", err)
		os.Exit(1)
	}

	if cfg.Interval > 0 {
		log.Info("Monitoring enabled, next update at %s", time.Now().Add(time.Duration(cfg.Interval)*time.Second).Format(time.RFC3339))
		ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := f.FetchAll(cfg.Sources); err != nil {
				log.Error("Error fetching sources: %v", err)
			}
			log.Info("Next update scheduled for %s", time.Now().Add(time.Duration(cfg.Interval)*time.Second).Format(time.RFC3339))
		}
	}
}
