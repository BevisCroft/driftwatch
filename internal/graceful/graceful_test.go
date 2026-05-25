package graceful

import (
	"context"
	"errors"
	"log"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func silentLogger() *log.Logger {
	return log.New(os.Stderr, "", 0)
}

func TestShutdown_NoHandlers(t *testing.T) {
	m := New(time.Second, silentLogger())
	ctx := context.Background()
	if err := m.Shutdown(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestShutdown_HandlersCalledInReverseOrder(t *testing.T) {
	m := New(time.Second, silentLogger())
	var order []int

	for i := 0; i < 3; i++ {
		i := i
		m.Register(func(_ context.Context) error {
			order = append(order, i)
			return nil
		})
	}

	if err := m.Shutdown(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(order) != 3 || order[0] != 2 || order[1] != 1 || order[2] != 0 {
		t.Errorf("expected LIFO order [2 1 0], got %v", order)
	}
}

func TestShutdown_ReturnsFirstError(t *testing.T) {
	m := New(time.Second, silentLogger())
	errA := errors.New("handler A failed")
	errB := errors.New("handler B failed")

	m.Register(func(_ context.Context) error { return errA })
	m.Register(func(_ context.Context) error { return errB })

	err := m.Shutdown(context.Background())
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	// LIFO: errB is called first, so it is the first error returned.
	if !errors.Is(err, errB) {
		t.Errorf("expected errB as first error, got %v", err)
	}
}

func TestShutdown_ContinuesAfterHandlerError(t *testing.T) {
	m := New(time.Second, silentLogger())
	var called int32

	m.Register(func(_ context.Context) error {
		atomic.AddInt32(&called, 1)
		return errors.New("fail")
	})
	m.Register(func(_ context.Context) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	_ = m.Shutdown(context.Background())
	if atomic.LoadInt32(&called) != 2 {
		t.Errorf("expected both handlers called, got %d", called)
	}
}

func TestShutdown_RespectsContextDeadline(t *testing.T) {
	m := New(50*time.Millisecond, silentLogger())

	m.Register(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return nil
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := m.Shutdown(ctx)
	if err == nil {
		t.Fatal("expected context deadline error, got nil")
	}
}
