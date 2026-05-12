package backoff

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestNext_FirstDelayEqualsBase(t *testing.T) {
	cfg := Config{Base: 100 * time.Millisecond, Max: 10 * time.Second, Factor: 2.0, Jitter: 0}
	b := NewWithClock(cfg, fixedClock(time.Now()))

	got := b.Next()
	if got != 100*time.Millisecond {
		t.Fatalf("expected 100ms, got %v", got)
	}
}

func TestNext_DoublesEachAttempt(t *testing.T) {
	cfg := Config{Base: 100 * time.Millisecond, Max: 10 * time.Second, Factor: 2.0, Jitter: 0}
	b := NewWithClock(cfg, fixedClock(time.Now()))

	d1 := b.Next() // 100ms
	d2 := b.Next() // 200ms
	d3 := b.Next() // 400ms

	if d2 != 2*d1 {
		t.Fatalf("expected d2=2*d1, got d1=%v d2=%v", d1, d2)
	}
	if d3 != 2*d2 {
		t.Fatalf("expected d3=2*d2, got d2=%v d3=%v", d2, d3)
	}
}

func TestNext_CapsAtMax(t *testing.T) {
	cfg := Config{Base: 1 * time.Second, Max: 3 * time.Second, Factor: 2.0, Jitter: 0}
	b := NewWithClock(cfg, fixedClock(time.Now()))

	for i := 0; i < 10; i++ {
		d := b.Next()
		if d > 3*time.Second {
			t.Fatalf("delay %v exceeded max on attempt %d", d, i)
		}
	}
}

func TestReset_ClearsAttempts(t *testing.T) {
	cfg := Config{Base: 100 * time.Millisecond, Max: 10 * time.Second, Factor: 2.0, Jitter: 0}
	b := NewWithClock(cfg, fixedClock(time.Now()))

	b.Next()
	b.Next()
	if b.Attempts() != 2 {
		t.Fatalf("expected 2 attempts, got %d", b.Attempts())
	}

	b.Reset()
	if b.Attempts() != 0 {
		t.Fatalf("expected 0 after reset, got %d", b.Attempts())
	}

	// Delay should be back to base after reset.
	if d := b.Next(); d != 100*time.Millisecond {
		t.Fatalf("expected base delay after reset, got %v", d)
	}
}

func TestNext_JitterStaysPositive(t *testing.T) {
	cfg := Config{Base: time.Millisecond, Max: time.Second, Factor: 2.0, Jitter: 0.9}
	b := New(cfg)

	for i := 0; i < 50; i++ {
		if d := b.Next(); d < time.Millisecond {
			t.Fatalf("jittered delay below 1ms: %v", d)
		}
	}
}

func TestDefaultConfig_SaneValues(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Base <= 0 || cfg.Max <= 0 || cfg.Factor <= 1 {
		t.Fatalf("default config has invalid values: %+v", cfg)
	}
}
