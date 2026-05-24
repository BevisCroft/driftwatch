// Package export provides functionality for exporting drift results to
// external formats and destinations such as JSON files, CSV, and HTTP endpoints.
package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Format represents the output format for exported drift results.
type Format string

const (
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
)

// Destination describes where exported data should be written.
type Destination struct {
	// Format is the serialisation format (json or csv).
	Format Format

	// FilePath, if non-empty, writes output to the named file.
	FilePath string

	// Endpoint, if non-empty, POSTs the payload to the URL.
	Endpoint string

	// Headers are additional HTTP headers sent with endpoint requests.
	Headers map[string]string
}

// Exporter writes drift results to one or more destinations.
type Exporter struct {
	dest   Destination
	client *http.Client
}

// New creates an Exporter for the given Destination.
func New(dest Destination) *Exporter {
	return &Exporter{
		dest: dest,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Export serialises results and writes them to the configured destination.
// At least one of FilePath or Endpoint must be set on the Destination.
func (e *Exporter) Export(ctx context.Context, results []drift.Result) error {
	if len(results) == 0 {
		return nil
	}

	var buf bytes.Buffer
	switch e.dest.Format {
	case FormatCSV:
		if err := writeCSV(&buf, results); err != nil {
			return fmt.Errorf("export: csv encode: %w", err)
		}
	default: // FormatJSON
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			return fmt.Errorf("export: json encode: %w", err)
		}
	}

	if e.dest.FilePath != "" {
		if err := writeFile(e.dest.FilePath, buf.Bytes()); err != nil {
			return err
		}
	}

	if e.dest.Endpoint != "" {
		if err := e.post(ctx, buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

// writeFile atomically writes data to path by using a temp file and rename.
func writeFile(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("export: write file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("export: rename file: %w", err)
	}
	return nil
}

// post sends data to the configured HTTP endpoint.
func (e *Exporter) post(ctx context.Context, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.dest.Endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("export: build request: %w", err)
	}

	contentType := "application/json"
	if e.dest.Format == FormatCSV {
		contentType = "text/csv"
	}
	req.Header.Set("Content-Type", contentType)

	for k, v := range e.dest.Headers {
		req.Header.Set(k, v)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("export: http post: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("export: endpoint returned %d", resp.StatusCode)
	}
	return nil
}

// writeCSV encodes results as CSV rows into w.
func writeCSV(w io.Writer, results []drift.Result) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"service", "field", "expected", "actual", "severity"}); err != nil {
		return err
	}
	for _, r := range results {
		for _, d := range r.Diffs {
			row := []string{
				r.Service,
				d.Field,
				fmt.Sprintf("%v", d.Expected),
				fmt.Sprintf("%v", d.Actual),
				r.Severity,
			}
			if err := cw.Write(row); err != nil {
				return err
			}
		}
	}
	cw.Flush()
	return cw.Error()
}
