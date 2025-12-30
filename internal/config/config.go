package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Sources    []Source `yaml:"sources"`
	OutputDir  string   `yaml:"output_dir"`
	Interval   int      `yaml:"interval_seconds"`
	HTTPClient struct {
		Timeout int `yaml:"timeout_seconds"`
	} `yaml:"http_client"`
}

type Source struct {
	URL      string `yaml:"url"`
	Filename string `yaml:"filename"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.HTTPClient.Timeout == 0 {
		cfg.HTTPClient.Timeout = 30
	}

	return &cfg, nil
}
