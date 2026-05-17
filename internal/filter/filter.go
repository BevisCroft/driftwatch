// Package filter provides namespace and label-based filtering for manifests
// before drift detection is performed.
package filter

import (
	"strings"

	"github.com/driftwatch/internal/manifest"
)

// Options holds the filtering criteria applied to a set of manifests.
type Options struct {
	// Namespaces restricts processing to manifests in the given namespaces.
	// An empty slice means all namespaces are accepted.
	Namespaces []string

	// LabelSelector filters manifests by a single key=value label expression.
	// An empty string disables label filtering.
	LabelSelector string
}

// Filter returns the subset of manifests that match all criteria in opts.
func Filter(manifests []manifest.Manifest, opts Options) []manifest.Manifest {
	var result []manifest.Manifest
	for _, m := range manifests {
		if !matchesNamespace(m, opts.Namespaces) {
			continue
		}
		if !matchesLabel(m, opts.LabelSelector) {
			continue
		}
		result = append(result, m)
	}
	return result
}

// matchesNamespace returns true when the manifest namespace is in the allow
// list, or when the allow list is empty (no restriction).
func matchesNamespace(m manifest.Manifest, namespaces []string) bool {
	if len(namespaces) == 0 {
		return true
	}
	for _, ns := range namespaces {
		if strings.EqualFold(m.Namespace, ns) {
			return true
		}
	}
	return false
}

// matchesLabel returns true when the manifest carries the requested label, or
// when the selector is empty. The selector must be in "key=value" format.
func matchesLabel(m manifest.Manifest, selector string) bool {
	if selector == "" {
		return true
	}
	parts := strings.SplitN(selector, "=", 2)
	if len(parts) != 2 {
		return false
	}
	key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	if m.Labels == nil {
		return false
	}
	return m.Labels[key] == value
}
