// Package snapshot provides functionality for capturing and persisting
// the state of deployed service manifests at a point in time.
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Snapshot represents a captured state of a manifest at a specific time.
type Snapshot struct {
	Name      string                 `json:"name"`
	Kind      string                 `json:"kind"`
	CapturedAt time.Time             `json:"captured_at"`
	Fields    map[string]interface{} `json:"fields"`
}

// Store manages reading and writing snapshots to disk.
type Store struct {
	dir string
}

// NewStore creates a new Store that persists snapshots under dir.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("snapshot: create dir %q: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Save writes a snapshot to disk, overwriting any previous snapshot
// for the same manifest name.
func (s *Store) Save(snap Snapshot) error {
	snap.CapturedAt = time.Now().UTC()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot: marshal %q: %w", snap.Name, err)
	}
	dest := filepath.Join(s.dir, snap.Name+".json")
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("snapshot: write %q: %w", dest, err)
	}
	return nil
}

// Load reads a previously saved snapshot for the given manifest name.
// Returns os.ErrNotExist if no snapshot has been saved yet.
func (s *Store) Load(name string) (Snapshot, error) {
	src := filepath.Join(s.dir, name+".json")
	data, err := os.ReadFile(src)
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: read %q: %w", src, err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: unmarshal %q: %w", name, err)
	}
	return snap, nil
}
