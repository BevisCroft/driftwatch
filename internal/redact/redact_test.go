package redact_test

import (
	"testing"

	"github.com/example/driftwatch/internal/redact"
)

func TestIsSensitive_MatchesSubstring(t *testing.T) {
	r := redact.New([]string{"password", "secret"})
	if !r.IsSensitive("spec.env.DB_PASSWORD") {
		t.Error("expected DB_PASSWORD to be sensitive")
	}
	if !r.IsSensitive("spec.SECRET_KEY") {
		t.Error("expected SECRET_KEY to be sensitive")
	}
}

func TestIsSensitive_CaseInsensitive(t *testing.T) {
	r := redact.New([]string{"token"})
	if !r.IsSensitive("spec.AUTH_TOKEN") {
		t.Error("expected AUTH_TOKEN to match 'token' pattern")
	}
}

func TestIsSensitive_NoMatch(t *testing.T) {
	r := redact.New([]string{"password"})
	if r.IsSensitive("spec.replicas") {
		t.Error("spec.replicas should not be sensitive")
	}
}

func TestIsSensitive_EmptyPatterns(t *testing.T) {
	r := redact.New(nil)
	if r.IsSensitive("spec.env.SECRET") {
		t.Error("no patterns registered; nothing should be sensitive")
	}
}

func TestAddPattern_AppliedAtRuntime(t *testing.T) {
	r := redact.New([]string{"password"})
	if r.IsSensitive("spec.api_key") {
		t.Fatal("api_key should not be sensitive before AddPattern")
	}
	r.AddPattern("api_key")
	if !r.IsSensitive("spec.api_key") {
		t.Error("api_key should be sensitive after AddPattern")
	}
}

func TestScrubMap_RedactsSensitiveKeys(t *testing.T) {
	r := redact.New([]string{"secret", "password"})
	input := map[string]string{
		"spec.replicas":       "3",
		"spec.env.DB_PASSWORD": "hunter2",
		"spec.env.LOG_LEVEL":  "info",
		"spec.secret_token":   "abc123",
	}
	out := r.ScrubMap(input)

	if out["spec.replicas"] != "3" {
		t.Errorf("replicas should be unchanged, got %q", out["spec.replicas"])
	}
	if out["spec.env.LOG_LEVEL"] != "info" {
		t.Errorf("LOG_LEVEL should be unchanged, got %q", out["spec.env.LOG_LEVEL"])
	}
	if out["spec.env.DB_PASSWORD"] != "[REDACTED]" {
		t.Errorf("DB_PASSWORD should be redacted, got %q", out["spec.env.DB_PASSWORD"])
	}
	if out["spec.secret_token"] != "[REDACTED]" {
		t.Errorf("secret_token should be redacted, got %q", out["spec.secret_token"])
	}
}

func TestScrubValue_SensitiveField(t *testing.T) {
	r := redact.New([]string{"token"})
	got := r.ScrubValue("spec.auth_token", "supersecret")
	if got != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %q", got)
	}
}

func TestScrubValue_NonSensitiveField(t *testing.T) {
	r := redact.New([]string{"token"})
	got := r.ScrubValue("spec.image", "nginx:1.25")
	if got != "nginx:1.25" {
		t.Errorf("expected value unchanged, got %q", got)
	}
}
