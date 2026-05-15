package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/driftwatch/internal/manifest"
)

func writeTemp(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
}

func TestLoad_ValidManifest(t *testing.T) {
	dir := t.TempDir()
	writeTemp(t, dir, "service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  replicas: "3"
`)

	l := manifest.NewLoader(dir)
	m, err := l.Load("service.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Kind != "Service" {
		t.Errorf("expected kind=Service, got %q", m.Kind)
	}
	if m.Metadata["name"] != "my-service" {
		t.Errorf("expected name=my-service, got %q", m.Metadata["name"])
	}
}

func TestLoad_MissingKind(t *testing.T) {
	dir := t.TempDir()
	writeTemp(t, dir, "bad.yaml", `apiVersion: v1\nmetadata:\n  name: oops\n`)

	l := manifest.NewLoader(dir)
	_, err := l.Load("bad.yaml")
	if err == nil {
		t.Fatal("expected error for missing kind, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	l := manifest.NewLoader(t.TempDir())
	_, err := l.Load("nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadAll(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []struct{ name, content string }{
		{"a.yaml", "kind: Deployment\nmetadata:\n  name: a\nspec: {}\n"},
		{"b.yml", "kind: ConfigMap\nmetadata:\n  name: b\nspec: {}\n"},
	} {
		writeTemp(t, dir, f.name, f.content)
	}

	l := manifest.NewLoader(dir)
	manifests, err := l.LoadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifests) != 2 {
		t.Errorf("expected 2 manifests, got %d", len(manifests))
	}
}
