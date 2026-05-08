package throttle

import (
	"fmt"
	"time"

	"portwatch/internal/alerting"
)

// AlertThrottler wraps an alerting.Handler and suppresses repeated alerts
// for the same port+protocol key within a configurable burst window.
type AlertThrottler struct {
	next    alerting.Handler
	throttle *Throttle
}

// NewAlertThrottler creates an AlertThrottler that forwards at most burst
// alerts per unique port key within the given window.
func NewAlertThrottler(next alerting.Handler, burst int, window time.Duration) *AlertThrottler {
	return &AlertThrottler{
		next:    next,
		throttle: New(burst, window),
	}
}

// Handle forwards the alert to the next handler only if the key is within
// its burst allowance. Suppressed alerts are silently dropped.
func (t *AlertThrottler) Handle(a alerting.Alert) {
	key := alertKey(a)
	if t.throttle.Allow(key) {
		t.next.Handle(a)
	}
}

// ResetKey clears the throttle state for the given port/protocol key,
// allowing the next alert for that key to pass through immediately.
func (t *AlertThrottler) ResetKey(key string) {
	t.throttle.Reset(key)
}

// Purge removes all expired throttle entries.
func (t *AlertThrottler) Purge() {
	t.throttle.Purge()
}

// alertKey derives a stable string key from an alert's metadata.
func alertKey(a alerting.Alert) string {
	port, _ := a.Meta["port"]
	proto, _ := a.Meta["protocol"]
	if port == "" && proto == "" {
		return a.Message
	}
	return fmt.Sprintf("%s/%s", proto, port)
}
