// Package window provides a sliding-window counter used to track
// event frequency over a rolling time period.
package window

import (
	"sync"
	"time"
)

// Clock allows injecting a fake time source in tests.
type Clock func() time.Time

// entry records a single timestamped event.
type entry struct {
	at time.Time
}

// Counter is a thread-safe sliding-window event counter keyed by string.
type Counter struct {
	mu     sync.Mutex
	window time.Duration
	clock  Clock
	buckets map[string][]entry
}

// New returns a Counter with the given window duration using the real clock.
func New(window time.Duration) *Counter {
	return NewWithClock(window, time.Now)
}

// NewWithClock returns a Counter with an injectable clock, useful for testing.
func NewWithClock(window time.Duration, clock Clock) *Counter {
	return &Counter{
		window:  window,
		clock:   clock,
		buckets: make(map[string][]entry),
	}
}

// Record adds one event for the given key at the current time and returns
// the total number of events within the window after the addition.
func (c *Counter) Record(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	c.buckets[key] = append(c.buckets[key], entry{at: now})
	c.buckets[key] = c.prune(c.buckets[key], now)
	return len(c.buckets[key])
}

// Count returns the number of events within the window for the given key
// without recording a new event.
func (c *Counter) Count(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	c.buckets[key] = c.prune(c.buckets[key], now)
	return len(c.buckets[key])
}

// Reset clears all recorded events for the given key.
func (c *Counter) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.buckets, key)
}

// prune removes events older than the window from the slice and returns
// the trimmed slice. Must be called with c.mu held.
func (c *Counter) prune(events []entry, now time.Time) []entry {
	cutoff := now.Add(-c.window)
	i := 0
	for i < len(events) && events[i].at.Before(cutoff) {
		i++
	}
	return events[i:]
}
