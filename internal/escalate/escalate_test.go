package escalate_test

import (
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/escalate"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeAlert(key string, level alerting.AlertLevel) alerting.Alert {
	return alerting.Alert{Key: key, Level: level, Message: "test"}
}

func TestEscalate_BelowThreshold_LevelUnchanged(t *testing.T) {
	now := time.Now()
	e := escalate.NewWithClock(escalate.Policy{Threshold: 3, Window: time.Minute}, fixedClock(now))

	a := makeAlert("tcp:8080", alerting.LevelWarning)
	for i := 0; i < 2; i++ {
		out := e.Evaluate(a)
		if out.Level != alerting.LevelWarning {
			t.Fatalf("iteration %d: expected Warning, got %v", i, out.Level)
		}
	}
}

func TestEscalate_AtThreshold_LevelPromoted(t *testing.T) {
	now := time.Now()
	e := escalate.NewWithClock(escalate.Policy{Threshold: 3, Window: time.Minute}, fixedClock(now))

	a := makeAlert("tcp:8080", alerting.LevelWarning)
	var out alerting.Alert
	for i := 0; i < 3; i++ {
		out = e.Evaluate(a)
	}
	if out.Level != alerting.LevelError {
		t.Fatalf("expected Error after threshold, got %v", out.Level)
	}
}

func TestEscalate_ErrorPromotesToCritical(t *testing.T) {
	now := time.Now()
	e := escalate.NewWithClock(escalate.Policy{Threshold: 2, Window: time.Minute}, fixedClock(now))

	a := makeAlert("tcp:443", alerting.LevelError)
	var out alerting.Alert
	for i := 0; i < 2; i++ {
		out = e.Evaluate(a)
	}
	if out.Level != alerting.LevelCritical {
		t.Fatalf("expected Critical, got %v", out.Level)
	}
}

func TestEscalate_OldHitsExpire(t *testing.T) {
	base := time.Now()
	current := base
	e := escalate.NewWithClock(escalate.Policy{Threshold: 3, Window: 30 * time.Second}, func() time.Time { return current })

	a := makeAlert("tcp:9090", alerting.LevelWarning)
	e.Evaluate(a)
	e.Evaluate(a)

	// Advance past the window so previous hits expire.
	current = base.Add(time.Minute)
	out := e.Evaluate(a)
	if out.Level != alerting.LevelWarning {
		t.Fatalf("expected Warning after expiry, got %v", out.Level)
	}
}

func TestEscalate_Reset_ClearsHits(t *testing.T) {
	now := time.Now()
	e := escalate.NewWithClock(escalate.Policy{Threshold: 2, Window: time.Minute}, fixedClock(now))

	a := makeAlert("tcp:22", alerting.LevelWarning)
	e.Evaluate(a)
	e.Evaluate(a) // would normally escalate on next call
	e.Reset("tcp:22")

	out := e.Evaluate(a)
	if out.Level != alerting.LevelWarning {
		t.Fatalf("expected Warning after reset, got %v", out.Level)
	}
}

func TestEscalate_EmptyKey_PassesThrough(t *testing.T) {
	e := escalate.New(escalate.Policy{Threshold: 1, Window: time.Minute})
	a := alerting.Alert{Key: "", Level: alerting.LevelInfo, Message: "no key"}
	out := e.Evaluate(a)
	if out.Level != alerting.LevelInfo {
		t.Fatalf("expected Info unchanged, got %v", out.Level)
	}
}
