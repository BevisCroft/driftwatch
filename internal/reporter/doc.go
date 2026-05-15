// Package reporter provides formatting and output capabilities for drift
// detection results produced by the drift package.
//
// It supports multiple output formats:
//
//   - text: human-readable summary suitable for terminal output
//   - json: machine-readable JSON for integration with external tooling
//
// Usage:
//
//	r := reporter.New(os.Stdout, reporter.FormatText)
//	if err := r.Write(results); err != nil {
//	    log.Fatal(err)
//	}
package reporter
