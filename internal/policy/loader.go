package policy

import (
	"fmt"
	"os"

	"github.com/driftwatch/internal/drift"
	"gopkg.in/yaml.v3"
)

// ruleYAML is the on-disk representation of a Rule.
type ruleYAML struct {
	Service  string `yaml:"service"`
	Field    string `yaml:"field"`
	Severity string `yaml:"severity"`
}

// fileSchema is the top-level structure of a policy file.
type fileSchema struct {
	Rules []ruleYAML `yaml:"rules"`
}

// LoadFile reads a YAML policy file and returns the parsed rules.
// Returns an error if the file cannot be read or contains invalid severity values.
func LoadFile(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("policy: read file: %w", err)
	}

	var schema fileSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("policy: parse yaml: %w", err)
	}

	rules := make([]Rule, 0, len(schema.Rules))
	for idx, r := range schema.Rules {
		sev, err := parseSeverity(r.Severity)
		if err != nil {
			return nil, fmt.Errorf("policy: rule[%d]: %w", idx, err)
		}
		rules = append(rules, Rule{
			ServiceGlob: r.Service,
			FieldGlob:   r.Field,
			Severity:    sev,
		})
	}
	return rules, nil
}

func parseSeverity(s string) (drift.Severity, error) {
	switch s {
	case "info":
		return drift.SeverityInfo, nil
	case "warn":
		return drift.SeverityWarn, nil
	case "error":
		return drift.SeverityError, nil
	default:
		return "", fmt.Errorf("unknown severity %q (want info|warn|error)", s)
	}
}
