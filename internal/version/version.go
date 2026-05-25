// Package version provides build-time version metadata and a lightweight
// HTTP handler that exposes version information for driftwatch deployments.
package version

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// Info holds build-time metadata injected via -ldflags.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// These variables are populated at build time via:
//
//	-ldflags "-X internal/version.version=v1.2.3 -X internal/version.commit=abc123 ..."
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// Get returns the current build Info, filling in runtime fields automatically.
func Get() Info {
	return Info{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// Handler returns an http.Handler that serves version information as JSON.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		info := Get()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Driftwatch-Version", info.Version)
		w.Header().Set("X-Driftwatch-Commit", info.Commit)

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(struct {
			Info      Info      `json:"info"`
			Timestamp time.Time `json:"timestamp"`
		}{
			Info:      info,
			Timestamp: time.Now().UTC(),
		})
	})
}
