package window

import (
	"fmt"
	"time"

	"github.com/example/portwatch/internal/alerting"
)

// FrequencyChecker wraps a Counter and emits an alert when the event
// frequency for a port exceeds a configured threshold within the window.
type FrequencyChecker struct {
	counter   *Counter
	threshold int
	alerter   *alerting.Alerter
}

// NewFrequencyChecker returns a FrequencyChecker that fires an alert when
// more than threshold events are recorded for a key within window.
func NewFrequencyChecker(window time.Duration, threshold int, a *alerting.Alerter) *FrequencyChecker {
	return &FrequencyChecker{
		counter:   New(window),
		threshold: threshold,
		alerter:   a,
	}
}

// Record increments the counter for the given port/protocol key. When the
// count exceeds the threshold an alerting.Alert is sent via the Alerter.
func (f *FrequencyChecker) Record(protocol string, port uint16) {
	key := fmt.Sprintf("%s:%d", protocol, port)
	count := f.counter.Record(key)

	if count > f.threshold {
		msg := fmt.Sprintf(
			"port %s/%d exceeded frequency threshold: %d events in window",
			protocol, port, count,
		)
		alert := alerting.Alert{
			Level:   alerting.LevelWarning,
			Message: msg,
			Fields: map[string]string{
				"protocol": protocol,
				"port":     fmt.Sprintf("%d", port),
				"count":    fmt.Sprintf("%d", count),
			},
		}
		f.alerter.Send(alert)
	}
}

// Reset clears the counter for the given key, e.g. after a port closes.
func (f *FrequencyChecker) Reset(protocol string, port uint16) {
	f.counter.Reset(fmt.Sprintf("%s:%d", protocol, port))
}
