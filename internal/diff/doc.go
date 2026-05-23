// Package diff provides field-level diffing utilities for comparing manifest
// spec maps between their baseline and currently deployed state.
//
// # Overview
//
// Compute produces an ordered slice of Change values, each describing whether
// a top-level spec field was added, removed, or modified between two snapshots.
//
// # Rendering
//
// RenderText and RenderMarkdown write formatted summaries to any io.Writer,
// suitable for log output or notification payloads respectively.
package diff
