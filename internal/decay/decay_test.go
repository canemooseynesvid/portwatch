package decay

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestRecord_AddsInitialScore(t *testing.T) {
	s := NewWithClock(time.Second, fixedClock(epoch))
	score := s.Record("tcp:8080", 10.0)
	if score != 10.0 {
		t.Fatalf("expected 10.0, got %f", score)
	}
}

func TestRecord_DecaysBeforeAdding(t *testing.T) {
	now := epoch
	s := NewWithClock(time.Second, func() time.Time { return now })

	s.Record("tcp:8080", 16.0)

	// Advance one half-life; score should halve to 8, then add 0.
	now = epoch.Add(time.Second)
	score := s.Record("tcp:8080", 0)
	if score < 7.9 || score > 8.1 {
		t.Fatalf("expected ~8.0 after one half-life, got %f", score)
	}
}

func TestScore_ReturnsDecayedValueWithoutMutating(t *testing.T) {
	now := epoch
	s := NewWithClock(time.Second, func() time.Time { return now })
	s.Record("tcp:9090", 32.0)

	now = epoch.Add(2 * time.Second) // two half-lives → 32 * 0.25 = 8
	score := s.Score("tcp:9090")
	if score < 7.9 || score > 8.1 {
		t.Fatalf("expected ~8.0 after two half-lives, got %f", score)
	}

	// Calling Score again at the same time should return the same value.
	score2 := s.Score("tcp:9090")
	if score != score2 {
		t.Fatalf("Score mutated state: %f vs %f", score, score2)
	}
}

func TestScore_MissingKeyReturnsZero(t *testing.T) {
	s := New(time.Second)
	if got := s.Score("udp:53"); got != 0 {
		t.Fatalf("expected 0 for unknown key, got %f", got)
	}
}

func TestReset_ClearsScore(t *testing.T) {
	s := NewWithClock(time.Second, fixedClock(epoch))
	s.Record("tcp:443", 50.0)
	s.Reset("tcp:443")
	if got := s.Score("tcp:443"); got != 0 {
		t.Fatalf("expected 0 after reset, got %f", got)
	}
}

func TestRecord_SeparateKeysAreIndependent(t *testing.T) {
	s := NewWithClock(time.Second, fixedClock(epoch))
	s.Record("tcp:80", 10.0)
	s.Record("udp:53", 20.0)

	if got := s.Score("tcp:80"); got < 9.9 || got > 10.1 {
		t.Fatalf("tcp:80 score unexpected: %f", got)
	}
	if got := s.Score("udp:53"); got < 19.9 || got > 20.1 {
		t.Fatalf("udp:53 score unexpected: %f", got)
	}
}
