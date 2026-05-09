// Package decay implements an exponential decay scorer for port activity.
// It tracks how "hot" a port is based on recent bind/close events and
// produces a normalized score that fades over time without new activity.
package decay

import (
	"math"
	"sync"
	"time"
)

// Clock allows injecting a fake time source in tests.
type Clock func() time.Time

// entry holds the current score and the last update time for a key.
type entry struct {
	score     float64
	updatedAt time.Time
}

// Scorer maintains per-key exponential decay scores.
type Scorer struct {
	mu       sync.Mutex
	entries  map[string]*entry
	halfLife time.Duration
	clock    Clock
}

// New returns a Scorer with the given half-life duration.
// The half-life controls how quickly scores decay toward zero.
func New(halfLife time.Duration) *Scorer {
	return NewWithClock(halfLife, time.Now)
}

// NewWithClock returns a Scorer with an injectable clock for testing.
func NewWithClock(halfLife time.Duration, clock Clock) *Scorer {
	return &Scorer{
		entries:  make(map[string]*entry),
		halfLife: halfLife,
		clock:    clock,
	}
}

// Record adds delta to the score for key, first decaying the existing value
// based on elapsed time since the last update.
func (s *Scorer) Record(key string, delta float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	e, ok := s.entries[key]
	if !ok {
		e = &entry{updatedAt: now}
		s.entries[key] = e
	}

	elapsed := now.Sub(e.updatedAt)
	e.score = s.decayed(e.score, elapsed) + delta
	e.updatedAt = now
	return e.score
}

// Score returns the current decayed score for key without modifying it.
func (s *Scorer) Score(key string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[key]
	if !ok {
		return 0
	}
	elapsed := s.clock().Sub(e.updatedAt)
	return s.decayed(e.score, elapsed)
}

// Reset removes the entry for key, clearing its score.
func (s *Scorer) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// decayed applies exponential decay: score * 0.5^(elapsed/halfLife).
func (s *Scorer) decayed(score float64, elapsed time.Duration) float64 {
	if s.halfLife <= 0 || elapsed <= 0 {
		return score
	}
	exponent := float64(elapsed) / float64(s.halfLife)
	return score * math.Pow(0.5, exponent)
}
