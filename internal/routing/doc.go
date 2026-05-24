// Package routing implements weighted round-robin selection across multiple
// manifest source endpoints.
//
// A Router is constructed with a slice of Endpoint values, each carrying a
// name, URL, and positive integer weight. Calls to Next advance through the
// endpoint list, honouring relative weights so that higher-weight endpoints
// receive proportionally more selections.
//
// The Router is safe for concurrent use. HTTP introspection endpoints are
// exposed via Handler, which mounts read-only routes under /routing/.
package routing
