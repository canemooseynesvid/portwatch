package grace_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"portwatch/internal/grace"
)

func TestAcquireRelease_Basic(t *testing.T) {
	d := grace.New()
	if !d.Acquire() {
		t.Fatal("expected Acquire to return true on open drainer")
	}
	d.Release()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := d.Drain(ctx); err != nil {
		t.Fatalf("Drain returned unexpected error: %v", err)
	}
}

func TestAcquire_ReturnsFalseAfterDrain(t *testing.T) {
	d := grace.New()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = d.Drain(ctx)

	if d.Acquire() {
		t.Fatal("expected Acquire to return false after Drain")
	}
}

func TestDrain_WaitsForInFlightWork(t *testing.T) {
	d := grace.New()

	var reached bool
	var wg sync.WaitGroup
	wg.Add(1)

	if !d.Acquire() {
		t.Fatal("Acquire failed")
	}

	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Millisecond)
		reached = true
		d.Release()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := d.Drain(ctx); err != nil {
		t.Fatalf("Drain returned error: %v", err)
	}
	wg.Wait()
	if !reached {
		t.Fatal("goroutine did not complete before Drain returned")
	}
}

func TestDrain_RespectsContextCancellation(t *testing.T) {
	d := grace.New()

	if !d.Acquire() {
		t.Fatal("Acquire failed")
	}
	// Never release — force a timeout.

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := d.Drain(ctx)
	if err == nil {
		t.Fatal("expected context deadline error, got nil")
	}
}

func TestWrap_ExecutesFn(t *testing.T) {
	d := grace.New()

	var called bool
	var mu sync.Mutex

	ok := d.Wrap(func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})
	if !ok {
		t.Fatal("Wrap returned false on open drainer")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = d.Drain(ctx)

	mu.Lock()
	defer mu.Unlock()
	if !called {
		t.Fatal("wrapped function was not called")
	}
}

func TestWrap_ReturnsFalseWhenClosed(t *testing.T) {
	d := grace.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = d.Drain(ctx)

	if d.Wrap(func() {}) {
		t.Fatal("expected Wrap to return false after Drain")
	}
}
