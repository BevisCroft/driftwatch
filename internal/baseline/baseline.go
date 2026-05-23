// Package baseline provides functionality for pinning and comparing
// approved configuration states for drift detection.
package baseline

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a pinned baseline for a single service.
type Entry struct {
	Service   string                 `json:"service"`
	PinnedAt  time.Time              `json:"pinned_at"`
	ApprovedBy string               `json:"approved_by"`
	Fields    map[string]interface{} `json:"fields"`
}

// Store persists and retrieves baseline entries on disk.
type Store struct {
	mu  sync.RWMutex
	dir string
}

// New creates a new Store rooted at dir, creating the directory if needed.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("baseline: create dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Pin saves an approved baseline entry for the given service.
func (s *Store) Pin(entry Entry) error {
	if entry.Service == "" {
		return errors.New("baseline: service name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	path := s.filePath(entry.Service)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("baseline: write %s: %w", path, err)
	}
	return nil
}

// Get retrieves the pinned baseline for a service.
// Returns (nil, nil) if no baseline exists.
func (s *Store) Get(service string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filePath(service))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("baseline: read %s: %w", service, err)
	}
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("baseline: unmarshal %s: %w", service, err)
	}
	return &entry, nil
}

// Delete removes the pinned baseline for a service.
func (s *Store) Delete(service string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.filePath(service)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("baseline: delete %s: %w", service, err)
	}
	return nil
}

func (s *Store) filePath(service string) string {
	return filepath.Join(s.dir, service+".baseline.json")
}
