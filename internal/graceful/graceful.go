// Package graceful provides utilities for orderly shutdown of long-running
// daemon processes. It listens for OS signals and coordinates the teardown
// of registered components within a configurable deadline.
package graceful

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ShutdownFunc is a function that is called during graceful shutdown.
type ShutdownFunc func(ctx context.Context) error

// Manager coordinates graceful shutdown for registered components.
type Manager struct {
	timeout   time.Duration
	handlers  []ShutdownFunc
	mu        sync.Mutex
	logger    *log.Logger
}

// New creates a Manager with the given shutdown timeout.
func New(timeout time.Duration, logger *log.Logger) *Manager {
	if logger == nil {
		logger = log.New(os.Stderr, "[graceful] ", log.LstdFlags)
	}
	return &Manager{
		timeout: timeout,
		logger:  logger,
	}
}

// Register adds a shutdown handler that will be invoked on shutdown.
// Handlers are called in reverse registration order (LIFO).
func (m *Manager) Register(fn ShutdownFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, fn)
}

// Wait blocks until an OS interrupt or termination signal is received,
// then invokes all registered shutdown handlers within the configured timeout.
// It returns the first non-nil error from any handler, or nil on clean exit.
func (m *Manager) Wait() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	<-quit
	m.logger.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	return m.shutdown(ctx)
}

// Shutdown triggers an immediate graceful shutdown without waiting for a signal.
// Useful in tests or when the calling code controls the shutdown lifecycle.
func (m *Manager) Shutdown(ctx context.Context) error {
	return m.shutdown(ctx)
}

func (m *Manager) shutdown(ctx context.Context) error {
	m.mu.Lock()
	handlers := make([]ShutdownFunc, len(m.handlers))
	copy(handlers, m.handlers)
	m.mu.Unlock()

	var first error
	for i := len(handlers) - 1; i >= 0; i-- {
		if err := handlers[i](ctx); err != nil {
			m.logger.Printf("shutdown handler error: %v", err)
			if first == nil {
				first = err
			}
		}
	}
	return first
}
