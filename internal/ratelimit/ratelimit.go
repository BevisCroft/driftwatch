// Package ratelimit provides a simple token-bucket rate limiter for
// controlling how frequently drift alerts and notifications are emitted
// per service, preventing alert storms during repeated drift cycles.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks per-service alert rates using a token-bucket approach.
type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     time.Duration // minimum duration between allowed events
	maxBurst int
}

type bucket struct {
	tokens    int
	lastRefil time.Time
}

// New creates a Limiter that allows at most maxBurst events per service
// and refills one token every rate duration.
func New(rate time.Duration, maxBurst int) *Limiter {
	if maxBurst < 1 {
		maxBurst = 1
	}
	return &Limiter{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		maxBurst: maxBurst,
	}
}

// Allow reports whether an event for the given service key is permitted.
// It refills tokens based on elapsed time and consumes one token if available.
func (l *Limiter) Allow(service string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[service]
	if !ok {
		l.buckets[service] = &bucket{tokens: l.maxBurst - 1, lastRefil: now}
		return true
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(b.lastRefil)
	newTokens := int(elapsed / l.rate)
	if newTokens > 0 {
		b.tokens += newTokens
		if b.tokens > l.maxBurst {
			b.tokens = l.maxBurst
		}
		b.lastRefil = b.lastRefil.Add(time.Duration(newTokens) * l.rate)
	}

	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// Reset clears the bucket for the given service, restoring full burst capacity.
func (l *Limiter) Reset(service string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, service)
}
