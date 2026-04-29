package snapshot_test

import (
	"testing"
	"time"

	"portwatch/internal/snapshot"
)

func makeEvent(kind snapshot.EventKind, port uint16) snapshot.Event {
	return snapshot.Event{
		Kind:       kind,
		Entry:      entry("tcp", "0.0.0.0", port),
		OccurredAt: time.Now(),
	}
}

func TestHistory_RecordAndLen(t *testing.T) {
	h := snapshot.NewHistory(10)
	h.Record(makeEvent(snapshot.EventBound, 8080))
	h.Record(makeEvent(snapshot.EventBound, 9090))
	if h.Len() != 2 {
		t.Errorf("expected len 2, got %d", h.Len())
	}
}

func TestHistory_Eviction(t *testing.T) {
	h := snapshot.NewHistory(3)
	for i := uint16(1); i <= 5; i++ {
		h.Record(makeEvent(snapshot.EventBound, i))
	}
	if h.Len() != 3 {
		t.Errorf("expected len 3 after eviction, got %d", h.Len())
	}
	events := h.Recent()
	if events[0].Entry.Port != 3 {
		t.Errorf("expected oldest retained port 3, got %d", events[0].Entry.Port)
	}
}

func TestHistory_RecentReturnsCopy(t *testing.T) {
	h := snapshot.NewHistory(10)
	h.Record(makeEvent(snapshot.EventReleased, 443))
	a := h.Recent()
	a[0].Entry.Port = 9999
	b := h.Recent()
	if b[0].Entry.Port == 9999 {
		t.Error("Recent() should return a copy, not a reference")
	}
}

func TestEventKind_String(t *testing.T) {
	if snapshot.EventBound.String() != "bound" {
		t.Errorf("unexpected string for EventBound")
	}
	if snapshot.EventReleased.String() != "released" {
		t.Errorf("unexpected string for EventReleased")
	}
}

func TestHistory_DefaultCapacity(t *testing.T) {
	h := snapshot.NewHistory(0)
	if h == nil {
		t.Fatal("expected non-nil history")
	}
	h.Record(makeEvent(snapshot.EventBound, 80))
	if h.Len() != 1 {
		t.Errorf("expected len 1, got %d", h.Len())
	}
}
