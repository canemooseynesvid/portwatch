// Package fingerprint computes a stable identity for a port binding based on
// its observable attributes (protocol, address, port, and owning process).
// This allows portwatch to detect when a "same" port is restarted under a
// different process or binary, which is a meaningful security signal.
package fingerprint

import (
	"crypto/sha256"
	"fmt"
	"io"
)

// Info holds the attributes that contribute to a port fingerprint.
type Info struct {
	Protocol string
	Address  string
	Port     uint16
	PID      int
	Comm     string // process name (from /proc/<pid>/comm)
}

// Fingerprint is a short hex digest uniquely identifying a port binding.
type Fingerprint string

// Compute returns a deterministic Fingerprint for the given Info.
// The digest covers protocol, address, port, pid, and comm so that any
// change in ownership or binary produces a distinct fingerprint.
func Compute(info Info) Fingerprint {
	h := sha256.New()
	_, _ = io.WriteString(h, fmt.Sprintf("%s|%s|%d|%d|%s",
		info.Protocol,
		info.Address,
		info.Port,
		info.PID,
		info.Comm,
	))
	return Fingerprint(fmt.Sprintf("%x", h.Sum(nil))[:16])
}

// Changed reports whether two fingerprints differ, providing a named helper
// for call-sites that want to express intent clearly.
func Changed(a, b Fingerprint) bool {
	return a != b
}

// Store is a lightweight in-memory map from a port key to its last-seen
// Fingerprint.  It is NOT safe for concurrent use; callers must synchronise.
type Store struct {
	m map[string]Fingerprint
}

// NewStore returns an initialised, empty Store.
func NewStore() *Store {
	return &Store{m: make(map[string]Fingerprint)}
}

// portKey returns the canonical map key for an Info without the process fields
// so that we can look up the previous fingerprint for the same port.
func portKey(info Info) string {
	return fmt.Sprintf("%s|%s|%d", info.Protocol, info.Address, info.Port)
}

// Track records the fingerprint for info and returns (previousFingerprint,
// currentFingerprint, changed).  On the first observation changed is false.
func (s *Store) Track(info Info) (prev, cur Fingerprint, changed bool) {
	cur = Compute(info)
	key := portKey(info)
	prev, exists := s.m[key]
	s.m[key] = cur
	if !exists {
		return "", cur, false
	}
	return prev, cur, Changed(prev, cur)
}

// Delete removes the stored fingerprint for a port, typically called when the
// port is observed as closed.
func (s *Store) Delete(info Info) {
	delete(s.m, portKey(info))
}

// Len returns the number of ports currently tracked.
func (s *Store) Len() int {
	return len(s.m)
}
