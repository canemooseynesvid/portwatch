package sequence_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/sequence"
)

func makeAlert(msg string) alerting.Alert {
	return alerting.Alert{
		Level:     alerting.LevelInfo,
		Message:   msg,
		Timestamp: time.Now(),
	}
}

func TestNext_Monotonic(t *testing.T) {
	s := sequence.New()
	for i := uint64(1); i <= 5; i++ {
		if got := s.Next(); got != i {
			t.Fatalf("expected %d, got %d", i, got)
		}
	}
}

func TestReset_ResetsCounter(t *testing.T) {
	s := sequence.New()
	s.Next()
	s.Next()
	s.Reset()
	if got := s.Next(); got != 1 {
		t.Fatalf("expected 1 after reset, got %d", got)
	}
}

func TestTag_PrefixesMessage(t *testing.T) {
	s := sequence.New()
	a := s.Tag(makeAlert("port opened"))
	if !strings.HasPrefix(a.Message, "[seq=1] ") {
		t.Fatalf("unexpected message: %q", a.Message)
	}
	if !strings.Contains(a.Message, "port opened") {
		t.Fatalf("original message lost: %q", a.Message)
	}
}

func TestTag_DoesNotMutateOriginal(t *testing.T) {
	s := sequence.New()
	orig := makeAlert("original")
	tagged := s.Tag(orig)
	if orig.Message == tagged.Message {
		t.Fatal("original alert was mutated")
	}
}

func TestMiddleware_ForwardsTaggedAlert(t *testing.T) {
	s := sequence.New()
	var received alerting.Alert
	next := alerting.HandlerFunc(func(a alerting.Alert) error {
		received = a
		return nil
	})
	mw := s.Middleware(next)
	if err := mw(makeAlert("hello")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(received.Message, "[seq=1] ") {
		t.Fatalf("expected seq prefix, got: %q", received.Message)
	}
}

func TestMiddleware_NilNext(t *testing.T) {
	s := sequence.New()
	mw := s.Middleware(nil)
	if err := mw(makeAlert("noop")); err != nil {
		t.Fatalf("expected nil error for nil next, got: %v", err)
	}
}

func TestNext_ConcurrentSafe(t *testing.T) {
	s := sequence.New()
	const goroutines = 50
	results := make([]uint64, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			results[i] = s.Next()
		}()
	}
	wg.Wait()
	seen := make(map[uint64]bool, goroutines)
	for _, v := range results {
		if seen[v] {
			t.Fatalf("duplicate sequence number: %d", v)
		}
		seen[v] = true
	}
}
