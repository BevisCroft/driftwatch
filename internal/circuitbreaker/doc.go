// Package circuitbreaker implements a per-service circuit breaker pattern for
// driftwatch's outbound notification and alerting subsystems.
//
// A Breaker tracks consecutive call failures for each named service. Once the
// failure count reaches the configured threshold the circuit transitions to the
// open state and subsequent Allow calls return ErrOpen immediately, preventing
// further calls to the unhealthy downstream target.
//
// After the cooldown duration has elapsed the circuit enters the half-open
// state, allowing a single probe call through. A successful probe (signalled
// via RecordSuccess) closes the circuit; another failure re-opens it.
//
// Usage:
//
//	br := circuitbreaker.New(5, 30*time.Second)
//
//	if err := br.Allow("slack"); err != nil {
//		// skip notification
//		return err
//	}
//	if err := sendSlackAlert(msg); err != nil {
//		br.RecordFailure("slack")
//		return err
//	}
//	br.RecordSuccess("slack")
package circuitbreaker
