// Package trend tracks port binding activity over time and detects
// anomalous spikes in new bindings within a rolling time window.
package trend

import (
	"sync"
	"time"
)

// Event records a single port-binding observation.
type Event struct {
	Port     uint16
	Protocol string
	At       time.Time
}

// Spike is returned when the binding rate for a key exceeds the threshold.
type Spike struct {
	Key   string
	Count int
	Window time.Duration
}

// Tracker accumulates binding events and reports spikes.
type Tracker struct {
	mu        sync.Mutex
	events    []Event
	window    time.Duration
	threshold int
	clock     func() time.Time
}

// New creates a Tracker that reports a Spike when more than threshold
// new bindings occur within window.
func New(window time.Duration, threshold int) *Tracker {
	return NewWithClock(window, threshold, time.Now)
}

// NewWithClock creates a Tracker with an injectable clock (for testing).
func NewWithClock(window time.Duration, threshold int, clock func() time.Time) *Tracker {
	return &Tracker{
		window:    window,
		threshold: threshold,
		clock:     clock,
	}
}

// Record adds a binding event and returns any detected Spike, or nil.
func (t *Tracker) Record(e Event) *Spike {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.clock()
	cutoff := now.Add(-t.window)

	// Append and prune in one pass.
	t.events = append(t.events, e)
	filtered := t.events[:0]
	for _, ev := range t.events {
		if ev.At.After(cutoff) {
			filtered = append(filtered, ev)
		}
	}
	t.events = filtered

	if len(t.events) > t.threshold {
		return &Spike{
			Key:    e.Protocol + ":",
			Count:  len(t.events),
			Window: t.window,
		}
	}
	return nil
}

// Len returns the number of events currently in the window.
func (t *Tracker) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.events)
}

// Reset discards all recorded events.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = t.events[:0]
}
