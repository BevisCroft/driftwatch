// Package diff provides structured field-level diffing between two manifest
// spec maps, producing human-readable and machine-readable change records.
package diff

import (
	"fmt"
	"sort"
)

// ChangeKind describes the type of change detected for a field.
type ChangeKind string

const (
	Added    ChangeKind = "added"
	Removed  ChangeKind = "removed"
	Modified ChangeKind = "modified"
)

// Change represents a single field-level difference between two specs.
type Change struct {
	Field    string
	Kind     ChangeKind
	OldValue any
	NewValue any
}

// String returns a human-readable representation of the change.
func (c Change) String() string {
	switch c.Kind {
	case Added:
		return fmt.Sprintf("%s: added %v", c.Field, c.NewValue)
	case Removed:
		return fmt.Sprintf("%s: removed (was %v)", c.Field, c.OldValue)
	case Modified:
		return fmt.Sprintf("%s: %v -> %v", c.Field, c.OldValue, c.NewValue)
	}
	return c.Field
}

// Compute returns the ordered list of field-level changes between baseline
// and current spec maps. Only top-level keys are compared.
func Compute(baseline, current map[string]any) []Change {
	var changes []Change

	for key, bVal := range baseline {
		cVal, ok := current[key]
		if !ok {
			changes = append(changes, Change{Field: key, Kind: Removed, OldValue: bVal})
			continue
		}
		if fmt.Sprintf("%v", bVal) != fmt.Sprintf("%v", cVal) {
			changes = append(changes, Change{Field: key, Kind: Modified, OldValue: bVal, NewValue: cVal})
		}
	}

	for key, cVal := range current {
		if _, ok := baseline[key]; !ok {
			changes = append(changes, Change{Field: key, Kind: Added, NewValue: cVal})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Field < changes[j].Field
	})
	return changes
}
