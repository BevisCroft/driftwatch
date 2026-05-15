package config_test

import (
	"os"
	"testing"

	"github.com/you/driftwatch/internal/config"
)

// TestLoad_EnvOverride verifies that callers can layer environment-driven
// paths on top of file-based config without modifying the loader itself.
func TestLoad_EnvOverride(t *testing.T) {
	raw := `
manifest_dir: ./manifests
poll_interval: 10s
reporter:
  format: json
`
	path := writeTemp(t, raw)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Simulate an operator overriding manifest_dir via env after load.
	if override := os.Getenv("DRIFTWATCH_MANIFEST_DIR"); override != "" {
		cfg.ManifestDir = override
	}

	if cfg.ManifestDir == "" {
		t.Error("manifest_dir must not be empty after optional env override")
	}
}

// TestLoad_ZeroPollInterval ensures a zero poll_interval is rejected.
func TestLoad_ZeroPollInterval(t *testing.T) {
	path := writeTemp(t, "manifest_dir: ./m\npoll_interval: 0s\n")
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for zero poll_interval, got nil")
	}
}
