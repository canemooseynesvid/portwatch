// Package watchlist manages a set of ports that should be actively watched
// for binding events, emitting elevated alerts when they appear.
package watchlist

import (
	"fmt"
	"sync"

	"portwatch/internal/portscanner"
)

// Entry describes a single watched port with optional metadata.
type Entry struct {
	Port     uint16
	Protocol string // "tcp" or "udp"
	Label    string // human-readable name, e.g. "SSH"
}

// key returns a canonical string key for the entry.
func key(protocol string, port uint16) string {
	return fmt.Sprintf("%s:%d", protocol, port)
}

// Watchlist holds a thread-safe set of watched port entries.
type Watchlist struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New creates an empty Watchlist.
func New() *Watchlist {
	return &Watchlist{
		entries: make(map[string]Entry),
	}
}

// Add registers a port entry in the watchlist.
func (w *Watchlist) Add(e Entry) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries[key(e.Protocol, e.Port)] = e
}

// Remove deregisters a port entry from the watchlist.
func (w *Watchlist) Remove(protocol string, port uint16) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.entries, key(protocol, port))
}

// Contains reports whether the given scanner entry is on the watchlist.
func (w *Watchlist) Contains(se portscanner.Entry) (Entry, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	e, ok := w.entries[key(se.Protocol, se.LocalPort)]
	return e, ok
}

// All returns a snapshot copy of all watchlist entries.
func (w *Watchlist) All() []Entry {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]Entry, 0, len(w.entries))
	for _, e := range w.entries {
		out = append(out, e)
	}
	return out
}

// Len returns the number of entries in the watchlist.
func (w *Watchlist) Len() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.entries)
}
