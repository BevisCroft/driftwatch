// Package window implements a sliding time-window counter used to measure
// how frequently drift events occur for a given service within a rolling
// time period.
//
// The Counter is safe for concurrent use. Events older than the configured
// window are evicted lazily on each Add or Count call, keeping memory
// consumption proportional to the event rate rather than elapsed time.
//
// Typical usage:
//
//	w := window.New(5 * time.Minute)
//	count := w.Add("my-service") // returns total events in last 5 min
//	if count > threshold {
//		// escalate alert
//	}
package window
