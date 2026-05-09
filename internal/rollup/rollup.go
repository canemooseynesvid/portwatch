// Package rollup aggregates repeated alerts into a single summary alert
// when the same key fires more than a configured threshold within a window.
package rollup

import (
	"fmt"
	"sync"
	"time"
)

// Event records a single occurrence of an alert key.
type Event struct {
	Key       string
	Message   string
	OccurredAt time.Time
}

// Summary is emitted when a rollup threshold is crossed.
type Summary struct {
	Key      string
	Count    int
	First    time.Time
	Last     time.Time
	Sample   string // message from the first occurrence
}

// String returns a human-readable rollup summary.
func (s Summary) String() string {
	return fmt.Sprintf("[rollup] %s fired %d times between %s and %s (sample: %s)",
		s.Key, s.Count,
		s.First.Format(time.RFC3339),
		s.Last.Format(time.RFC3339),
		s.Sample,
	)
}

type bucket struct {
	count  int
	first  time.Time
	last   time.Time
	sample string
}

// Aggregator groups repeated events and emits a Summary once the threshold
// is reached, then resets the bucket for that key.
type Aggregator struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	window    time.Duration
	threshold int
	clock     func() time.Time
}

// New returns an Aggregator that fires a Summary after threshold occurrences
// of the same key within window.
func New(threshold int, window time.Duration) *Aggregator {
	return NewWithClock(threshold, window, time.Now)
}

// NewWithClock is like New but accepts a custom clock for testing.
func NewWithClock(threshold int, window time.Duration, clock func() time.Time) *Aggregator {
	return &Aggregator{
		buckets:   make(map[string]*bucket),
		window:    window,
		threshold: threshold,
		clock:     clock,
	}
}

// Record registers one occurrence of key with the given message.
// It returns a non-nil Summary when the threshold is reached; otherwise nil.
func (a *Aggregator) Record(key, message string) *Summary {
	now := a.clock()

	a.mu.Lock()
	defer a.mu.Unlock()

	b, ok := a.buckets[key]
	if !ok || now.Sub(b.last) > a.window {
		b = &bucket{sample: message, first: now}
		a.buckets[key] = b
	}

	b.count++
	b.last = now

	if b.count >= a.threshold {
		sum := &Summary{
			Key:    key,
			Count:  b.count,
			First:  b.first,
			Last:   b.last,
			Sample: b.sample,
		}
		delete(a.buckets, key)
		return sum
	}

	return nil
}

// Reset clears all accumulated state.
func (a *Aggregator) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.buckets = make(map[string]*bucket)
}
