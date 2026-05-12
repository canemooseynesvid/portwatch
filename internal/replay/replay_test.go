package replay_test

import (
	"context"
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/replay"
	"portwatch/internal/snapshot"
)

func makeHistory(t *testing.T, n int) *snapshot.History {
	t.Helper()
	h := snapshot.NewHistory(20)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		e := snapshot.Entry{Port: uint16(8000 + i), Protocol: "tcp"}
		h.Record(snapshot.Event{
			Kind:  snapshot.EventAdded,
			Entry: e,
			At:    base.Add(time.Duration(i) * time.Second),
		})
	}
	return h
}

func TestReplayer_DispatchesAllEvents(t *testing.T) {
	h := makeHistory(t, 5)
	r := replay.New(h, replay.DefaultOptions())

	var got []alerting.Alert
	n, err := r.Run(context.Background(), func(a alerting.Alert) error {
		got = append(got, a)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("dispatched = %d, want 5", n)
	}
	if len(got) != 5 {
		t.Errorf("received %d alerts, want 5", len(got))
	}
}

func TestReplayer_EmptyHistory(t *testing.T) {
	h := snapshot.NewHistory(10)
	r := replay.New(h, replay.DefaultOptions())

	n, err := r.Run(context.Background(), func(a alerting.Alert) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 dispatched, got %d", n)
	}
}

func TestReplayer_FilterSkipsEvents(t *testing.T) {
	h := makeHistory(t, 6)
	opts := replay.DefaultOptions()
	// Only pass even-indexed ports (8000, 8002, 8004)
	opts.Filter = func(a alerting.Alert) bool {
		return a.Timestamp.Second()%2 == 0
	}
	r := replay.New(h, opts)

	n, err := r.Run(context.Background(), func(a alerting.Alert) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("dispatched = %d, want 3", n)
	}
}

func TestReplayer_ContextCancellation(t *testing.T) {
	h := makeHistory(t, 10)
	ctx, cancel := context.WithCancel(context.Background())

	called := 0
	_, err := replay.New(h, replay.DefaultOptions()).Run(ctx, func(a alerting.Alert) error {
		called++
		if called == 3 {
			cancel()
		}
		return nil
	})

	if err == nil {
		t.Error("expected context cancellation error")
	}
}
