// Package fingerprint provides structural fingerprinting of manifests,
// enabling fast change detection across polling cycles without full field comparison.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

// Manifest is the minimal interface required for fingerprinting.
type Manifest interface {
	ServiceName() string
	Kind() string
	Spec() map[string]any
}

// Store holds fingerprints keyed by service name.
type Store struct {
	mu    sync.RWMutex
	index map[string]string
}

// New returns an empty fingerprint Store.
func New() *Store {
	return &Store{
		index: make(map[string]string),
	}
}

// Compute returns a deterministic SHA-256 hex fingerprint for the given manifest.
// The fingerprint covers the service name, kind, and full spec.
func Compute(m Manifest) (string, error) {
	payload := map[string]any{
		"service": m.ServiceName(),
		"kind":    m.Kind(),
		"spec":    m.Spec(),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("fingerprint: marshal failed: %w", err)
	}

	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// Set stores the fingerprint for the named service.
func (s *Store) Set(service, fp string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index[service] = fp
}

// Get returns the stored fingerprint and whether it exists.
func (s *Store) Get(service string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fp, ok := s.index[service]
	return fp, ok
}

// Changed returns true if the manifest's computed fingerprint differs from the
// stored one, or if no fingerprint has been recorded yet.
func (s *Store) Changed(m Manifest) (bool, error) {
	current, err := Compute(m)
	if err != nil {
		return false, err
	}

	prev, ok := s.Get(m.ServiceName())
	if !ok {
		return true, nil
	}
	return current != prev, nil
}

// Update computes and stores the latest fingerprint for the manifest.
func (s *Store) Update(m Manifest) error {
	fp, err := Compute(m)
	if err != nil {
		return err
	}
	s.Set(m.ServiceName(), fp)
	return nil
}
