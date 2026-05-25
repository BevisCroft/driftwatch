// Package schema provides manifest schema validation for driftwatch.
// It checks that manifests conform to expected structural rules before
// drift detection is performed.
package schema

import (
	"errors"
	"fmt"
	"strings"
)

// Violation describes a single schema rule failure.
type Violation struct {
	Field   string
	Message string
}

func (v Violation) Error() string {
	return fmt.Sprintf("field %q: %s", v.Field, v.Message)
}

// Result holds the outcome of a schema validation pass.
type Result struct {
	Service    string
	Violations []Violation
}

// Valid returns true when no violations were found.
func (r Result) Valid() bool { return len(r.Violations) == 0 }

// Validator validates manifests against a set of rules.
type Validator struct {
	requiredFields []string
	forbiddenKeys  []string
}

// New returns a Validator with sensible defaults.
func New(opts ...Option) *Validator {
	v := &Validator{
		requiredFields: []string{"kind", "name"},
	}
	for _, o := range opts {
		o(v)
	}
	return v
}

// Option configures a Validator.
type Option func(*Validator)

// WithRequiredFields overrides the default required-field list.
func WithRequiredFields(fields ...string) Option {
	return func(v *Validator) { v.requiredFields = fields }
}

// WithForbiddenKeys adds keys that must not appear in the spec.
func WithForbiddenKeys(keys ...string) Option {
	return func(v *Validator) { v.forbiddenKeys = append(v.forbiddenKeys, keys...) }
}

// Validate checks a manifest (represented as a flat map) and returns a Result.
func (v *Validator) Validate(service string, manifest map[string]any) (Result, error) {
	if service == "" {
		return Result{}, errors.New("service name must not be empty")
	}
	res := Result{Service: service}
	for _, f := range v.requiredFields {
		if _, ok := manifest[f]; !ok {
			res.Violations = append(res.Violations, Violation{Field: f, Message: "required field is missing"})
		}
	}
	spec, _ := manifest["spec"].(map[string]any)
	for _, k := range v.forbiddenKeys {
		if _, ok := spec[k]; ok {
			res.Violations = append(res.Violations, Violation{Field: "spec." + k, Message: "forbidden key present"})
		}
	}
	if kind, ok := manifest["kind"].(string); ok && strings.TrimSpace(kind) == "" {
		res.Violations = append(res.Violations, Violation{Field: "kind", Message: "must not be blank"})
	}
	return res, nil
}
