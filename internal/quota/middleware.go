package quota

import (
	"fmt"

	"portwatch/internal/alerting"
)

// AlertMiddleware wraps an alerting.Handler and enforces per-alert-key quotas.
// Alerts that exceed the configured quota are silently dropped.
type AlertMiddleware struct {
	quota *Quota
	next  alerting.Handler
}

// NewAlertMiddleware returns an AlertMiddleware that limits each unique alert
// key to at most q.max emissions within q.window before forwarding to next.
func NewAlertMiddleware(q *Quota, next alerting.Handler) *AlertMiddleware {
	return &AlertMiddleware{quota: q, next: next}
}

// Handle checks the quota for the alert's key. If within quota the alert is
// forwarded to the next handler; otherwise it is dropped and nil is returned.
func (m *AlertMiddleware) Handle(a alerting.Alert) error {
	key := alertKey(a)
	if !m.quota.Allow(key) {
		return nil
	}
	return m.next.Handle(a)
}

// alertKey builds a deduplication key from the alert's level, tag, and message.
func alertKey(a alerting.Alert) string {
	return fmt.Sprintf("%s:%s:%s", a.Level, a.Tag, a.Message)
}
