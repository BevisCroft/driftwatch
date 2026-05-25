// Package watchlist maintains the authoritative set of services that
// driftwatch should actively monitor for configuration drift.
//
// # Overview
//
// A Watchlist is a thread-safe registry of Entry values keyed by service
// name. Each entry carries optional namespace and label metadata that
// downstream components (filter, scheduler, alerting) can consult to
// decide how to handle a particular service.
//
// # HTTP API
//
// Handler exposes three REST endpoints:
//
//	GET    /watchlist        – return all registered entries as JSON
//	POST   /watchlist        – register a new entry (JSON body)
//	DELETE /watchlist/{name} – remove the named service
package watchlist
