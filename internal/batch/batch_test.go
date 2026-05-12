package batch_test

import (
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/batch"
)

func makeAlert(msg string) alerting.Alert {
	return alerting.Alert{
		Message: msg,
		Level:   alerting.LevelInfo,
	}
}

func TestBatcher_FlushOnCapacity(t *testing.T) {
	var mu sync.Mutex
	var got [][]alerting.Alert

	b := batch.New(10*time.Second, 3, func(alerts []alerting.Alert) error {
		mu.Lock()
		got = append(got, alerts)
		mu.Unlock()
		return nil
	})
	defer b.Stop()

	_ = b.Add(makeAlert("a"))
	_ = b.Add(makeAlert("b"))
	_ = b.Add(makeAlert("c")) // triggers flush

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 flush, got %d", len(got))
	}
	if len(got[0]) != 3 {
		t.Fatalf("expected batch of 3, got %d", len(got[0]))
	}
}

func TestBatcher_ManualFlush(t *testing.T) {
	var flushed []alerting.Alert

	b := batch.New(10*time.Second, 100, func(alerts []alerting.Alert) error {
		flushed = append(flushed, alerts...)
		return nil
	})
	defer b.Stop()

	_ = b.Add(makeAlert("x"))
	_ = b.Add(makeAlert("y"))
	_ = b.Flush()

	if len(flushed) != 2 {
		t.Fatalf("expected 2 alerts after manual flush, got %d", len(flushed))
	}
}

func TestBatcher_FlushEmptyIsNoop(t *testing.T) {
	called := false
	b := batch.New(10*time.Second, 10, func(alerts []alerting.Alert) error {
		called = true
		return nil
	})
	defer b.Stop()

	if err := b.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("flush function should not be called on empty buffer")
	}
}

func TestBatcher_WindowFlush(t *testing.T) {
	done := make(chan struct{})

	b := batch.New(30*time.Millisecond, 100, func(alerts []alerting.Alert) error {
		close(done)
		return nil
	})
	defer b.Stop()

	_ = b.Add(makeAlert("timer-test"))

	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for window flush")
	}
}

func TestBatcher_SecondBatchAfterCapacityFlush(t *testing.T) {
	var mu sync.Mutex
	var batches [][]alerting.Alert

	b := batch.New(10*time.Second, 2, func(alerts []alerting.Alert) error {
		mu.Lock()
		batches = append(batches, alerts)
		mu.Unlock()
		return nil
	})
	defer b.Stop()

	for _, msg := range []string{"a", "b", "c", "d"} {
		_ = b.Add(makeAlert(msg))
	}

	mu.Lock()
	defer mu.Unlock()
	if len(batches) != 2 {
		t.Fatalf("expected 2 batches, got %d", len(batches))
	}
}
