package schema

import (
	"portwatch/internal/alerting"
	"portwatch/internal/portscanner"
)

// Checker integrates the Validator into the alerting pipeline by scanning a
// slice of entries and forwarding any violations to an alerting.Alerter.
type Checker struct {
	v       *Validator
	alerter *alerting.Alerter
}

// NewChecker creates a Checker using the provided validator and alerter.
func NewChecker(v *Validator, a *alerting.Alerter) *Checker {
	return &Checker{v: v, alerter: a}
}

// Check validates every entry in entries and sends alerts for each violation.
// It returns the total number of violations found.
func (c *Checker) Check(entries []portscanner.Entry) int {
	total := 0
	for _, e := range entries {
		for _, a := range c.v.Validate(e) {
			c.alerter.Send(a)
			total++
		}
	}
	return total
}
