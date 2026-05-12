package dedup

import (
	"testing"
	"time"

	"portwatch/internal/alerting"
)

func fixedClock(t time.Time) clock {
	return func() time.Time { return t }
}

func makeAlert(level alerting.Level, tag, msg string) alerting.Alert {
	return alerting.Alert{
		Level:   level,
		Tag:     tag,
		Message: msg,
	}
}

func TestIsDuplicate_FirstCallNotDuplicate(t *testing.T) {
	now := time.Now()
	d := NewWithClock(time.Minute, fixedClock(now))
	a := makeAlert(alerting.LevelWarn, "port", "port 8080 opened")
	if d.IsDuplicate(a) {
		t.Fatal("expected first call to not be a duplicate")
	}
}

func TestIsDuplicate_SecondCallWithinWindowIsDuplicate(t *testing.T) {
	now := time.Now()
	d := NewWithClock(time.Minute, fixedClock(now))
	a := makeAlert(alerting.LevelWarn, "port", "port 8080 opened")
	d.IsDuplicate(a)
	if !d.IsDuplicate(a) {
		t.Fatal("expected second identical call within window to be a duplicate")
	}
}

func TestIsDuplicate_AfterWindowExpires_NotDuplicate(t *testing.T) {
	now := time.Now()
	current := now
	clock := func() time.Time { return current }
	d := NewWithClock(time.Minute, clock)
	a := makeAlert(alerting.LevelWarn, "port", "port 8080 opened")
	d.IsDuplicate(a)
	current = now.Add(2 * time.Minute)
	if d.IsDuplicate(a) {
		t.Fatal("expected call after window expiry to not be a duplicate")
	}
}

func TestIsDuplicate_DifferentAlertsAreIndependent(t *testing.T) {
	now := time.Now()
	d := NewWithClock(time.Minute, fixedClock(now))
	a1 := makeAlert(alerting.LevelWarn, "port", "port 8080 opened")
	a2 := makeAlert(alerting.LevelCritical, "port", "port 443 conflict")
	d.IsDuplicate(a1)
	if d.IsDuplicate(a2) {
		t.Fatal("expected different alerts to be independent")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	current := now
	clock := func() time.Time { return current }
	d := NewWithClock(time.Minute, clock)
	a := makeAlert(alerting.LevelInfo, "scan", "scan complete")
	d.IsDuplicate(a)
	if d.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", d.Len())
	}
	current = now.Add(2 * time.Minute)
	d.Purge()
	if d.Len() != 0 {
		t.Fatalf("expected 0 entries after purge, got %d", d.Len())
	}
}

func TestPurge_RetainsActiveEntries(t *testing.T) {
	now := time.Now()
	current := now
	clock := func() time.Time { return current }
	d := NewWithClock(time.Minute, clock)
	a := makeAlert(alerting.LevelInfo, "scan", "scan complete")
	d.IsDuplicate(a)
	current = now.Add(30 * time.Second)
	d.Purge()
	if d.Len() != 1 {
		t.Fatalf("expected entry to be retained, got %d entries", d.Len())
	}
}
