package throttle_test

import (
	"testing"
	"time"

	"portwatch/internal/throttle"
)

func fixedClock(t time.Time) throttle.Clock {
	return func() time.Time { return t }
}

func TestAllow_WithinBurst(t *testing.T) {
	now := time.Now()
	th := throttle.NewWithClock(3, time.Minute, fixedClock(now))

	for i := 0; i < 3; i++ {
		if !th.Allow("key") {
			t.Fatalf("expected Allow to return true on call %d", i+1)
		}
	}
}

func TestAllow_ExceedsBurst(t *testing.T) {
	now := time.Now()
	th := throttle.NewWithClock(2, time.Minute, fixedClock(now))

	th.Allow("key")
	th.Allow("key")
	if th.Allow("key") {
		t.Fatal("expected Allow to return false after burst exhausted")
	}
}

func TestAllow_WindowRefreshesTokens(t *testing.T) {
	base := time.Now()
	clock := base
	th := throttle.NewWithClock(1, time.Second, func() time.Time { return clock })

	th.Allow("key") // consume token
	if th.Allow("key") {
		t.Fatal("expected false while still in window")
	}

	clock = base.Add(2 * time.Second) // advance past window
	if !th.Allow("key") {
		t.Fatal("expected true after window expiry")
	}
}

func TestAllow_SeparateKeys(t *testing.T) {
	now := time.Now()
	th := throttle.NewWithClock(1, time.Minute, fixedClock(now))

	th.Allow("a")
	if !th.Allow("b") {
		t.Fatal("key b should have its own bucket")
	}
}

func TestReset_ClearsKey(t *testing.T) {
	now := time.Now()
	th := throttle.NewWithClock(1, time.Minute, fixedClock(now))

	th.Allow("key")
	if th.Allow("key") {
		t.Fatal("expected false before reset")
	}
	th.Reset("key")
	if !th.Allow("key") {
		t.Fatal("expected true after reset")
	}
}

func TestPurge_RemovesExpired(t *testing.T) {
	base := time.Now()
	clock := base
	th := throttle.NewWithClock(2, time.Second, func() time.Time { return clock })

	th.Allow("x")
	th.Allow("y")
	if th.Len() != 2 {
		t.Fatalf("expected 2 keys, got %d", th.Len())
	}

	clock = base.Add(5 * time.Second)
	th.Purge()
	if th.Len() != 0 {
		t.Fatalf("expected 0 keys after purge, got %d", th.Len())
	}
}

func TestPurge_RetainsActiveKeys(t *testing.T) {
	base := time.Now()
	clock := base
	th := throttle.NewWithClock(2, time.Second, func() time.Time { return clock })

	th.Allow("active")
	th.Allow("expired")

	// Advance time so only keys whose window started before (clock - window) are expired.
	// "active" gets a fresh Allow call after the clock advances, keeping it current.
	clock = base.Add(5 * time.Second)
	th.Allow("active") // refresh the window for "active"

	th.Purge()
	if th.Len() != 1 {
		t.Fatalf("expected 1 active key after purge, got %d", th.Len())
	}
}
