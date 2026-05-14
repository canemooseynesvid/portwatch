// Package sequence assigns monotonically increasing sequence numbers to alerts,
// enabling downstream consumers to detect gaps or reordering.
package sequence

import (
	"fmt"
	"sync/atomic"

	"portwatch/internal/alerting"
)

// Sequencer maintains a global counter and stamps each alert with a sequence number.
type Sequencer struct {
	counter atomic.Uint64
}

// New returns a ready-to-use Sequencer starting at sequence number 1.
func New() *Sequencer {
	return &Sequencer{}
}

// Next returns the next sequence number and advances the counter.
func (s *Sequencer) Next() uint64 {
	return s.counter.Add(1)
}

// Reset sets the counter back to zero. Intended for testing only.
func (s *Sequencer) Reset() {
	s.counter.Store(0)
}

// Tag annotates an alert's message with a sequence number prefix.
// The original alert is not mutated; a shallow copy is returned.
func (s *Sequencer) Tag(a alerting.Alert) alerting.Alert {
	seq := s.Next()
	a.Message = fmt.Sprintf("[seq=%d] %s", seq, a.Message)
	return a
}

// Middleware returns an alerting.HandlerFunc that stamps each alert before
// forwarding it to next.
func (s *Sequencer) Middleware(next alerting.HandlerFunc) alerting.HandlerFunc {
	return func(a alerting.Alert) error {
		if next == nil {
			return nil
		}
		return next(s.Tag(a))
	}
}
