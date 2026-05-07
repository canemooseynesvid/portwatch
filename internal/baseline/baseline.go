// Package baseline records and persists a known-good set of port bindings
// so that portwatch can distinguish expected ports from unexpected ones across
// restarts.
package baseline

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry represents a single persisted port binding in the baseline.
type Entry struct {
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     uint16 `json:"port"`
	PID      int    `json:"pid,omitempty"`
	AddedAt  time.Time `json:"added_at"`
}

// Baseline holds the set of approved port bindings.
type Baseline struct {
	mu      sync.RWMutex
	entries map[string]Entry
	path    string
}

// New loads a baseline from the given file path, creating an empty one if the
// file does not exist.
func New(path string) (*Baseline, error) {
	b := &Baseline{
		entries: make(map[string]Entry),
		path:    path,
	}
	if err := b.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return b, nil
}

func entryKey(protocol, address string, port uint16) string {
	return protocol + "|" + address + "|" + string(rune(port))
}

// Contains reports whether the given binding is present in the baseline.
func (b *Baseline) Contains(protocol, address string, port uint16) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.entries[entryKey(protocol, address, port)]
	return ok
}

// Add inserts a binding into the baseline and persists it to disk.
func (b *Baseline) Add(e Entry) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := entryKey(e.Protocol, e.Address, e.Port)
	e.AddedAt = time.Now().UTC()
	b.entries[key] = e
	return b.save()
}

// Remove deletes a binding from the baseline and persists the change.
func (b *Baseline) Remove(protocol, address string, port uint16) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.entries, entryKey(protocol, address, port))
	return b.save()
}

// All returns a copy of all baseline entries.
func (b *Baseline) All() []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]Entry, 0, len(b.entries))
	for _, e := range b.entries {
		out = append(out, e)
	}
	return out
}

func (b *Baseline) load() error {
	f, err := os.Open(b.path)
	if err != nil {
		return err
	}
	defer f.Close()
	var entries []Entry
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return err
	}
	for _, e := range entries {
		b.entries[entryKey(e.Protocol, e.Address, e.Port)] = e
	}
	return nil
}

func (b *Baseline) save() error {
	entries := make([]Entry, 0, len(b.entries))
	for _, e := range b.entries {
		entries = append(entries, e)
	}
	f, err := os.CreateTemp("", "portwatch-baseline-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	if err := json.NewEncoder(f).Encode(entries); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()
	return os.Rename(tmp, b.path)
}
