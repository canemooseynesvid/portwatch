package rollup

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alerting"
)

var fixedClock = func() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

func TestRecord_BelowThreshold(t *testing.T) {
	agg := NewWithClock(3, time.Minute, fixedClock)
	if s := agg.Record("key", "msg"); s != nil {
		t.Fatalf("expected nil before threshold, got %v", s)
	}
	if s := agg.Record("key", "msg"); s != nil {
		t.Fatalf("expected nil before threshold, got %v", s)
	}
}

func TestRecord_AtThreshold(t *testing.T) {
	agg := NewWithClock(3, time.Minute, fixedClock)
	agg.Record("key", "first")
	agg.Record("key", "second")
	s := agg.Record("key", "third")
	if s == nil {
		t.Fatal("expected summary at threshold")
	}
	if s.Count != 3 {
		t.Errorf("count = %d, want 3", s.Count)
	}
	if s.Sample != "first" {
		t.Errorf("sample = %q, want %q", s.Sample, "first")
	}
}

func TestRecord_ResetAfterThreshold(t *testing.T) {
	agg := NewWithClock(2, time.Minute, fixedClock)
	agg.Record("key", "a")
	agg.Record("key", "b") // fires
	// bucket should be cleared; next record starts fresh
	if s := agg.Record("key", "c"); s != nil {
		t.Fatalf("expected nil after reset, got %v", s)
	}
}

func TestRecord_WindowExpiry(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	calls := 0
	clock := func() time.Time {
		calls++
		if calls <= 1 {
			return now
		}
		return now.Add(2 * time.Minute) // beyond window
	}
	agg := NewWithClock(2, time.Minute, clock)
	agg.Record("key", "old")
	// Second call is outside the window — bucket resets.
	if s := agg.Record("key", "new"); s != nil {
		t.Fatalf("expected nil after window expiry, got %v", s)
	}
}

func TestReset_ClearsState(t *testing.T) {
	agg := NewWithClock(3, time.Minute, fixedClock)
	agg.Record("key", "a")
	agg.Reset()
	// After reset, threshold counter restarts.
	agg.Record("key", "b")
	if s := agg.Record("key", "c"); s != nil {
		t.Fatalf("expected nil, state should have been reset")
	}
}

func TestAlertMiddleware_PassesThroughBelowThreshold(t *testing.T) {
	var got []alerting.Alert
	next := alerting.HandlerFunc(func(a alerting.Alert) { got = append(got, a) })
	mw := NewAlertMiddleware(next, 3, time.Minute)

	mw.Handle(alerting.Alert{Level: alerting.LevelWarn, Message: "port open", Details: map[string]interface{}{"port": 8080, "protocol": "tcp"}})
	mw.Handle(alerting.Alert{Level: alerting.LevelWarn, Message: "port open", Details: map[string]interface{}{"port": 8080, "protocol": "tcp"}})

	if len(got) != 2 {
		t.Errorf("expected 2 forwarded alerts, got %d", len(got))
	}
}

func TestAlertMiddleware_EmitsRollupAtThreshold(t *testing.T) {
	var got []alerting.Alert
	next := alerting.HandlerFunc(func(a alerting.Alert) { got = append(got, a) })
	mw := NewAlertMiddleware(next, 3, time.Minute)

	for i := 0; i < 3; i++ {
		mw.Handle(alerting.Alert{Level: alerting.LevelWarn, Message: "port open", Details: map[string]interface{}{"port": 9090, "protocol": "tcp"}})
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 total forwarded (2 passthrough + 1 rollup), got %d", len(got))
	}
	last := got[2]
	if last.Message == "port open" {
		t.Error("expected rollup summary message, got original message")
	}
}
