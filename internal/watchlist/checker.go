package watchlist

import (
	"fmt"

	"github.com/example/portwatch/internal/alerting"
	"github.com/example/portwatch/internal/portscanner"
)

// Checker watches a Watchlist and emits alerts when watched ports appear or disappear.
type Checker struct {
	wl      *Watchlist
	alerter *alerting.Alerter
}

// NewChecker creates a Checker backed by the given Watchlist and Alerter.
func NewChecker(wl *Watchlist, alerter *alerting.Alerter) *Checker {
	return &Checker{wl: wl, alerter: alerter}
}

// CheckAdded inspects a newly observed port entry and emits an Info alert if it
// matches a watched port.
func (c *Checker) CheckAdded(e portscanner.Entry) {
	if !c.wl.Contains(e.Port, e.Protocol) {
		return
	}
	alert := alerting.NewPortBindAlert(e)
	alert.Level = alerting.LevelInfo
	alert.Message = formatWatchedMessage(e)
	c.alerter.Send(alert)
}

// CheckRemoved inspects a recently closed port entry and emits an Info alert if
// it matches a watched port.
func (c *Checker) CheckRemoved(e portscanner.Entry) {
	if !c.wl.Contains(e.Port, e.Protocol) {
		return
	}
	alert := alerting.NewPortClosedAlert(e)
	alert.Level = alerting.LevelInfo
	alert.Message = formatWatchedClosedMessage(e)
	c.alerter.Send(alert)
}

func formatWatchedMessage(e portscanner.Entry) string {
	return fmt.Sprintf("watched port bound: %s/%d (pid %d)", e.Protocol, e.Port, e.PID)
}

func formatWatchedClosedMessage(e portscanner.Entry) string {
	return fmt.Sprintf("watched port closed: %s/%d (pid %d)", e.Protocol, e.Port, e.PID)
}
