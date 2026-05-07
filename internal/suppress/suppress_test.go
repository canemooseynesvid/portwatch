package suppress

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSuppress_IsSuppressed(t *testing.T) {
	base := time.Now()
	l := New()
	l.now = fixedClock(base)

	l.Suppress("tcp:8080", 5*time.Second)

	if !l.IsSuppressed("tcp:8080") {
		t.Fatal("expected key to be suppressed")
	}
}

func TestSuppress_Expiry(t *testing.T) {
	base := time.Now()
	l := New()
	l.now = fixedClock(base)

	l.Suppress("tcp:9090", 2*time.Second)

	// advance clock past expiry
	l.now = fixedClock(base.Add(3 * time.Second))

	if l.IsSuppressed("tcp:9090") {
		t.Fatal("expected key to have expired")
	}
}

func TestSuppress_Remove(t *testing.T) {
	l := New()
	l.Suppress("tcp:443", time.Hour)
	l.Remove("tcp:443")

	if l.IsSuppressed("tcp:443") {
		t.Fatal("expected key to be removed")
	}
}

func TestSuppress_Purge(t *testing.T) {
	base := time.Now()
	l := New()
	l.now = fixedClock(base)

	l.Suppress("tcp:1111", 1*time.Second)
	l.Suppress("tcp:2222", time.Hour)

	l.now = fixedClock(base.Add(2 * time.Second))

	removed := l.Purge()
	if removed != 1 {
		t.Fatalf("expected 1 purged, got %d", removed)
	}
	if l.Len() != 1 {
		t.Fatalf("expected 1 remaining, got %d", l.Len())
	}
}

func TestSuppress_UnknownKey(t *testing.T) {
	l := New()
	if l.IsSuppressed("udp:5353") {
		t.Fatal("unknown key should not be suppressed")
	}
}

func TestSuppress_All(t *testing.T) {
	l := New()
	l.Suppress("tcp:80", time.Hour)
	l.Suppress("tcp:443", time.Hour)

	all := l.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
