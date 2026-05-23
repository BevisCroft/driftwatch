// Package digest computes and compares stable content hashes for manifests,
// enabling fast drift detection by comparing digests before doing field-level
// comparison.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/driftwatch/internal/manifest"
)

// Digester computes SHA-256 digests for manifest content.
type Digester struct{}

// New returns a new Digester.
func New() *Digester {
	return &Digester{}
}

// Compute returns the hex-encoded SHA-256 digest of the manifest's spec and
// metadata fields. The kind and name are included so that renames are caught.
func (d *Digester) Compute(m manifest.Manifest) (string, error) {
	payload := map[string]any{
		"kind":      m.Kind,
		"name":      m.Name,
		"namespace": m.Namespace,
		"spec":      m.Spec,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("digest: marshal manifest %q: %w", m.Name, err)
	}

	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// Equal returns true when the digests of two manifests match, indicating no
// drift. An error is returned if either manifest cannot be hashed.
func (d *Digester) Equal(a, b manifest.Manifest) (bool, error) {
	ha, err := d.Compute(a)
	if err != nil {
		return false, err
	}

	hb, err := d.Compute(b)
	if err != nil {
		return false, err
	}

	return ha == hb, nil
}
