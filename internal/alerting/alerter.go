// Package alerting provides notification support for detected configuration drift.
package alerting

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Level represents the severity of a drift alert.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert holds a formatted drift notification.
type Alert struct {
	Service   string
	Level     Level
	Message   string
	Timestamp time.Time
}

// Alerter sends alerts for drift results.
type Alerter struct {
	out       io.Writer
	threshold int // minimum number of drifted fields to escalate to ERROR
}

// New creates an Alerter writing to the given writer.
// If w is nil, os.Stderr is used.
func New(w io.Writer, threshold int) *Alerter {
	if w == nil {
		w = os.Stderr
	}
	if threshold <= 0 {
		threshold = 3
	}
	return &Alerter{out: w, threshold: threshold}
}

// Notify emits alerts for any results that contain drift.
func (a *Alerter) Notify(results []drift.Result) []Alert {
	var alerts []Alert
	for _, r := range results {
		if !r.HasDrift {
			continue
		}
		lvl := LevelWarn
		if len(r.Differences) >= a.threshold {
			lvl = LevelError
		}
		msg := buildMessage(r)
		al := Alert{
			Service:   r.Service,
			Level:     lvl,
			Message:   msg,
			Timestamp: time.Now().UTC(),
		}
		alerts = append(alerts, al)
		fmt.Fprintf(a.out, "[%s] %s: %s\n", al.Level, al.Service, al.Message)
	}
	return alerts
}

func buildMessage(r drift.Result) string {
	parts := make([]string, 0, len(r.Differences))
	for _, d := range r.Differences {
		parts = append(parts, fmt.Sprintf("%s (want %q, got %q)", d.Field, d.Expected, d.Actual))
	}
	return fmt.Sprintf("%d field(s) drifted: %s", len(r.Differences), strings.Join(parts, "; "))
}
