package suppress

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/snapshot"
)

// Checker wraps a suppression List and gates alert emission.
type Checker struct {
	list    *List
	default_ time.Duration
}

// NewChecker returns a Checker backed by list with a default suppression
// duration of d applied when Suppress is called without an explicit duration.
func NewChecker(list *List, d time.Duration) *Checker {
	return &Checker{list: list, default_: d}
}

// Allow returns true if the alert for entry should be forwarded.
// It consults the underlying List using the canonical port key.
func (c *Checker) Allow(e snapshot.Entry) bool {
	return !c.list.IsSuppressed(portKey(e))
}

// SuppressEntry silences alerts for entry for the default duration.
func (c *Checker) SuppressEntry(e snapshot.Entry) {
	c.list.Suppress(portKey(e), c.default_)
}

// SuppressFor silences alerts for entry for the given duration.
func (c *Checker) SuppressFor(e snapshot.Entry, d time.Duration) {
	c.list.Suppress(portKey(e), d)
}

// FilterAlert returns nil if the alert's port key is suppressed, otherwise
// returns the alert unchanged. This satisfies a simple middleware signature.
func (c *Checker) FilterAlert(a *alerting.Alert) *alerting.Alert {
	if a == nil {
		return nil
	}
	key, ok := a.Fields["port_key"]
	if !ok {
		return a
	}
	if c.list.IsSuppressed(fmt.Sprintf("%v", key)) {
		return nil
	}
	return a
}

// portKey builds a stable string key from a snapshot entry.
func portKey(e snapshot.Entry) string {
	return fmt.Sprintf("%s:%d", e.Protocol, e.Port)
}
