package lifecycle

import (
	"context"
	"fmt"
	"sync"
)

// Worker is a long-running background task managed by the lifecycle.
type Worker interface {
	// Name returns a human-readable identifier for the worker.
	Name() string
	// Run starts the worker and blocks until ctx is cancelled.
	Run(ctx context.Context) error
}

// Runner launches a set of Workers concurrently and collects their errors.
type Runner struct {
	workers []Worker
}

// NewRunner returns an empty Runner.
func NewRunner() *Runner {
	return &Runner{}
}

// Add registers a worker with the runner.
func (r *Runner) Add(w Worker) {
	r.workers = append(r.workers, w)
}

// Run starts all workers concurrently. It returns when all workers have exited.
// The first non-nil error is returned; remaining workers are given a chance to
// finish because the shared context is cancelled on the first error.
func (r *Runner) Run(ctx context.Context) error {
	if len(r.workers) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(r.workers))
	var wg sync.WaitGroup

	for _, w := range r.workers {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := w.Run(ctx); err != nil && ctx.Err() == nil {
				errCh <- fmt.Errorf("worker %s: %w", w.Name(), err)
				cancel()
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
