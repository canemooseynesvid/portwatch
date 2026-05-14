// Package budget enforces a maximum alert emission budget over a rolling
// time window. Once the budget is exhausted, further alerts are dropped
// until the window resets.
package budget

import (
	"sync"
	"time"

	"portwatch/internal/alerting"
)

// clock is a seam for testing.
type clock func() time.Time

// Budget tracks how many alerts have been emitted within the current window
// and rejects new ones once the limit is reached.
type Budget struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	now     clock
	count   int
	windowStart time.Time
}

// New creates a Budget that allows at most limit alerts per window.
func New(limit int, window time.Duration) *Budget {
	return NewWithClock(limit, window, time.Now)
}

// NewWithClock creates a Budget with a custom clock, useful for testing.
func NewWithClock(limit int, window time.Duration, now clock) *Budget {
	return &Budget{
		limit:       limit,
		window:      window,
		now:         now,
		windowStart: now(),
	}
}

// Allow returns true if the alert may be emitted within the current budget.
// It advances the window automatically when the period has elapsed.
func (b *Budget) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.now()
	if now.Sub(b.windowStart) >= b.window {
		b.windowStart = now
		b.count = 0
	}

	if b.count >= b.limit {
		return false
	}
	b.count++
	return true
}

// Remaining returns how many alerts can still be emitted in the current window.
func (b *Budget) Remaining() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.now()
	if now.Sub(b.windowStart) >= b.window {
		return b.limit
	}
	r := b.limit - b.count
	if r < 0 {
		return 0
	}
	return r
}

// NewMiddleware returns an alerting.Handler that drops alerts when the budget
// is exhausted, passing allowed alerts through to next.
func NewMiddleware(b *Budget, next alerting.Handler) alerting.Handler {
	return alerting.HandlerFunc(func(a alerting.Alert) error {
		if !b.Allow() {
			return nil
		}
		return next.Handle(a)
	})
}
