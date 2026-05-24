// Package normalize provides utilities for canonicalizing manifest fields
// before comparison, ensuring that semantically equivalent representations
// do not produce false-positive drift results.
package normalize

import (
	"sort"
	"strings"
)

// Normalizer applies a set of registered field normalization rules to
// manifest spec maps before they are compared by the drift detector.
type Normalizer struct {
	rules []Rule
}

// Rule describes a single normalization transformation.
type Rule struct {
	// Field is the dot-separated spec key this rule applies to, e.g. "labels".
	Field string
	// Transform is the function applied to the raw field value.
	Transform func(v interface{}) interface{}
}

// New returns a Normalizer pre-loaded with the default rule set.
func New() *Normalizer {
	n := &Normalizer{}
	n.rules = defaultRules()
	return n
}

// AddRule appends a custom normalization rule.
func (n *Normalizer) AddRule(r Rule) {
	n.rules = append(n.rules, r)
}

// Apply returns a shallow-copied spec map with all matching rules applied.
// The original map is never mutated.
func (n *Normalizer) Apply(spec map[string]interface{}) map[string]interface{} {
	out := shallowCopy(spec)
	for _, r := range n.rules {
		parts := strings.SplitN(r.Field, ".", 2)
		top := parts[0]
		if v, ok := out[top]; ok {
			out[top] = r.Transform(v)
		}
	}
	return out
}

// shallowCopy duplicates the top-level keys of m.
func shallowCopy(m map[string]interface{}) map[string]interface{} {
	dup := make(map[string]interface{}, len(m))
	for k, v := range m {
		dup[k] = v
	}
	return dup
}

// defaultRules returns the built-in normalization rules.
func defaultRules() []Rule {
	return []Rule{
		{
			Field: "replicas",
			Transform: func(v interface{}) interface{} {
				switch val := v.(type) {
				case float64:
					if val == 0 {
						return float64(1) // default replica count
					}
				}
				return v
			},
		},
		{
			Field: "tags",
			Transform: func(v interface{}) interface{} {
				slice, ok := v.([]interface{})
				if !ok {
					return v
				}
				strs := make([]string, 0, len(slice))
				for _, s := range slice {
					if sv, ok := s.(string); ok {
						strs = append(strs, strings.TrimSpace(strings.ToLower(sv)))
					}
				}
				sort.Strings(strs)
				out := make([]interface{}, len(strs))
				for i, s := range strs {
					out[i] = s
				}
				return out
			},
		},
	}
}
