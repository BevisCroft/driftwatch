// Package audit provides an append-only audit log for drift detection events.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Event     string    `json:"event"`
	Details   string    `json:"details,omitempty"`
}

// Logger writes audit entries to a file in newline-delimited JSON.
type Logger struct {
	mu   sync.Mutex
	file *os.File
}

// New opens (or creates) the audit log file at path.
func New(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("audit: open log file: %w", err)
	}
	return &Logger{file: f}, nil
}

// Record appends an entry to the audit log.
func (l *Logger) Record(service, event, details string) error {
	entry := Entry{
		Timestamp: time.Now().UTC(),
		Service:   service,
		Event:     event,
		Details:   details,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, err = fmt.Fprintf(l.file, "%s\n", data)
	if err != nil {
		return fmt.Errorf("audit: write entry: %w", err)
	}
	return nil
}

// Close closes the underlying log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}
