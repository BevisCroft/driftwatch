// Package tagindex implements an in-memory inverted index that maps
// arbitrary key-value tags to the set of services that carry them.
//
// It is designed to support fast multi-tag lookups across large numbers
// of deployed service manifests, enabling downstream components such as
// the filter, policy, and rollup modules to query services by metadata
// without scanning every manifest on each cycle.
//
// All operations are safe for concurrent use.
package tagindex
