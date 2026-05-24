// Package backoff implements exponential backoff with optional jitter.
//
// It is used by driftwatch components that communicate with external
// endpoints (alerting webhooks, notifiers) to avoid thundering-herd
// behaviour when those endpoints are temporarily unavailable.
//
// Usage:
//
//	b := backoff.New(backoff.Strategy{
//		InitialInterval: time.Second,
//		MaxInterval:     5 * time.Minute,
//		Multiplier:      2.0,
//		JitterFraction:  0.3,
//	})
//
//	for {
//		err := notify()
//		if err == nil {
//			b.Reset(serviceKey)
//			break
//		}
//		time.Sleep(b.Next(serviceKey))
//	}
package backoff
