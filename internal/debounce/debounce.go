// Package debounce suppresses repeated alerts for the same key within a
// configurable quiet period. Only the first occurrence is forwarded; subsequent
// occurrences are dropped until the window expires.
package debounce

import (
	"sync"
	"time"
)

// Clock allows tests to inject a fake time source.
type Clock func() time.Time

// entry tracks the first-seen timestamp for a key.
type entry struct {
	firstSeen time.Time
}

// Debouncer holds state for key-based debouncing.
type Debouncer struct {
	mu      sync.Mutex
	window  time.Duration
	clock   Clock
	records map[string]entry
}

// New returns a Debouncer with the given quiet window using the real clock.
func New(window time.Duration) *Debouncer {
	return NewWithClock(window, time.Now)
}

// NewWithClock returns a Debouncer with a custom clock (useful for testing).
func NewWithClock(window time.Duration, clock Clock) *Debouncer {
	return &Debouncer{
		window:  window,
		clock:   clock,
		records: make(map[string]entry),
	}
}

// Allow returns true the first time a key is seen within the window.
// Subsequent calls with the same key return false until the window expires.
func (d *Debouncer) Allow(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.clock()
	if e, ok := d.records[key]; ok {
		if now.Sub(e.firstSeen) < d.window {
			return false
		}
	}
	d.records[key] = entry{firstSeen: now}
	return true
}

// Reset removes the record for a key, allowing the next call to Allow to pass.
func (d *Debouncer) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.records, key)
}

// Purge removes all keys whose windows have expired.
func (d *Debouncer) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.clock()
	for k, e := range d.records {
		if now.Sub(e.firstSeen) >= d.window {
			delete(d.records, k)
		}
	}
}

// Len returns the number of active (non-expired) keys currently tracked.
func (d *Debouncer) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.clock()
	count := 0
	for _, e := range d.records {
		if now.Sub(e.firstSeen) < d.window {
			count++
		}
	}
	return count
}
