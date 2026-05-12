// Package batch provides alert batching: it accumulates alerts over a
// configurable window and flushes them as a group to a downstream handler.
package batch

import (
	"context"
	"sync"
	"time"

	"github.com/user/portwatch/internal/alerting"
)

// FlushFunc is called with the accumulated batch of alerts when the window
// closes or the capacity is reached.
type FlushFunc func(alerts []alerting.Alert) error

// Batcher accumulates alerts and flushes them periodically.
type Batcher struct {
	mu       sync.Mutex
	buf      []alerting.Alert
	cap      int
	window   time.Duration
	flush    FlushFunc
	clock    func() time.Time
	timer    *time.Timer
	cancel   context.CancelFunc
}

// New creates a Batcher that flushes after window elapses or capacity is
// reached, whichever comes first.
func New(window time.Duration, capacity int, fn FlushFunc) *Batcher {
	return NewWithClock(window, capacity, fn, time.Now)
}

// NewWithClock is like New but accepts a custom clock for testing.
func NewWithClock(window time.Duration, capacity int, fn FlushFunc, clock func() time.Time) *Batcher {
	if capacity <= 0 {
		capacity = 64
	}
	return &Batcher{
		cap:    capacity,
		window: window,
		flush:  fn,
		clock:  clock,
	}
}

// Add appends an alert to the current batch. If the batch reaches capacity it
// is flushed immediately.
func (b *Batcher) Add(a alerting.Alert) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buf) == 0 {
		b.armTimer()
	}
	b.buf = append(b.buf, a)

	if len(b.buf) >= b.cap {
		return b.flushLocked()
	}
	return nil
}

// Flush forces an immediate flush of any buffered alerts.
func (b *Batcher) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

// Stop cancels the background timer without flushing.
func (b *Batcher) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.timer != nil {
		b.timer.Stop()
	}
}

func (b *Batcher) armTimer() {
	if b.timer != nil {
		b.timer.Stop()
	}
	b.timer = time.AfterFunc(b.window, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		_ = b.flushLocked() //nolint:errcheck
	})
}

func (b *Batcher) flushLocked() error {
	if len(b.buf) == 0 {
		return nil
	}
	batch := make([]alerting.Alert, len(b.buf))
	copy(batch, b.buf)
	b.buf = b.buf[:0]
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	return b.flush(batch)
}
