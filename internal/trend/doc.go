// Package trend tracks how frequently individual services experience
// configuration drift over a sliding observation window.
//
// A Tracker accumulates drift observations (service name + field path)
// and can produce ranked summaries showing which services are drifting
// most often. Observations older than the configured window are ignored
// when computing summaries and can be evicted explicitly via Purge.
//
// Typical usage:
//
//	tr := trend.New(30 * time.Minute)
//
//	// record each drift result from the detector
//	for _, r := range results {
//		for _, d := range r.Diffs {
//			tr.Record(r.Service, d.Field)
//		}
//	}
//
//	// inspect the top drifting services
//	for _, s := range tr.Summaries() {
//		fmt.Printf("%s drifted %d times\n", s.Service, s.Count)
//	}
package trend
