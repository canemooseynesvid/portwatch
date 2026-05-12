package circuitbreaker_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/circuitbreaker"
)

var (
	t0    = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	reset = 5 * time.Second
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestState_String(t *testing.T) {
	cases := []struct {
		s    circuitbreaker.State
		want string
	}{
		{circuitbreaker.StateClosed, "closed"},
		{circuitbreaker.StateOpen, "open"},
		{circuitbreaker.StateHalfOpen, "half-open"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}

func TestBreaker_ClosedAllowsAll(t *testing.T) {
	b := circuitbreaker.NewWithClock(3, reset, fixedClock(t0))
	for i := 0; i < 10; i++ {
		if !b.Allow() {
			t.Fatalf("expected Allow()=true on iteration %d", i)
		}
	}
}

func TestBreaker_OpensAfterThreshold(t *testing.T) {
	b := circuitbreaker.NewWithClock(3, reset, fixedClock(t0))
	for i := 0; i < 3; i++ {
		b.RecordFailure()
	}
	if b.State() != circuitbreaker.StateOpen {
		t.Fatalf("expected StateOpen, got %s", b.State())
	}
	if b.Allow() {
		t.Fatal("expected Allow()=false when open")
	}
}

func TestBreaker_HalfOpenAfterTimeout(t *testing.T) {
	now := t0
	clock := func() time.Time { return now }
	b := circuitbreaker.NewWithClock(2, reset, clock)
	b.RecordFailure()
	b.RecordFailure()
	now = t0.Add(reset + time.Millisecond)
	if !b.Allow() {
		t.Fatal("expected Allow()=true in half-open state")
	}
	if b.State() != circuitbreaker.StateHalfOpen {
		t.Fatalf("expected StateHalfOpen, got %s", b.State())
	}
}

func TestBreaker_RecordSuccessCloses(t *testing.T) {
	now := t0
	clock := func() time.Time { return now }
	b := circuitbreaker.NewWithClock(2, reset, clock)
	b.RecordFailure()
	b.RecordFailure()
	now = t0.Add(reset + time.Millisecond)
	b.Allow() // transition to half-open
	b.RecordSuccess()
	if b.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected StateClosed after success, got %s", b.State())
	}
}

// --- middleware tests ---

type fakeHandler struct {
	calls  int
	failOn int // fail when calls == failOn (0 = never)
}

func (f *fakeHandler) Handle(_ alerting.Alert) error {
	f.calls++
	if f.failOn > 0 && f.calls == f.failOn {
		return errors.New("handler error")
	}
	return nil
}

func makeAlert(title string) alerting.Alert {
	return alerting.Alert{Level: alerting.LevelWarn, Title: title}
}

func TestAlertMiddleware_PassesWhenClosed(t *testing.T) {
	h := &fakeHandler{}
	m := circuitbreaker.NewAlertMiddleware(h, 3, reset)
	for i := 0; i < 5; i++ {
		if err := m.Handle(makeAlert("test")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if h.calls != 5 {
		t.Fatalf("expected 5 calls, got %d", h.calls)
	}
}

func TestAlertMiddleware_OpensAfterFailures(t *testing.T) {
	h := &fakeHandler{failOn: 1}
	m := circuitbreaker.NewAlertMiddleware(h, 1, reset)
	_ = m.Handle(makeAlert("boom")) // triggers failure
	key := fmt.Sprintf("%s:%s", alerting.LevelWarn, "boom")
	if m.BreakerState(key) != circuitbreaker.StateOpen {
		t.Fatalf("expected circuit open after threshold failure")
	}
	// subsequent calls are dropped silently
	prev := h.calls
	_ = m.Handle(makeAlert("boom"))
	if h.calls != prev {
		t.Fatal("expected call to be dropped while circuit is open")
	}
}
