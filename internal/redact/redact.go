// Package redact provides utilities for scrubbing sensitive fields from
// alert metadata before the alert is forwarded to external channels.
package redact

import (
	"strings"

	"portwatch/internal/alerting"
)

// Field names that are considered sensitive and will be replaced.
var defaultSensitiveKeys = []string{
	"token",
	"secret",
	"password",
	"api_key",
	"apikey",
	"auth",
	"credential",
}

const redactedValue = "[REDACTED]"

// Redactor scrubs sensitive metadata fields from alerts.
type Redactor struct {
	keys []string
}

// New returns a Redactor that masks the default set of sensitive keys.
func New() *Redactor {
	return &Redactor{keys: defaultSensitiveKeys}
}

// NewWithKeys returns a Redactor that masks the provided keys (case-insensitive)
// in addition to the defaults.
func NewWithKeys(extra ...string) *Redactor {
	merged := make([]string, len(defaultSensitiveKeys)+len(extra))
	copy(merged, defaultSensitiveKeys)
	for i, k := range extra {
		merged[len(defaultSensitiveKeys)+i] = strings.ToLower(k)
	}
	return &Redactor{keys: merged}
}

// Apply returns a copy of the alert with sensitive metadata fields replaced.
// The original alert is never mutated.
func (r *Redactor) Apply(a alerting.Alert) alerting.Alert {
	if len(a.Meta) == 0 {
		return a
	}
	clean := make(map[string]string, len(a.Meta))
	for k, v := range a.Meta {
		if r.isSensitive(k) {
			clean[k] = redactedValue
		} else {
			clean[k] = v
		}
	}
	a.Meta = clean
	return a
}

// Middleware returns an alerting.Handler that redacts each alert before
// forwarding it to next.
func (r *Redactor) Middleware(next alerting.Handler) alerting.Handler {
	return alerting.HandlerFunc(func(a alerting.Alert) error {
		return next.Handle(r.Apply(a))
	})
}

func (r *Redactor) isSensitive(key string) bool {
	lower := strings.ToLower(key)
	for _, k := range r.keys {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}
