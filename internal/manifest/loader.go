package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manifest represents a parsed service manifest file.
type Manifest struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   map[string]string `yaml:"metadata"`
	Spec       map[string]any    `yaml:"spec"`
}

// Loader reads and parses manifest files from disk.
type Loader struct {
	BaseDir string
}

// NewLoader creates a Loader rooted at the given directory.
func NewLoader(baseDir string) *Loader {
	return &Loader{BaseDir: baseDir}
}

// Load reads a YAML manifest file relative to BaseDir and returns a Manifest.
func (l *Loader) Load(name string) (*Manifest, error) {
	path := filepath.Join(l.BaseDir, name)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("manifest loader: read %q: %w", path, err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("manifest loader: parse %q: %w", path, err)
	}

	if m.Kind == "" {
		return nil, fmt.Errorf("manifest loader: %q missing required field 'kind'", path)
	}

	return &m, nil
}

// LoadAll loads every *.yaml / *.yml file found in BaseDir (non-recursive).
func (l *Loader) LoadAll() ([]*Manifest, error) {
	patterns := []string{
		filepath.Join(l.BaseDir, "*.yaml"),
		filepath.Join(l.BaseDir, "*.yml"),
	}

	var manifests []*Manifest
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			m, err := l.Load(filepath.Base(match))
			if err != nil {
				return nil, err
			}
			manifests = append(manifests, m)
		}
	}
	return manifests, nil
}
