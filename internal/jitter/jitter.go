// Package jitter provides utilities for adding randomised jitter to poll
// intervals, preventing thundering-herd effects when multiple goroutines wake
// up simultaneously.
package jitter

import (
	"math/rand"
	"sync"
	"time"
)

// Source is the interface satisfied by any random-number source used by Jitter.
type Source interface {
	Int63n(n int64) int64
}

// Jitter adds a bounded random offset to a base duration.
type Jitter struct {
	mu     sync.Mutex
	src    Source
	factor float64 // fraction of base duration, e.g. 0.2 => ±20 %
}

// New returns a Jitter that spreads durations by up to factor*base on each
// call to Apply. factor must be in the range (0, 1].
func New(factor float64) *Jitter {
	if factor <= 0 || factor > 1 {
		factor = 0.1
	}
	return &Jitter{
		src:    rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
		factor: factor,
	}
}

// NewWithSource returns a Jitter backed by the supplied Source (useful in
// tests where deterministic output is required).
func NewWithSource(factor float64, src Source) *Jitter {
	if factor <= 0 || factor > 1 {
		factor = 0.1
	}
	return &Jitter{src: src, factor: factor}
}

// Apply returns base ± (factor * base * rand), always returning a positive
// duration of at least 1 millisecond.
func (j *Jitter) Apply(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	window := int64(float64(base) * j.factor)
	if window < 1 {
		return base
	}
	j.mu.Lock()
	offset := j.src.Int63n(window*2) - window
	j.mu.Unlock()
	d := time.Duration(int64(base) + offset)
	if d < time.Millisecond {
		d = time.Millisecond
	}
	return d
}

// Factor returns the configured jitter factor.
func (j *Jitter) Factor() float64 { return j.factor }
