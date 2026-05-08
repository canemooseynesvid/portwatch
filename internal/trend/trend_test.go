package trend

import (
	"testing"
	"time"
)

var baseTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeEvent(proto string, at time.Time) Event {
	return Event{Port: 8080, Protocol: proto, At: at}
}

func TestRecord_BelowThreshold_NoSpike(t *testing.T) {
	tr := NewWithClock(time.Minute, 3, fixedClock(baseTime))
	for i := 0; i < 3; i++ {
		spike := tr.Record(makeEvent("tcp", baseTime))
		if spike != nil {
			t.Fatalf("expected no spike at count %d, got %+v", i+1, spike)
		}
	}
}

func TestRecord_ExceedsThreshold_ReturnsSpike(t *testing.T) {
	tr := NewWithClock(time.Minute, 3, fixedClock(baseTime))
	var spike *Spike
	for i := 0; i < 4; i++ {
		spike = tr.Record(makeEvent("tcp", baseTime))
	}
	if spike == nil {
		t.Fatal("expected spike, got nil")
	}
	if spike.Count != 4 {
		t.Errorf("expected count 4, got %d", spike.Count)
	}
	if spike.Window != time.Minute {
		t.Errorf("unexpected window %v", spike.Window)
	}
}

func TestRecord_OldEventsExpire(t *testing.T) {
	now := baseTime
	clockFn := func() time.Time { return now }
	tr := NewWithClock(time.Minute, 3, clockFn)

	// Record 3 events in the past (outside window).
	old := baseTime.Add(-2 * time.Minute)
	for i := 0; i < 3; i++ {
		tr.Record(makeEvent("tcp", old))
	}

	// Advance clock; old events should be pruned on next Record.
	now = baseTime.Add(time.Minute)
	spike := tr.Record(makeEvent("tcp", now))
	if spike != nil {
		t.Fatalf("expected no spike after expiry, got %+v", spike)
	}
	if tr.Len() != 1 {
		t.Errorf("expected 1 live event, got %d", tr.Len())
	}
}

func TestReset_ClearsEvents(t *testing.T) {
	tr := NewWithClock(time.Minute, 10, fixedClock(baseTime))
	for i := 0; i < 5; i++ {
		tr.Record(makeEvent("udp", baseTime))
	}
	tr.Reset()
	if tr.Len() != 0 {
		t.Errorf("expected 0 after reset, got %d", tr.Len())
	}
}

func TestLen_ReflectsWindowedCount(t *testing.T) {
	tr := NewWithClock(time.Minute, 100, fixedClock(baseTime))
	for i := 0; i < 7; i++ {
		tr.Record(makeEvent("tcp", baseTime))
	}
	if tr.Len() != 7 {
		t.Errorf("expected 7, got %d", tr.Len())
	}
}
