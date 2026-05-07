// Package ratelimit provides a token-bucket rate limiter for controlling
// how frequently alerts are emitted for a given key.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter enforces a maximum number of events per window per key.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	max     int
	window  time.Duration
}

type bucket struct {
	count     int
	windowEnd time.Time
}

// New creates a Limiter that allows at most max events per window for each key.
func New(max int, window time.Duration) *Limiter {
	if max <= 0 {
		max = 1
	}
	return &Limiter{
		buckets: make(map[string]*bucket),
		max:     max,
		window:  window,
	}
}

// Allow reports whether an event for the given key is within the rate limit.
// It increments the counter and returns false if the limit has been exceeded.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok || now.After(b.windowEnd) {
		l.buckets[key] = &bucket{
			count:     1,
			windowEnd: now.Add(l.window),
		}
		return true
	}

	if b.count >= l.max {
		return false
	}
	b.count++
	return true
}

// Remaining returns the number of events still allowed for the given key in
// the current window. If the key has no active bucket, the full limit is returned.
func (l *Limiter) Remaining(key string) int {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok || now.After(b.windowEnd) {
		return l.max
	}
	return l.max - b.count
}

// Reset clears the rate-limit state for the given key.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// Prune removes expired buckets to prevent unbounded memory growth.
func (l *Limiter) Prune() {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, b := range l.buckets {
		if now.After(b.windowEnd) {
			delete(l.buckets, k)
		}
	}
}
