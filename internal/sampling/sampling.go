// Package sampling provides probabilistic and rate-based sampling
// for drift detection results, reducing noise in high-volume environments.
package sampling

import (
	"math/rand"
	"sync"
	"time"
)

// Strategy defines how samples are selected.
type Strategy string

const (
	StrategyRandom     Strategy = "random"
	StrategyDeterministic Strategy = "deterministic"
)

// Config holds sampler configuration.
type Config struct {
	// Rate is the fraction of events to allow through [0.0, 1.0].
	Rate     float64
	Strategy Strategy
}

// Sampler decides whether a given service's drift event should be processed.
type Sampler struct {
	mu      sync.Mutex
	cfg     Config
	rng     *rand.Rand
	counters map[string]uint64
}

// New creates a Sampler with the given Config.
// Rate is clamped to [0.0, 1.0].
func New(cfg Config) *Sampler {
	if cfg.Rate < 0 {
		cfg.Rate = 0
	}
	if cfg.Rate > 1 {
		cfg.Rate = 1
	}
	if cfg.Strategy == "" {
		cfg.Strategy = StrategyRandom
	}
	return &Sampler{
		cfg:      cfg,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		counters: make(map[string]uint64),
	}
}

// Allow returns true if the event for the given service should be processed.
func (s *Sampler) Allow(service string) bool {
	if s.cfg.Rate == 0 {
		return false
	}
	if s.cfg.Rate == 1 {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	switch s.cfg.Strategy {
	case StrategyDeterministic:
		return s.allowDeterministic(service)
	default:
		return s.rng.Float64() < s.cfg.Rate
	}
}

// allowDeterministic uses a per-service counter to deterministically
// sample every N-th event according to the configured rate.
func (s *Sampler) allowDeterministic(service string) bool {
	s.counters[service]++
	n := s.counters[service]
	period := uint64(1.0 / s.cfg.Rate)
	if period < 1 {
		period = 1
	}
	return n%period == 1
}

// Reset clears per-service counters (useful in tests).
func (s *Sampler) Reset(service string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.counters, service)
}

// Rate returns the configured sampling rate.
func (s *Sampler) Rate() float64 {
	return s.cfg.Rate
}
