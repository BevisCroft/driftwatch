// Package notifier provides webhook-based notification delivery for drift events.
// It supports configurable HTTP endpoints with retry logic and payload templating.
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/driftwatch/internal/drift"
)

// WebhookConfig holds configuration for a single webhook endpoint.
type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Timeout time.Duration     `yaml:"timeout"`
	Retries int               `yaml:"retries"`
}

// Payload is the JSON body sent to a webhook endpoint.
type Payload struct {
	Timestamp string         `json:"timestamp"`
	Service   string         `json:"service"`
	Drifted   bool           `json:"drifted"`
	Changes   []drift.Change `json:"changes,omitempty"`
}

// Notifier dispatches drift results to configured webhook endpoints.
type Notifier struct {
	cfg    []WebhookConfig
	client *http.Client
	log    *slog.Logger
}

// New creates a Notifier with the given webhook configurations.
// A nil logger falls back to the default slog logger.
func New(cfg []WebhookConfig, log *slog.Logger) *Notifier {
	if log == nil {
		log = slog.Default()
	}
	return &Notifier{
		cfg:    cfg,
		client: &http.Client{},
		log:    log,
	}
}

// Notify sends the drift result to all configured webhook endpoints.
// Errors from individual webhooks are logged but do not abort delivery to
// remaining endpoints. The first error encountered is returned to the caller.
func (n *Notifier) Notify(ctx context.Context, result drift.Result) error {
	payload := Payload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   result.Service,
		Drifted:   result.Drifted,
		Changes:   result.Changes,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notifier: marshal payload: %w", err)
	}

	var firstErr error
	for _, wh := range n.cfg {
		if err := n.send(ctx, wh, body); err != nil {
			n.log.Error("notifier: webhook delivery failed",
				"url", wh.URL, "service", result.Service, "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// send posts body to a single webhook, retrying up to cfg.Retries times on
// non-2xx responses or transient network errors.
func (n *Notifier) send(ctx context.Context, cfg WebhookConfig, body []byte) error {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	retries := cfg.Retries
	if retries < 0 {
		retries = 0
	}

	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		reqCtx, cancel := context.WithTimeout(ctx, timeout)
		err := n.doRequest(reqCtx, cfg, body)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		n.log.Warn("notifier: retrying webhook",
			"url", cfg.URL, "attempt", attempt+1, "error", err)
	}
	return lastErr
}

// doRequest executes a single HTTP POST to the webhook URL.
func (n *Notifier) doRequest(ctx context.Context, cfg WebhookConfig, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, cfg.URL)
	}
	return nil
}
