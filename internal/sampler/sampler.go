// Package sampler provides adaptive scan interval sampling based on
// observed port-change activity. When activity is high the sampler
// shortens the effective poll interval; when quiet it backs off toward
// the configured maximum to reduce overhead.
package sampler

import (
	"sync"
	"time"
)

// Sampler tracks recent scan deltas and returns a recommended next
// poll duration that lies within [Min, Max].
type Sampler struct {
	mu      sync.Mutex
	Min     time.Duration
	Max     time.Duration
	window  int // number of recent deltas to consider
	deltas  []int // recent port-change counts
	thresh  int   // delta count that triggers minimum interval
}

// New returns a Sampler with the supplied bounds and sliding-window size.
// thresh is the cumulative change count within the window that causes the
// sampler to recommend Min.
func New(min, max time.Duration, window, thresh int) *Sampler {
	if window < 1 {
		window = 1
	}
	return &Sampler{
		Min:    min,
		Max:    max,
		window: window,
		thresh: thresh,
		deltas: make([]int, 0, window),
	}
}

// Record registers the number of port changes observed in the latest scan.
func (s *Sampler) Record(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deltas = append(s.deltas, delta)
	if len(s.deltas) > s.window {
		s.deltas = s.deltas[len(s.deltas)-s.window:]
	}
}

// Next returns the recommended poll interval for the upcoming scan.
func (s *Sampler) Next() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.deltas) == 0 {
		return s.Max
	}

	sum := 0
	for _, d := range s.deltas {
		sum += d
	}

	if sum >= s.thresh {
		return s.Min
	}

	// Linear interpolation between Min and Max based on activity ratio.
	ratio := float64(sum) / float64(s.thresh)
	span := float64(s.Max - s.Min)
	return s.Max - time.Duration(ratio*span)
}

// Reset clears all recorded deltas.
func (s *Sampler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deltas = s.deltas[:0]
}
