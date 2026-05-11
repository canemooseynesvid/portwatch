package jitter_test

import (
	"testing"
	"time"

	"portwatch/internal/jitter"
)

// fixedSource always returns the same value, making tests deterministic.
type fixedSource struct{ val int64 }

func (f *fixedSource) Int63n(n int64) int64 {
	if f.val >= n {
		return n - 1
	}
	return f.val
}

func TestApply_ZeroOffset(t *testing.T) {
	// A fixed source returning 0 gives offset = 0 - window, i.e. base - window.
	base := 10 * time.Second
	factor := 0.1
	window := int64(float64(base) * factor) // 1 000 000 000 ns = 1 s
	src := &fixedSource{val: 0}             // offset = 0 - window = -window
	j := jitter.NewWithSource(factor, src)
	got := j.Apply(base)
	want := time.Duration(int64(base) - window)
	if got != want {
		t.Fatalf("Apply() = %v, want %v", got, want)
	}
}

func TestApply_MaxOffset(t *testing.T) {
	base := 10 * time.Second
	factor := 0.2
	window := int64(float64(base) * factor)
	// fixedSource returns window*2-1, so offset = (window*2-1) - window = window-1
	src := &fixedSource{val: window*2 - 1}
	j := jitter.NewWithSource(factor, src)
	got := j.Apply(base)
	want := time.Duration(int64(base) + window - 1)
	if got != want {
		t.Fatalf("Apply() = %v, want %v", got, want)
	}
}

func TestApply_NeverBelowMillisecond(t *testing.T) {
	// Extremely small base where offset could go negative.
	base := time.Millisecond
	src := &fixedSource{val: 0}
	j := jitter.NewWithSource(0.9, src)
	got := j.Apply(base)
	if got < time.Millisecond {
		t.Fatalf("Apply() = %v, want >= 1ms", got)
	}
}

func TestApply_ZeroDuration(t *testing.T) {
	j := jitter.New(0.1)
	if got := j.Apply(0); got != 0 {
		t.Fatalf("Apply(0) = %v, want 0", got)
	}
}

func TestNew_ClampsBadFactor(t *testing.T) {
	for _, f := range []float64{-1, 0, 1.5} {
		j := jitter.New(f)
		if j.Factor() != 0.1 {
			t.Errorf("New(%v).Factor() = %v, want 0.1 (clamped)", f, j.Factor())
		}
	}
}

func TestApply_WithinBounds(t *testing.T) {
	// Statistical check: 1000 samples must all stay within ±factor*base.
	base := 500 * time.Millisecond
	factor := 0.3
	j := jitter.New(factor)
	lo := time.Duration(float64(base) * (1 - factor))
	hi := time.Duration(float64(base) * (1 + factor))
	for i := 0; i < 1000; i++ {
		d := j.Apply(base)
		if d < lo || d > hi {
			t.Fatalf("sample %d: Apply() = %v out of [%v, %v]", i, d, lo, hi)
		}
	}
}
