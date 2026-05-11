package sampler

import (
	"testing"
	"time"
)

const (
	minInterval = 500 * time.Millisecond
	maxInterval = 10 * time.Second
)

func newSampler() *Sampler {
	return New(minInterval, maxInterval, 5, 10)
}

func TestNext_DefaultsToMax(t *testing.T) {
	s := newSampler()
	if got := s.Next(); got != maxInterval {
		t.Fatalf("expected max %v, got %v", maxInterval, got)
	}
}

func TestNext_HighActivityReturnsMin(t *testing.T) {
	s := newSampler()
	s.Record(5)
	s.Record(5) // sum=10 >= thresh=10
	if got := s.Next(); got != minInterval {
		t.Fatalf("expected min %v, got %v", minInterval, got)
	}
}

func TestNext_LowActivityReturnsMax(t *testing.T) {
	s := newSampler()
	s.Record(0)
	s.Record(0)
	if got := s.Next(); got != maxInterval {
		t.Fatalf("expected max %v, got %v", maxInterval, got)
	}
}

func TestNext_IntermediateActivity(t *testing.T) {
	s := newSampler()
	s.Record(5) // sum=5, ratio=0.5 → midpoint
	got := s.Next()
	mid := minInterval + (maxInterval-minInterval)/2
	// Allow small rounding tolerance.
	diff := got - mid
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Millisecond {
		t.Fatalf("expected ~%v, got %v", mid, got)
	}
}

func TestRecord_SlidingWindow(t *testing.T) {
	s := New(minInterval, maxInterval, 3, 10)
	// Fill window with high activity then push it out.
	s.Record(10)
	s.Record(10)
	s.Record(10)
	// Window is now full; add three zero-delta records to evict the high ones.
	s.Record(0)
	s.Record(0)
	s.Record(0)
	if got := s.Next(); got != maxInterval {
		t.Fatalf("expected max after eviction, got %v", got)
	}
}

func TestReset_ClearsDeltas(t *testing.T) {
	s := newSampler()
	s.Record(10)
	s.Reset()
	if got := s.Next(); got != maxInterval {
		t.Fatalf("expected max after reset, got %v", got)
	}
}

func TestNew_WindowClampedToOne(t *testing.T) {
	s := New(minInterval, maxInterval, 0, 5)
	if s.window != 1 {
		t.Fatalf("expected window=1, got %d", s.window)
	}
}
