package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/you/driftwatch/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "driftwatch-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTemp(t, "manifest_dir: ./my-manifests\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ManifestDir != "./my-manifests" {
		t.Errorf("manifest_dir = %q, want ./my-manifests", cfg.ManifestDir)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("poll_interval = %s, want 30s", cfg.PollInterval)
	}
	if cfg.Reporter.Format != "text" {
		t.Errorf("reporter.format = %q, want text", cfg.Reporter.Format)
	}
}

func TestLoad_FullConfig(t *testing.T) {
	raw := `
manifest_dir: /etc/manifests
poll_interval: 1m
log_level: debug
reporter:
  format: json
  out_file: /tmp/drift.json
`
	path := writeTemp(t, raw)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != time.Minute {
		t.Errorf("poll_interval = %s, want 1m", cfg.PollInterval)
	}
	if cfg.Reporter.Format != "json" {
		t.Errorf("reporter.format = %q, want json", cfg.Reporter.Format)
	}
	if cfg.Reporter.OutFile != "/tmp/drift.json" {
		t.Errorf("reporter.out_file = %q, want /tmp/drift.json", cfg.Reporter.OutFile)
	}
}

func TestLoad_InvalidFormat(t *testing.T) {
	path := writeTemp(t, "manifest_dir: ./m\nreporter:\n  format: xml\n")
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_EmptyManifestDir(t *testing.T) {
	path := writeTemp(t, "manifest_dir: \"\"\n")
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty manifest_dir, got nil")
	}
}
