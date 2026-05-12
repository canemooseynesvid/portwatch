// Package escalate promotes alert severity when the same key fires repeatedly
// within a sliding window. A warning that fires N times becomes an error;
// an error that fires N times becomes a critical.
package escalate

import (
	"sync"
	"time"

	"portwatch/internal/alerting"
)

// Policy controls when escalation occurs.
type Policy struct {
	// Threshold is the number of repeated alerts required to escalate.
	Threshold int
	// Window is the duration over which hits are counted.
	Window time.Duration
}

type hit struct {
	at    time.Time
	level alerting.AlertLevel
}

type clock func() time.Time

// Escalator tracks repeated alerts and promotes their level when a threshold
// is exceeded within the configured window.
type Escalator struct {
	mu     sync.Mutex
	policy Policy
	hits   map[string][]hit
	now    clock
}

// New returns an Escalator with the given policy using the real clock.
func New(p Policy) *Escalator {
	return NewWithClock(p, time.Now)
}

// NewWithClock returns an Escalator with a custom clock (useful for testing).
func NewWithClock(p Policy, c clock) *Escalator {
	return &Escalator{
		policy: p,
		hits:   make(map[string][]hit),
		now:    c,
	}
}

// Evaluate records the alert and returns it, possibly with an elevated level.
func (e *Escalator) Evaluate(a alerting.Alert) alerting.Alert {
	key := a.Key
	if key == "" {
		return a
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := e.now()
	cutoff := now.Add(-e.policy.Window)

	// Prune old hits.
	prev := e.hits[key]
	filtered := prev[:0]
	for _, h := range prev {
		if h.at.After(cutoff) {
			filtered = append(filtered, h)
		}
	}
	filtered = append(filtered, hit{at: now, level: a.Level})
	e.hits[key] = filtered

	if len(filtered) >= e.policy.Threshold {
		a.Level = promote(a.Level)
	}
	return a
}

// Reset clears all recorded hits for the given key.
func (e *Escalator) Reset(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.hits, key)
}

func promote(l alerting.AlertLevel) alerting.AlertLevel {
	switch l {
	case alerting.LevelInfo:
		return alerting.LevelWarning
	case alerting.LevelWarning:
		return alerting.LevelError
	case alerting.LevelError:
		return alerting.LevelCritical
	default:
		return l
	}
}
