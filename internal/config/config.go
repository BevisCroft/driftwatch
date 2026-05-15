// Package config handles loading and validating driftwatch daemon configuration.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level daemon configuration.
type Config struct {
	ManifestDir  string        `yaml:"manifest_dir"`
	PollInterval time.Duration `yaml:"poll_interval"`
	Reporter     ReporterConfig `yaml:"reporter"`
	LogLevel     string        `yaml:"log_level"`
}

// ReporterConfig controls how drift results are reported.
type ReporterConfig struct {
	Format  string `yaml:"format"`  // "text" or "json"
	OutFile string `yaml:"out_file"` // empty means stdout
}

// defaults returns a Config populated with sensible defaults.
func defaults() Config {
	return Config{
		ManifestDir:  "./manifests",
		PollInterval: 30 * time.Second,
		LogLevel:     "info",
		Reporter: ReporterConfig{
			Format: "text",
		},
	}
}

// Load reads a YAML config file from path and returns a validated Config.
// Missing fields fall back to defaults.
func Load(path string) (*Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation: %w", err)
	}

	return &cfg, nil
}

// validate checks that required fields have acceptable values.
func (c *Config) validate() error {
	if c.ManifestDir == "" {
		return fmt.Errorf("manifest_dir must not be empty")
	}
	if c.PollInterval <= 0 {
		return fmt.Errorf("poll_interval must be positive, got %s", c.PollInterval)
	}
	switch c.Reporter.Format {
	case "text", "json":
		// valid
	default:
		return fmt.Errorf("reporter.format must be \"text\" or \"json\", got %q", c.Reporter.Format)
	}
	return nil
}
