package window

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestRecord_IncrementsCount(t *testing.T) {
	now := time.Now()
	c := NewWithClock(5*time.Second, fixedClock(now))

	if got := c.Record("tcp:8080"); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
	if got := c.Record("tcp:8080"); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestRecord_SeparateKeysAreIndependent(t *testing.T) {
	now := time.Now()
	c := NewWithClock(5*time.Second, fixedClock(now))

	c.Record("tcp:8080")
	c.Record("tcp:8080")
	c.Record("udp:53")

	if got := c.Count("tcp:8080"); got != 2 {
		t.Errorf("tcp:8080: expected 2, got %d", got)
	}
	if got := c.Count("udp:53"); got != 1 {
		t.Errorf("udp:53: expected 1, got %d", got)
	}
}

func TestRecord_OldEventsExpire(t *testing.T) {
	base := time.Now()
	var now time.Time
	c := NewWithClock(5*time.Second, func() time.Time { return now })

	now = base
	c.Record("tcp:9000")
	c.Record("tcp:9000")

	// Advance past the window.
	now = base.Add(6 * time.Second)
	c.Record("tcp:9000")

	if got := c.Count("tcp:9000"); got != 1 {
		t.Fatalf("expected 1 after expiry, got %d", got)
	}
}

func TestCount_DoesNotAddEvent(t *testing.T) {
	now := time.Now()
	c := NewWithClock(5*time.Second, fixedClock(now))

	c.Count("tcp:443")
	c.Count("tcp:443")

	if got := c.Count("tcp:443"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestReset_ClearsKey(t *testing.T) {
	now := time.Now()
	c := NewWithClock(5*time.Second, fixedClock(now))

	c.Record("tcp:80")
	c.Record("tcp:80")
	c.Reset("tcp:80")

	if got := c.Count("tcp:80"); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestReset_DoesNotAffectOtherKeys(t *testing.T) {
	now := time.Now()
	c := NewWithClock(5*time.Second, fixedClock(now))

	c.Record("tcp:80")
	c.Record("udp:123")
	c.Reset("tcp:80")

	if got := c.Count("udp:123"); got != 1 {
		t.Fatalf("expected udp:123 to be unaffected, got %d", got)
	}
}
