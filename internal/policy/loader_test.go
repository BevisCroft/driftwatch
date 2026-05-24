package policy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func writePolicy(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "policy-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestLoadFile_ValidPolicy(t *testing.T) {
	path := writePolicy(t, `
rules:
  - service: "api"
    field: "spec.replicas"
    severity: warn
  - service: "*"
    field: "spec.image"
    severity: error
`)
	rules, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Severity != drift.SeverityWarn {
		t.Errorf("rule[0]: expected warn, got %s", rules[0].Severity)
	}
	if rules[1].Severity != drift.SeverityError {
		t.Errorf("rule[1]: expected error, got %s", rules[1].Severity)
	}
}

func TestLoadFile_InvalidSeverity(t *testing.T) {
	path := writePolicy(t, `
rules:
  - service: "*"
    field: "*"
    severity: critical
`)
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for unknown severity, got nil")
	}
}

func TestLoadFile_FileNotFound(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadFile_EmptyRules(t *testing.T) {
	path := writePolicy(t, "rules: []\n")
	rules, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(rules))
	}
}
