package budget_test

import (
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/budget"
)

var (
	t0       = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	window   = 10 * time.Second
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_WithinBudget(t *testing.T) {
	b := budget.NewWithClock(3, window, fixedClock(t0))
	for i := 0; i < 3; i++ {
		if !b.Allow() {
			t.Fatalf("call %d: expected Allow()=true", i+1)
		}
	}
}

func TestAllow_ExceedsBudget(t *testing.T) {
	b := budget.NewWithClock(2, window, fixedClock(t0))
	b.Allow()
	b.Allow()
	if b.Allow() {
		t.Fatal("expected Allow()=false after budget exhausted")
	}
}

func TestAllow_WindowResets(t *testing.T) {
	now := t0
	clock := func() time.Time { return now }
	b := budget.NewWithClock(1, window, clock)

	b.Allow() // exhaust
	if b.Allow() {
		t.Fatal("expected false within window")
	}

	now = t0.Add(window + time.Millisecond)
	if !b.Allow() {
		t.Fatal("expected true after window reset")
	}
}

func TestRemaining_DecreasesAndResets(t *testing.T) {
	now := t0
	clock := func() time.Time { return now }
	b := budget.NewWithClock(3, window, clock)

	if got := b.Remaining(); got != 3 {
		t.Fatalf("want 3, got %d", got)
	}
	b.Allow()
	if got := b.Remaining(); got != 2 {
		t.Fatalf("want 2, got %d", got)
	}

	now = t0.Add(window + time.Millisecond)
	if got := b.Remaining(); got != 3 {
		t.Fatalf("want 3 after reset, got %d", got)
	}
}

func TestMiddleware_PassesWithinBudget(t *testing.T) {
	var received []alerting.Alert
	next := alerting.HandlerFunc(func(a alerting.Alert) error {
		received = append(received, a)
		return nil
	})

	b := budget.NewWithClock(2, window, fixedClock(t0))
	h := budget.NewMiddleware(b, next)

	alert := alerting.NewPortBindAlert(80, "tcp", "")
	for i := 0; i < 3; i++ {
		_ = h.Handle(alert)
	}

	if len(received) != 2 {
		t.Fatalf("want 2 forwarded, got %d", len(received))
	}
}
