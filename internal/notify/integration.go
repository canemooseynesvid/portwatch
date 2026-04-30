package notify

import (
	"fmt"

	"portwatch/internal/alerting"
)

// AlertAdapter bridges the alerting system to the Notifier.
// It wraps a Notifier and implements alerting.Handler.
type AlertAdapter struct {
	notifier *Notifier
}

// NewAlertAdapter creates an AlertAdapter backed by the given Notifier.
func NewAlertAdapter(n *Notifier) *AlertAdapter {
	return &AlertAdapter{notifier: n}
}

// Handle satisfies alerting.Handler, forwarding alerts as notifications.
func (a *AlertAdapter) Handle(alert alerting.Alert) {
	subject := fmt.Sprintf("[portwatch/%s] %s", alert.Level, alert.Title)
	body := fmt.Sprintf("%s\nPort: %d  Protocol: %s  PID: %d  Process: %s\nTime: %s",
		alert.Message,
		alert.Port,
		alert.Protocol,
		alert.PID,
		alert.Process,
		alert.Timestamp.Format("2006-01-02 15:04:05"),
	)
	dedupKey := fmt.Sprintf("%s:%d:%s", alert.Protocol, alert.Port, alert.Title)
	_ = a.notifier.Notify(dedupKey, subject, body)
}
