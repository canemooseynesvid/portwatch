package trend

import (
	"fmt"

	"portwatch/internal/alerting"
)

// NewSpikeAlert constructs an alerting.Alert for a detected binding-rate spike.
func NewSpikeAlert(s Spike) alerting.Alert {
	msg := fmt.Sprintf(
		"binding spike detected: %d new ports opened within %v (threshold exceeded)",
		s.Count, s.Window,
	)
	return alerting.NewAlert(alerting.LevelWarn, "trend/spike", msg,
		map[string]string{
			"key":    s.Key,
			"count":  fmt.Sprintf("%d", s.Count),
			"window": s.Window.String(),
		},
	)
}

// Checker wraps a Tracker and emits spike alerts via an Alerter.
type Checker struct {
	tracker *Tracker
	alerter *alerting.Alerter
}

// NewChecker creates a Checker that records events and forwards spikes.
func NewChecker(tracker *Tracker, alerter *alerting.Alerter) *Checker {
	return &Checker{tracker: tracker, alerter: alerter}
}

// Observe records a binding event and sends a spike alert if the threshold
// is exceeded. It returns the Spike if one was generated, otherwise nil.
func (c *Checker) Observe(e Event) *Spike {
	spike := c.tracker.Record(e)
	if spike != nil {
		c.alerter.Send(NewSpikeAlert(*spike))
	}
	return spike
}
