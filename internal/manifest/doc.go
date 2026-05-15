// Package manifest provides utilities for loading and representing
// service manifest files used by driftwatch to detect configuration drift.
//
// A Manifest is a structured YAML document that describes the desired state
// of a deployed service — including its kind, metadata, and spec fields.
//
// Basic usage:
//
//	loader := manifest.NewLoader("/etc/driftwatch/manifests")
//
//	// Load a single manifest by filename:
//	m, err := loader.Load("api-service.yaml")
//
//	// Load all manifests in the directory:
//	all, err := loader.LoadAll()
//
// The loaded Manifest values are subsequently passed to the drift-detection
// engine, which compares them against the live state of running services.
package manifest
