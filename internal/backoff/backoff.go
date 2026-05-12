// Package backoff implements exponential backoff with jitter for retry logic.
package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Clock allows injecting a time source for testing.
type Clock func() time.Time

// Backoff tracks retry state for a single key.
type Backoff struct {
	attempts int
	base     time.Duration
	max      time.Duration
	factor   float64
	jitter   float64
	clock    Clock
}

// Config holds parameters for a Backoff instance.
type Config struct {
	Base   time.Duration
	Max    time.Duration
	Factor float64
	Jitter float64 // fraction of delay to randomise, 0–1
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Base:   200 * time.Millisecond,
		Max:    30 * time.Second,
		Factor: 2.0,
		Jitter: 0.2,
	}
}

// New creates a Backoff with the given config.
func New(cfg Config) *Backoff {
	return NewWithClock(cfg, func() time.Time { return time.Now() })
}

// NewWithClock creates a Backoff with an injectable clock.
func NewWithClock(cfg Config, clock Clock) *Backoff {
	if cfg.Factor <= 1 {
		cfg.Factor = 2.0
	}
	if cfg.Base <= 0 {
		cfg.Base = 200 * time.Millisecond
	}
	if cfg.Max <= 0 {
		cfg.Max = 30 * time.Second
	}
	return &Backoff{
		base:   cfg.Base,
		max:    cfg.Max,
		factor: cfg.Factor,
		jitter: cfg.Jitter,
		clock:  clock,
	}
}

// Next returns the delay for the current attempt and increments the counter.
func (b *Backoff) Next() time.Duration {
	delay := float64(b.base) * math.Pow(b.factor, float64(b.attempts))
	if delay > float64(b.max) {
		delay = float64(b.max)
	}
	if b.jitter > 0 {
		// add ±jitter fraction
		offset := delay * b.jitter * (rand.Float64()*2 - 1)
		delay += offset
		if delay < float64(time.Millisecond) {
			delay = float64(time.Millisecond)
		}
	}
	b.attempts++
	return time.Duration(delay)
}

// Attempts returns the number of times Next has been called.
func (b *Backoff) Attempts() int { return b.attempts }

// Reset clears the attempt counter.
func (b *Backoff) Reset() { b.attempts = 0 }
