// Package schema validates port scanner entries against expected structural
// constraints, emitting alerts when entries deviate from known-good shapes.
package schema

import (
	"fmt"

	"portwatch/internal/alerting"
	"portwatch/internal/portscanner"
)

// Rule describes a single structural constraint applied to a port entry.
type Rule struct {
	Name    string
	Check   func(e portscanner.Entry) bool
	Message string
	Level   alerting.Level
}

// Validator holds a set of rules and evaluates entries against them.
type Validator struct {
	rules []Rule
}

// DefaultRules returns the built-in set of structural rules.
func DefaultRules() []Rule {
	return []Rule{
		{
			Name:    "non-zero-port",
			Check:   func(e portscanner.Entry) bool { return e.Port > 0 },
			Message: "entry has zero or negative port number",
			Level:   alerting.LevelWarning,
		},
		{
			Name:    "known-protocol",
			Check:   func(e portscanner.Entry) bool { return e.Protocol == "tcp" || e.Protocol == "udp" },
			Message: "entry has unrecognised protocol",
			Level:   alerting.LevelWarning,
		},
		{
			Name:    "non-empty-addr",
			Check:   func(e portscanner.Entry) bool { return e.Addr != "" },
			Message: "entry has empty local address",
			Level:   alerting.LevelInfo,
		},
	}
}

// New creates a Validator with the supplied rules.
func New(rules []Rule) *Validator {
	copy := make([]Rule, len(rules))
	for i, r := range rules {
		copy[i] = r
	}
	return &Validator{rules: copy}
}

// Validate checks e against every rule and returns one alert per violation.
func (v *Validator) Validate(e portscanner.Entry) []alerting.Alert {
	var alerts []alerting.Alert
	for _, r := range v.rules {
		if !r.Check(e) {
			a := alerting.NewPortBindAlert(e)
			a.Level = r.Level
			a.Message = fmt.Sprintf("schema violation [%s]: %s (port %d/%s)",
				r.Name, r.Message, e.Port, e.Protocol)
			alerts = append(alerts, a)
		}
	}
	return alerts
}

// RuleCount returns the number of rules registered in the validator.
func (v *Validator) RuleCount() int { return len(v.rules) }
