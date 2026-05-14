// Package grace provides a drain mechanism that waits for in-flight alert
// handlers to finish before allowing a clean shutdown.
package grace

import (
	"context"
	"sync"
	"time"
)

// DefaultDrainTimeout is used when no deadline is set on the context passed to
// Drain.
const DefaultDrainTimeout = 5 * time.Second

// Drainer tracks in-flight units of work and exposes a Drain method that
// blocks until all work completes or the context is cancelled.
type Drainer struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	closed  bool
}

// New returns a ready-to-use Drainer.
func New() *Drainer {
	return &Drainer{}
}

// Acquire signals that a unit of work has started. It returns false if the
// Drainer has already been closed (i.e. Drain was called), in which case the
// caller must not proceed.
func (d *Drainer) Acquire() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return false
	}
	d.wg.Add(1)
	return true
}

// Release signals that a unit of work has finished. It must be called exactly
// once for every successful Acquire.
func (d *Drainer) Release() {
	d.wg.Done()
}

// Drain closes the Drainer so that no new work can be acquired, then waits
// for all in-flight work to finish. If ctx expires before all work completes,
// Drain returns ctx.Err(); otherwise it returns nil.
func (d *Drainer) Drain(ctx context.Context) error {
	d.mu.Lock()
	d.closed = true
	d.mu.Unlock()

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Wrap returns a handler function that guards the provided fn with the
// Drainer: it calls Acquire before invoking fn and Release when fn returns.
// If the Drainer is already closed, Wrap returns false immediately without
// calling fn.
func (d *Drainer) Wrap(fn func()) bool {
	if !d.Acquire() {
		return false
	}
	go func() {
		defer d.Release()
		fn()
	}()
	return true
}
