package sampler

import (
	"time"

	"github.com/your-org/portwatch/internal/snapshot"
)

// Middleware wraps a Sampler and integrates it with the snapshot diff
// lifecycle: after each scan it records the number of changed entries
// and exposes the recommended next interval.
type Middleware struct {
	s *Sampler
}

// NewMiddleware returns a Middleware backed by the given Sampler.
func NewMiddleware(s *Sampler) *Middleware {
	return &Middleware{s: s}
}

// AfterScan should be called after each port scan with the diff produced
// by snapshot.Diff. It records the total number of added and removed
// entries as the activity delta for this scan cycle.
func (m *Middleware) AfterScan(diff snapshot.Diff) {
	delta := len(diff.Added) + len(diff.Removed)
	m.s.Record(delta)
}

// Next returns the recommended duration to wait before the next scan.
func (m *Middleware) Next() time.Duration {
	return m.s.Next()
}
