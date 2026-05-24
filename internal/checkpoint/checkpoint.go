// Package checkpoint provides persistent cycle checkpointing for driftwatch,
// allowing the daemon to resume from the last known-good scan position after restart.
package checkpoint

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry records the outcome of a single completed scan cycle.
type Entry struct {
	CycleID   string    `json:"cycle_id"`
	Timestamp time.Time `json:"timestamp"`
	Manifests int       `json:"manifests"`
	DriftCount int      `json:"drift_count"`
	Error     string    `json:"error,omitempty"`
}

// Store persists checkpoint entries to a JSON file on disk.
type Store struct {
	mu   sync.RWMutex
	path string
}

// New returns a Store that writes checkpoints to the given file path.
func New(path string) *Store {
	return &Store{path: path}
}

// Save writes the entry to disk, replacing any previous checkpoint.
func (s *Store) Save(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Load reads the most recent checkpoint from disk.
// Returns nil, nil when no checkpoint exists yet.
func (s *Store) Load() (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// Delete removes the checkpoint file from disk.
func (s *Store) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
