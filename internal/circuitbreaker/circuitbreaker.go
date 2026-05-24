// Package circuitbreaker provides a simple circuit breaker for protecting
// downstream notification and alerting calls from cascading failures.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the current circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// String returns a human-readable state label.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrOpen is returned when a call is rejected because the circuit is open.
var ErrOpen = errors.New("circuit breaker is open")

// Breaker tracks consecutive failures for a named service and opens the
// circuit after a configurable threshold, allowing retries after a cooldown.
type Breaker struct {
	mu           sync.Mutex
	threshold    int
	cooldown     time.Duration
	now          func() time.Time
	circuits     map[string]*circuit
}

type circuit struct {
	state      State
	failures   int
	openedAt   time.Time
}

// New creates a Breaker. threshold is the number of consecutive failures
// before opening; cooldown is how long to wait before entering half-open.
func New(threshold int, cooldown time.Duration) *Breaker {
	return &Breaker{
		threshold: threshold,
		cooldown:  cooldown,
		now:       time.Now,
		circuits:  make(map[string]*circuit),
	}
}

// Allow returns nil if the call should proceed, or ErrOpen if it is blocked.
func (b *Breaker) Allow(service string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	c := b.getOrCreate(service)
	switch c.state {
	case StateOpen:
		if b.now().Sub(c.openedAt) >= b.cooldown {
			c.state = StateHalfOpen
			return nil
		}
		return ErrOpen
	default:
		return nil
	}
}

// RecordSuccess resets the failure count and closes the circuit.
func (b *Breaker) RecordSuccess(service string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	c := b.getOrCreate(service)
	c.failures = 0
	c.state = StateClosed
}

// RecordFailure increments the failure counter and opens the circuit if the
// threshold is reached.
func (b *Breaker) RecordFailure(service string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	c := b.getOrCreate(service)
	c.failures++
	if c.failures >= b.threshold {
		c.state = StateOpen
		c.openedAt = b.now()
	}
}

// StateOf returns the current state for a service.
func (b *Breaker) StateOf(service string) State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.getOrCreate(service).state
}

func (b *Breaker) getOrCreate(service string) *circuit {
	if c, ok := b.circuits[service]; ok {
		return c
	}
	c := &circuit{state: StateClosed}
	b.circuits[service] = c
	return c
}
