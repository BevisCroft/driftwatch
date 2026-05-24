// Package backoff provides exponential backoff with jitter for retry logic
// used when alerting or notification endpoints are temporarily unavailable.
package backoff

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// Strategy defines how retries are spaced over time.
type Strategy struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	JitterFraction  float64 // 0.0 = no jitter, 1.0 = full jitter
}

// Backoff tracks per-key retry state.
type Backoff struct {
	strategy Strategy
	mu       sync.Mutex
	attempts map[string]int
	now      func() time.Time
}

// New creates a Backoff with the given strategy.
func New(s Strategy) *Backoff {
	if s.Multiplier <= 1.0 {
		s.Multiplier = 2.0
	}
	if s.MaxInterval <= 0 {
		s.MaxInterval = 5 * time.Minute
	}
	if s.InitialInterval <= 0 {
		s.InitialInterval = time.Second
	}
	return &Backoff{
		strategy: s,
		attempts: make(map[string]int),
		now:      time.Now,
	}
}

// Next returns the duration to wait before the next retry for the given key
// and increments the attempt counter.
func (b *Backoff) Next(key string) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := b.attempts[key]
	b.attempts[key] = n + 1

	base := float64(b.strategy.InitialInterval) * math.Pow(b.strategy.Multiplier, float64(n))
	if base > float64(b.strategy.MaxInterval) {
		base = float64(b.strategy.MaxInterval)
	}

	jitter := 0.0
	if b.strategy.JitterFraction > 0 {
		jitter = rand.Float64() * b.strategy.JitterFraction * base
	}

	return time.Duration(base + jitter)
}

// Reset clears the attempt counter for the given key.
func (b *Backoff) Reset(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.attempts, key)
}

// Attempts returns the current attempt count for the given key.
func (b *Backoff) Attempts(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.attempts[key]
}
