package trend

import (
	"fmt"
	"sync"
	"time"

	"portwatch/internal/alerting"
)

// Checker watches for port-binding spikes and emits alerts via an Alerter.
type Checker struct {
	mu        sync.Mutex
	alerter   *alerting.Alerter
	tracker   *Trend
	threshold int
}

// NewChecker creates a Checker that fires a spike alert when more than
// threshold port bindings are observed within window.
func NewChecker(alerter *alerting.Alerter, threshold int, window time.Duration) *Checker {
	return &Checker{
		alerter:   alerter,
		tracker:   New(window),
		threshold: threshold,
	}
}

// Record registers a new port-binding event for the given protocol and port.
// If the spike threshold is exceeded a Warning alert is sent.
func (c *Checker) Record(proto string, port uint16) {
	c.mu.Lock()
	defer c.mu.Unlock()

	spike, count := c.tracker.Record(proto, port)
	if spike {
		a := NewSpikeAlert(proto, count, c.threshold)
		a.Message = fmt.Sprintf(
			"port-binding spike detected on %s: %d new bindings (threshold %d)",
			proto, count, c.threshold,
		)
		c.alerter.Send(a)
	}
}

// Reset clears all recorded trend data, suppressing future alerts until new
// events accumulate again.
func (c *Checker) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tracker.Reset()
}
