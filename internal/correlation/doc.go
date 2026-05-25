// Package correlation identifies co-occurring drift events across services.
//
// A Tracker accumulates drift results as they are produced by the scheduler
// and surfaces pairs of services that drifted on the same manifest field
// within a configurable time window. This helps operators distinguish
// isolated incidents from systemic rollout problems or shared-config issues.
//
// Usage:
//
//	tr := correlation.New(5 * time.Minute)
//	tr.Record("payments", driftResults)
//	tr.Record("orders", driftResults)
//	for _, m := range tr.Correlate() {
//		log.Printf("%s and %s both drifted on %s", m.ServiceA, m.ServiceB, m.Field)
//	}
package correlation
