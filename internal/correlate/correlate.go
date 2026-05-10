// Package correlate links related alerts together by shared attributes
// such as port, protocol, or process, providing context for alert storms.
package correlate

import (
	"fmt"
	"sync"
	"time"

	"portwatch/internal/alerting"
)

// Group holds a set of alerts that share a correlation key.
type Group struct {
	Key    string
	Alerts []alerting.Alert
	First  time.Time
	Last   time.Time
}

// Correlator groups incoming alerts by a derived key within a time window.
type Correlator struct {
	mu      sync.Mutex
	window  time.Duration
	groups  map[string]*Group
	clock   func() time.Time
}

// New creates a Correlator that groups alerts within the given window.
func New(window time.Duration) *Correlator {
	return NewWithClock(window, time.Now)
}

// NewWithClock creates a Correlator with an injectable clock for testing.
func NewWithClock(window time.Duration, clock func() time.Time) *Correlator {
	return &Correlator{
		window: window,
		groups: make(map[string]*Group),
		clock:  clock,
	}
}

// Record adds an alert to the appropriate correlation group.
// Expired groups are purged before insertion.
func (c *Correlator) Record(a alerting.Alert) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	c.purgeExpired(now)

	k := correlationKey(a)
	g, ok := c.groups[k]
	if !ok {
		g = &Group{Key: k, First: now}
		c.groups[k] = g
	}
	g.Alerts = append(g.Alerts, a)
	g.Last = now
}

// Groups returns a snapshot of all current correlation groups.
func (c *Correlator) Groups() []Group {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	c.purgeExpired(now)

	out := make([]Group, 0, len(c.groups))
	for _, g := range c.groups {
		copy := *g
		copy.Alerts = append([]alerting.Alert(nil), g.Alerts...)
		out = append(out, copy)
	}
	return out
}

// purgeExpired removes groups whose last event is older than the window.
// Caller must hold c.mu.
func (c *Correlator) purgeExpired(now time.Time) {
	cutoff := now.Add(-c.window)
	for k, g := range c.groups {
		if g.Last.Before(cutoff) {
			delete(c.groups, k)
		}
	}
}

// correlationKey derives a grouping key from alert metadata.
func correlationKey(a alerting.Alert) string {
	port, _ := a.Meta["port"]
	proto, _ := a.Meta["protocol"]
	return fmt.Sprintf("%s/%s", proto, port)
}
