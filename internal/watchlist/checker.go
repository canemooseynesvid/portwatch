package watchlist

import (
	"portwatch/internal/alerting"
	"portwatch/internal/portscanner"
)

// Checker evaluates scanner entries against a Watchlist and emits alerts
// when a watched port is detected as newly bound.
type Checker struct {
	wl      *Watchlist
	alerter *alerting.Alerter
}

// NewChecker creates a Checker backed by the given Watchlist and Alerter.
func NewChecker(wl *Watchlist, a *alerting.Alerter) *Checker {
	return &Checker{wl: wl, alerter: a}
}

// CheckAdded fires a warning-level alert for every newly-added scanner entry
// that appears on the watchlist.
func (c *Checker) CheckAdded(entries []portscanner.Entry) {
	for _, e := range entries {
		if we, ok := c.wl.Contains(e); ok {
			alert := alerting.NewPortBindAlert(e)
			alert.Level = alerting.Warning
			alert.Message = formatWatchedMessage(we, e)
			c.alerter.Send(alert)
		}
	}
}

// CheckRemoved fires an info-level alert when a watched port is no longer bound.
func (c *Checker) CheckRemoved(entries []portscanner.Entry) {
	for _, e := range entries {
		if we, ok := c.wl.Contains(e); ok {
			alert := alerting.NewPortClosedAlert(e)
			alert.Level = alerting.Info
			alert.Message = formatWatchedClosedMessage(we, e)
			c.alerter.Send(alert)
		}
	}
}

func formatWatchedMessage(we Entry, e portscanner.Entry) string {
	if we.Label != "" {
		return "watched port bound: " + we.Label
	}
	return "watched port bound"
}

func formatWatchedClosedMessage(we Entry, e portscanner.Entry) string {
	if we.Label != "" {
		return "watched port released: " + we.Label
	}
	return "watched port released"
}
