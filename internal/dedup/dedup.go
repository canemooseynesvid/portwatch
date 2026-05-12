// Package dedup provides alert deduplication based on a content hash,
// suppressing repeated identical alerts within a configurable time window.
package dedup

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"portwatch/internal/alerting"
)

// clock is a seam for testing.
type clock func() time.Time

// entry tracks when an alert hash was last seen.
type entry struct {
	lastSeen time.Time
}

// Deduplicator drops alerts whose content hash has been seen within the window.
type Deduplicator struct {
	mu     sync.Mutex
	window time.Duration
	seen   map[string]entry
	now    clock
}

// New returns a Deduplicator with the given suppression window.
func New(window time.Duration) *Deduplicator {
	return NewWithClock(window, time.Now)
}

// NewWithClock returns a Deduplicator using a custom clock (useful in tests).
func NewWithClock(window time.Duration, now clock) *Deduplicator {
	return &Deduplicator{
		window: window,
		seen:   make(map[string]entry),
		now:    now,
	}
}

// IsDuplicate reports whether alert a is a duplicate within the window.
// It also records the alert if it is not a duplicate.
func (d *Deduplicator) IsDuplicate(a alerting.Alert) bool {
	h := hashAlert(a)
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now()
	if e, ok := d.seen[h]; ok && now.Sub(e.lastSeen) < d.window {
		return true
	}
	d.seen[h] = entry{lastSeen: now}
	return false
}

// Purge removes all entries whose window has expired.
func (d *Deduplicator) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now()
	for k, e := range d.seen {
		if now.Sub(e.lastSeen) >= d.window {
			delete(d.seen, k)
		}
	}
}

// Len returns the number of tracked hashes.
func (d *Deduplicator) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
}

// hashAlert produces a stable hash from an alert's level, tag, and message.
func hashAlert(a alerting.Alert) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d|%s|%s", a.Level, a.Tag, a.Message)
	return fmt.Sprintf("%x", h.Sum(nil))
}
