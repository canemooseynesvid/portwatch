// Package quota enforces per-key alert emission limits over a sliding window.
// It tracks how many alerts have been emitted for a given key and blocks
// further emission once the configured maximum is reached within the window.
package quota

import (
	"sync"
	"time"
)

// clock allows injecting a fake time source in tests.
type clock func() time.Time

// entry holds the count and window start for a single key.
type entry struct {
	count    int
	windowAt time.Time
}

// Quota enforces a maximum number of events per key within a rolling window.
type Quota struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	now     clock
	buckets map[string]*entry
}

// New creates a Quota that allows at most max events per key within window.
func New(max int, window time.Duration) *Quota {
	return NewWithClock(max, window, time.Now)
}

// NewWithClock creates a Quota with an injectable clock (useful for testing).
func NewWithClock(max int, window time.Duration, now clock) *Quota {
	return &Quota{
		max:     max,
		window:  window,
		now:     now,
		buckets: make(map[string]*entry),
	}
}

// Allow returns true and increments the counter if the key is within quota.
// Returns false if the quota has been exhausted for the current window.
func (q *Quota) Allow(key string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.now()
	e, ok := q.buckets[key]
	if !ok || now.Sub(e.windowAt) >= q.window {
		q.buckets[key] = &entry{count: 1, windowAt: now}
		return true
	}
	if e.count >= q.max {
		return false
	}
	e.count++
	return true
}

// Remaining returns how many more events are allowed for key in the current window.
func (q *Quota) Remaining(key string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.now()
	e, ok := q.buckets[key]
	if !ok || now.Sub(e.windowAt) >= q.window {
		return q.max
	}
	rem := q.max - e.count
	if rem < 0 {
		return 0
	}
	return rem
}

// Reset clears the quota state for all keys.
func (q *Quota) Reset() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.buckets = make(map[string]*entry)
}
