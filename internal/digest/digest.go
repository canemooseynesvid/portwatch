// Package digest computes and tracks a rolling fingerprint of the active port
// snapshot. It can detect when the overall set of bindings has changed without
// requiring a full diff, which is useful for cheap change-detection in tight
// poll loops.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
)

// Entry is the minimal information needed to fingerprint a single binding.
type Entry struct {
	Protocol string
	Addr     string
	Port     uint16
	Inode    uint64
}

// Tracker maintains the most-recently-computed digest and reports whether the
// snapshot has changed since the last call to Update.
type Tracker struct {
	mu   sync.Mutex
	last string
}

// New returns an initialised Tracker. The first call to Update will always
// report a change because the stored digest starts empty.
func New() *Tracker {
	return &Tracker{}
}

// Update computes a deterministic digest of entries and returns true when it
// differs from the previously stored value. The new digest is stored so
// subsequent calls reflect the latest snapshot.
func (t *Tracker) Update(entries []Entry) (changed bool, digest string) {
	d := compute(entries)
	t.mu.Lock()
	defer t.mu.Unlock()
	changed = d != t.last
	t.last = d
	return changed, d
}

// Last returns the digest recorded by the most recent call to Update.
func (t *Tracker) Last() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.last
}

// compute builds a deterministic SHA-256 fingerprint over the sorted entries.
func compute(entries []Entry) string {
	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		keys = append(keys, fmt.Sprintf("%s|%s|%d|%d", e.Protocol, e.Addr, e.Port, e.Inode))
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		_, _ = fmt.Fprintln(h, k)
	}
	return hex.EncodeToString(h.Sum(nil))
}
