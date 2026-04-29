package snapshot

import (
	"sync"
	"time"
)

// Entry represents a single recorded port binding at a point in time.
type Entry struct {
	Protocol string
	Address  string
	Port     uint16
	PID      int
	SeenAt   time.Time
}

// Key uniquely identifies a port binding.
type Key struct {
	Protocol string
	Address  string
	Port     uint16
}

// KeyOf returns the Key for an Entry.
func KeyOf(e Entry) Key {
	return Key{Protocol: e.Protocol, Address: e.Address, Port: e.Port}
}

// Snapshot holds the last-known set of active port bindings.
type Snapshot struct {
	mu      sync.RWMutex
	entries map[Key]Entry
}

// New returns an empty Snapshot.
func New() *Snapshot {
	return &Snapshot{entries: make(map[Key]Entry)}
}

// Set replaces the current snapshot with the provided entries.
func (s *Snapshot) Set(entries []Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := make(map[Key]Entry, len(entries))
	for _, e := range entries {
		next[KeyOf(e)] = e
	}
	s.entries = next
}

// Diff computes added and removed entries compared to the current snapshot,
// then atomically updates the snapshot to the new set.
func (s *Snapshot) Diff(next []Entry) (added, removed []Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nextMap := make(map[Key]Entry, len(next))
	for _, e := range next {
		nextMap[KeyOf(e)] = e
	}

	for k, e := range nextMap {
		if _, exists := s.entries[k]; !exists {
			added = append(added, e)
		}
	}
	for k, e := range s.entries {
		if _, exists := nextMap[k]; !exists {
			removed = append(removed, e)
		}
	}

	s.entries = nextMap
	return added, removed
}

// All returns a copy of the current entries.
func (s *Snapshot) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
