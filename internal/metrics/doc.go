// Package metrics exposes operational counters and gauges for the
// driftwatch daemon.
//
// Metrics are collected in-process (no external dependencies) and
// served over HTTP via Handler so they can be scraped by Prometheus
// or any compatible collector.
//
// All exported methods are safe for concurrent use.
package metrics
