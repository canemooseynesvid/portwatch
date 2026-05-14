// Package tag provides a lightweight label store for attaching
// arbitrary key-value metadata to alerts as they flow through the
// portwatch pipeline.
package tag

import (
	"fmt"
	"strings"
	"sync"
)

// Tags is an immutable snapshot of key-value labels.
type Tags map[string]string

// String returns a deterministic, human-readable representation.
func (t Tags) String() string {
	if len(t) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(t))
	for k, v := range t {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// Store is a thread-safe registry that accumulates tags for a named
// entity (e.g. an alert ID or port key).
type Store struct {
	mu   sync.RWMutex
	items map[string]Tags
}

// New returns an empty Store.
func New() *Store {
	return &Store{items: make(map[string]Tags)}
}

// Set attaches key=value to the given entity, creating the tag map if
// it does not yet exist. Existing keys are overwritten.
func (s *Store) Set(entity, key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.items[entity] == nil {
		s.items[entity] = make(Tags)
	}
	s.items[entity][key] = value
}

// Get returns the Tags for entity and whether they exist.
func (s *Store) Get(entity string) (Tags, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.items[entity]
	if !ok {
		return nil, false
	}
	copy := make(Tags, len(t))
	for k, v := range t {
		copy[k] = v
	}
	return copy, true
}

// Delete removes all tags for entity.
func (s *Store) Delete(entity string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, entity)
}

// Len returns the number of entities tracked.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}
