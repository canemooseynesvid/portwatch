// Package throttle provides a token-bucket style throttle for limiting
// how frequently a given key can trigger an action within portwatch.
package throttle

import (
	"sync"
	"time"
)

// Clock allows injecting a custom time source for testing.
type Clock func() time.Time

// entry tracks the token count and last refill time for a single key.
type entry struct {
	tokens    int
	lastRefil time.Time
}

// Throttle enforces a maximum burst count per key within a rolling window.
type Throttle struct {
	mu       sync.Mutex
	bucket   map[string]*entry
	burst    int
	window   time.Duration
	clock    Clock
}

// New creates a Throttle that allows up to burst events per key within window.
func New(burst int, window time.Duration) *Throttle {
	return NewWithClock(burst, window, time.Now)
}

// NewWithClock creates a Throttle with a custom clock (useful for testing).
func NewWithClock(burst int, window time.Duration, clock Clock) *Throttle {
	return &Throttle{
		bucket: make(map[string]*entry),
		burst:  burst,
		window: window,
		clock:  clock,
	}
}

// Allow returns true if the key is within its burst limit for the current window.
// Each call consumes one token. Tokens are fully replenished after window elapses.
func (t *Throttle) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.clock()
	e, ok := t.bucket[key]
	if !ok || now.Sub(e.lastRefil) >= t.window {
		t.bucket[key] = &entry{tokens: t.burst - 1, lastRefil: now}
		return true
	}
	if e.tokens <= 0 {
		return false
	}
	e.tokens--
	return true
}

// Reset clears the throttle state for the given key.
func (t *Throttle) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.bucket, key)
}

// Purge removes all expired entries, freeing memory for keys no longer active.
func (t *Throttle) Purge() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.clock()
	for k, e := range t.bucket {
		if now.Sub(e.lastRefil) >= t.window {
			delete(t.bucket, k)
		}
	}
}

// Len returns the number of keys currently tracked.
func (t *Throttle) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.bucket)
}
