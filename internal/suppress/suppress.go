// Package suppress provides a time-based suppression list for alerts.
// Entries added to the list are silenced for a configurable duration,
// after which they are automatically re-enabled.
package suppress

import (
	"sync"
	"time"
)

// Entry holds suppression metadata for a single key.
type Entry struct {
	Key       string
	Suppressed time.Time
	Expires    time.Time
}

// List manages a set of suppressed alert keys.
type List struct {
	mu      sync.RWMutex
	entries map[string]Entry
	now     func() time.Time
}

// New returns an initialised suppression List.
func New() *List {
	return &List{
		entries: make(map[string]Entry),
		now:     time.Now,
	}
}

// Suppress silences alerts for key for the given duration.
func (l *List) Suppress(key string, d time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.entries[key] = Entry{
		Key:        key,
		Suppressed: now,
		Expires:    now.Add(d),
	}
}

// IsSuppressed reports whether key is currently suppressed.
func (l *List) IsSuppressed(key string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	e, ok := l.entries[key]
	if !ok {
		return false
	}
	return l.now().Before(e.Expires)
}

// Remove lifts suppression for key immediately.
func (l *List) Remove(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

// Purge removes all expired entries and returns the number removed.
func (l *List) Purge() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	removed := 0
	for k, e := range l.entries {
		if !now.Before(e.Expires) {
			delete(l.entries, k)
			removed++
		}
	}
	return removed
}

// All returns a snapshot of all current (including expired) entries.
func (l *List) All() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Entry, 0, len(l.entries))
	for _, e := range l.entries {
		out = append(out, e)
	}
	return out
}

// Len returns the total number of entries (including expired).
func (l *List) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}
