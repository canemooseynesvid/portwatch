package baseline

import (
	"fmt"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/portscanner"
)

// Checker compares live port entries against a known baseline and emits alerts
// for bindings that are not recognised.
type Checker struct {
	baseline *Baseline
	alerter  *alerting.Alerter
}

// NewChecker creates a Checker backed by the given Baseline and Alerter.
func NewChecker(b *Baseline, a *alerting.Alerter) *Checker {
	return &Checker{baseline: b, alerter: a}
}

// Check inspects a single port entry and emits a warning alert when the binding
// is not present in the baseline.
func (c *Checker) Check(e portscanner.Entry) {
	if c.baseline.Contains(e.Protocol, e.LocalAddr, e.LocalPort) {
		return
	}
	alert := alerting.Alert{
		Level:   alerting.Warning,
		Message: fmt.Sprintf("unexpected binding: %s %s:%d is not in the baseline", e.Protocol, e.LocalAddr, e.LocalPort),
		Details: map[string]any{
			"protocol": e.Protocol,
			"address":  e.LocalAddr,
			"port":     e.LocalPort,
		},
	}
	c.alerter.Send(alert)
}

// Learn adds the given entry to the baseline, making future occurrences
// acceptable.
func (c *Checker) Learn(e portscanner.Entry) error {
	return c.baseline.Add(Entry{
		Protocol: e.Protocol,
		Address:  e.LocalAddr,
		Port:     e.LocalPort,
	})
}

// Forget removes the given entry from the baseline.
func (c *Checker) Forget(protocol, address string, port uint16) error {
	return c.baseline.Remove(protocol, address, port)
}
