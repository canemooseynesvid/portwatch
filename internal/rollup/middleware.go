package rollup

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/alerting"
)

// AlertMiddleware wraps an alerting.Handler and rolls up repeated alerts
// for the same port/protocol key before forwarding them.
type AlertMiddleware struct {
	next      alerting.Handler
	agg       *Aggregator
	threshold int
	window    time.Duration
}

// NewAlertMiddleware returns an AlertMiddleware that forwards individual alerts
// until threshold is reached, then replaces them with a single rollup alert.
func NewAlertMiddleware(next alerting.Handler, threshold int, window time.Duration) *AlertMiddleware {
	return &AlertMiddleware{
		next:      next,
		agg:       New(threshold, window),
		threshold: threshold,
		window:    window,
	}
}

// Handle satisfies alerting.Handler.
func (m *AlertMiddleware) Handle(a alerting.Alert) {
	key := alertKey(a)
	sum := m.agg.Record(key, a.Message)
	if sum != nil {
		rollupAlert := alerting.Alert{
			Level:     a.Level,
			Message:   sum.String(),
			Timestamp: sum.Last,
			Details:   a.Details,
		}
		m.next.Handle(rollupAlert)
		return
	}
	// Below threshold — pass through unchanged.
	m.next.Handle(a)
}

// alertKey builds a deduplication key from the alert's level and message prefix.
func alertKey(a alerting.Alert) string {
	if port, ok := a.Details["port"]; ok {
		proto, _ := a.Details["protocol"]
		return fmt.Sprintf("%s:%v/%v", a.Level, port, proto)
	}
	// Fallback: use first 40 chars of the message.
	msg := a.Message
	if len(msg) > 40 {
		msg = msg[:40]
	}
	return fmt.Sprintf("%s:%s", a.Level, msg)
}
