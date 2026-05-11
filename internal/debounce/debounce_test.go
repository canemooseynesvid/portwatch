package debounce

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestAllow_FirstCallPermitted(t *testing.T) {
	d := NewWithClock(5*time.Second, fixedClock(epoch))
	if !d.Allow("key1") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallWithinWindowBlocked(t *testing.T) {
	now := epoch
	d := NewWithClock(5*time.Second, func() time.Time { return now })
	d.Allow("key1")
	now = epoch.Add(2 * time.Second)
	if d.Allow("key1") {
		t.Fatal("expected second call within window to be blocked")
	}
}

func TestAllow_PermittedAfterWindowExpires(t *testing.T) {
	now := epoch
	d := NewWithClock(5*time.Second, func() time.Time { return now })
	d.Allow("key1")
	now = epoch.Add(6 * time.Second)
	if !d.Allow("key1") {
		t.Fatal("expected call after window expiry to be allowed")
	}
}

func TestAllow_SeparateKeysAreIndependent(t *testing.T) {
	d := NewWithClock(5*time.Second, fixedClock(epoch))
	d.Allow("key1")
	if !d.Allow("key2") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestReset_AllowsKeyImmediately(t *testing.T) {
	now := epoch
	d := NewWithClock(5*time.Second, func() time.Time { return now })
	d.Allow("key1")
	now = epoch.Add(1 * time.Second)
	d.Reset("key1")
	if !d.Allow("key1") {
		t.Fatal("expected key to be allowed after reset")
	}
}

func TestPurge_RemovesExpiredKeys(t *testing.T) {
	now := epoch
	d := NewWithClock(5*time.Second, func() time.Time { return now })
	d.Allow("key1")
	d.Allow("key2")
	now = epoch.Add(6 * time.Second)
	d.Allow("key3") // fresh key at new time
	d.Purge()
	if d.Len() != 1 {
		t.Fatalf("expected 1 active key after purge, got %d", d.Len())
	}
}

func TestLen_CountsOnlyActiveKeys(t *testing.T) {
	now := epoch
	d := NewWithClock(5*time.Second, func() time.Time { return now })
	d.Allow("a")
	d.Allow("b")
	if d.Len() != 2 {
		t.Fatalf("expected 2, got %d", d.Len())
	}
	now = epoch.Add(10 * time.Second)
	if d.Len() != 0 {
		t.Fatalf("expected 0 after expiry, got %d", d.Len())
	}
}
