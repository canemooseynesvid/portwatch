// Package cooldown provides a per-key cooldown mechanism that prevents
// repeated actions within a configurable quiet period.
package cooldown

import (
	"sync"
	"time"
)

// Clock abstracts time for testing.
type Clock func() time.Time

// Cooldown tracks the last activation time per key and enforces a minimum
// interval before the same key is allowed to fire again.
type Cooldown struct {
	mu       sync.Mutex
	period   time.Duration
	last     map[string]time.Time
	clock    Clock
}

// New creates a Cooldown with the given quiet period using the real clock.
func New(period time.Duration) *Cooldown {
	return NewWithClock(period, time.Now)
}

// NewWithClock creates a Cooldown with a custom clock (useful for tests).
func NewWithClock(period time.Duration, clock Clock) *Cooldown {
	return &Cooldown{
		period: period,
		last:   make(map[string]time.Time),
		clock:  clock,
	}
}

// Allow returns true if the key has not fired within the cooldown period,
// and records the current time as the latest activation for that key.
func (c *Cooldown) Allow(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	if t, ok := c.last[key]; ok && now.Sub(t) < c.period {
		return false
	}
	c.last[key] = now
	return true
}

// Reset removes the cooldown record for a key, allowing it to fire immediately.
func (c *Cooldown) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.last, key)
}

// Purge removes all keys whose last activation is older than the cooldown period.
// This prevents unbounded memory growth in long-running daemons.
func (c *Cooldown) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	for k, t := range c.last {
		if now.Sub(t) >= c.period {
			delete(c.last, k)
		}
	}
}

// Len returns the number of keys currently tracked.
func (c *Cooldown) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.last)
}
