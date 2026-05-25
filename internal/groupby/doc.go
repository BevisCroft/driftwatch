// Package groupby provides a Grouper that partitions drift.Result slices
// by a chosen dimension — such as service name, severity level, or
// Kubernetes namespace — to support aggregated reporting and alerting
// workflows.
//
// Built-in keys:
//
//	KeyService   — groups by result.Service
//	KeySeverity  — groups by result.Severity
//	KeyNamespace — groups by the "namespace" field in result.Live,
//	               falling back to "default" when absent
//
// Custom extractors can be registered at runtime via Register, allowing
// callers to group by team ownership, environment, or any other label
// derived from the result payload.
package groupby
