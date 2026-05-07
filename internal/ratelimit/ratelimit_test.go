package ratelimit_test

import (
	"testing"
	"time"

	"portwatch/internal/ratelimit"
)

func TestAllow_WithinLimit(t *testing.T) {
	l := ratelimit.New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("tcp:8080") {
			t.Fatalf("expected Allow to return true on call %d", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	l := ratelimit.New(2, time.Minute)
	l.Allow("tcp:9000")
	l.Allow("tcp:9000")
	if l.Allow("tcp:9000") {
		t.Fatal("expected Allow to return false after limit exceeded")
	}
}

func TestAllow_SeparateKeys(t *testing.T) {
	l := ratelimit.New(1, time.Minute)
	if !l.Allow("tcp:80") {
		t.Fatal("expected true for tcp:80")
	}
	if !l.Allow("udp:53") {
		t.Fatal("expected true for udp:53 (separate key)")
	}
	if l.Allow("tcp:80") {
		t.Fatal("expected false for tcp:80 after limit")
	}
}

func TestAllow_WindowExpiry(t *testing.T) {
	l := ratelimit.New(1, 50*time.Millisecond)
	if !l.Allow("tcp:443") {
		t.Fatal("expected true on first call")
	}
	if l.Allow("tcp:443") {
		t.Fatal("expected false within window")
	}
	time.Sleep(60 * time.Millisecond)
	if !l.Allow("tcp:443") {
		t.Fatal("expected true after window expired")
	}
}

func TestReset_ClearsState(t *testing.T) {
	l := ratelimit.New(1, time.Minute)
	l.Allow("tcp:22")
	l.Reset("tcp:22")
	if !l.Allow("tcp:22") {
		t.Fatal("expected true after reset")
	}
}

func TestPrune_RemovesExpired(t *testing.T) {
	l := ratelimit.New(1, 30*time.Millisecond)
	l.Allow("tcp:8888")
	time.Sleep(40 * time.Millisecond)
	// Should not panic and should remove the expired bucket
	l.Prune()
	// After prune, the key should be treated as fresh
	if !l.Allow("tcp:8888") {
		t.Fatal("expected true after prune cleared expired bucket")
	}
}

func TestNew_ZeroMaxDefaultsToOne(t *testing.T) {
	l := ratelimit.New(0, time.Minute)
	if !l.Allow("key") {
		t.Fatal("expected first call to be allowed")
	}
	if l.Allow("key") {
		t.Fatal("expected second call to be denied with max=1")
	}
}
