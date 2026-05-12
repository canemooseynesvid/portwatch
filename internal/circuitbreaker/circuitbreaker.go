// Package circuitbreaker implements a simple circuit breaker that prevents
// repeated alert delivery when a downstream handler is failing continuously.
package circuitbreaker

import (
	"sync"
	"time"
)

// State represents the current state of the circuit breaker.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // blocking calls
	StateHalfOpen              // testing recovery
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Breaker tracks consecutive failures and opens the circuit after a threshold.
type Breaker struct {
	mu           sync.Mutex
	state        State
	failures      int
	threshold     int
	resetTimeout  time.Duration
	openedAt      time.Time
	clock         func() time.Time
}

// New returns a Breaker that opens after threshold consecutive failures
// and attempts recovery after resetTimeout.
func New(threshold int, resetTimeout time.Duration) *Breaker {
	return NewWithClock(threshold, resetTimeout, time.Now)
}

// NewWithClock is like New but accepts a custom clock for testing.
func NewWithClock(threshold int, resetTimeout time.Duration, clock func() time.Time) *Breaker {
	return &Breaker{
		threshold:    threshold,
		resetTimeout: resetTimeout,
		clock:        clock,
		state:        StateClosed,
	}
}

// Allow reports whether the caller is permitted to proceed.
// It transitions from Open to HalfOpen once the reset timeout elapses.
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if b.clock().Sub(b.openedAt) >= b.resetTimeout {
			b.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess resets the breaker to Closed state.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure increments the failure counter and opens the circuit
// if the threshold is reached.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.threshold {
		b.state = StateOpen
		b.openedAt = b.clock()
	}
}

// State returns the current circuit state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
