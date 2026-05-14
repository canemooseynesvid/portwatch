package quota

import (
	"testing"
	"time"

	"portwatch/internal/alerting"
)

var fixedNow = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return fixedNow }

func TestAllow_WithinQuota(t *testing.T) {
	q := NewWithClock(3, time.Minute, fixedClock)
	for i := 0; i < 3; i++ {
		if !q.Allow("k") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
}

func TestAllow_ExceedsQuota(t *testing.T) {
	q := NewWithClock(2, time.Minute, fixedClock)
	q.Allow("k")
	q.Allow("k")
	if q.Allow("k") {
		t.Fatal("expected Allow=false after quota exhausted")
	}
}

func TestAllow_SeparateKeysAreIndependent(t *testing.T) {
	q := NewWithClock(1, time.Minute, fixedClock)
	q.Allow("a")
	if !q.Allow("b") {
		t.Fatal("expected Allow=true for independent key")
	}
}

func TestAllow_WindowExpiry(t *testing.T) {
	now := fixedNow
	q := NewWithClock(1, time.Minute, func() time.Time { return now })
	q.Allow("k") // exhausts quota
	now = now.Add(61 * time.Second)
	if !q.Allow("k") {
		t.Fatal("expected Allow=true after window expired")
	}
}

func TestRemaining_DecreasesWithUse(t *testing.T) {
	q := NewWithClock(5, time.Minute, fixedClock)
	if got := q.Remaining("k"); got != 5 {
		t.Fatalf("expected 5 remaining, got %d", got)
	}
	q.Allow("k")
	q.Allow("k")
	if got := q.Remaining("k"); got != 3 {
		t.Fatalf("expected 3 remaining, got %d", got)
	}
}

func TestReset_ClearsState(t *testing.T) {
	q := NewWithClock(1, time.Minute, fixedClock)
	q.Allow("k")
	q.Reset()
	if !q.Allow("k") {
		t.Fatal("expected Allow=true after reset")
	}
}

// --- middleware ---

type collectorHandler struct{ alerts []alerting.Alert }

func (c *collectorHandler) Handle(a alerting.Alert) error {
	c.alerts = append(c.alerts, a)
	return nil
}

func makeAlert(msg string) alerting.Alert {
	return alerting.Alert{Level: alerting.LevelWarn, Tag: "test", Message: msg}
}

func TestMiddleware_ForwardsWithinQuota(t *testing.T) {
	q := NewWithClock(2, time.Minute, fixedClock)
	col := &collectorHandler{}
	m := NewAlertMiddleware(q, col)
	m.Handle(makeAlert("port 8080 bound"))
	m.Handle(makeAlert("port 8080 bound"))
	if len(col.alerts) != 2 {
		t.Fatalf("expected 2 forwarded alerts, got %d", len(col.alerts))
	}
}

func TestMiddleware_DropsWhenExhausted(t *testing.T) {
	q := NewWithClock(1, time.Minute, fixedClock)
	col := &collectorHandler{}
	m := NewAlertMiddleware(q, col)
	m.Handle(makeAlert("port 9090 bound"))
	m.Handle(makeAlert("port 9090 bound")) // should be dropped
	if len(col.alerts) != 1 {
		t.Fatalf("expected 1 forwarded alert, got %d", len(col.alerts))
	}
}
