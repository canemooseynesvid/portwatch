package notify_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"portwatch/internal/notify"
)

type fakeChannel struct {
	name   string
	calls  []string
	failOn string
}

func (f *fakeChannel) Name() string { return f.name }
func (f *fakeChannel) Send(subject, body string) error {
	if f.failOn != "" && strings.Contains(subject, f.failOn) {
		return errors.New("injected error")
	}
	f.calls = append(f.calls, subject)
	return nil
}

func TestNotify_SendsToAllChannels(t *testing.T) {
	ch1 := &fakeChannel{name: "ch1"}
	ch2 := &fakeChannel{name: "ch2"}
	n := notify.New(0, ch1, ch2)

	if err := n.Notify("key1", "subject", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ch1.calls) != 1 || len(ch2.calls) != 1 {
		t.Errorf("expected 1 call each, got %d and %d", len(ch1.calls), len(ch2.calls))
	}
}

func TestNotify_ThrottlesSameKey(t *testing.T) {
	ch := &fakeChannel{name: "ch"}
	n := notify.New(10*time.Minute, ch)

	_ = n.Notify("key", "first", "body")
	_ = n.Notify("key", "second", "body")

	if len(ch.calls) != 1 {
		t.Errorf("expected 1 call due to throttle, got %d", len(ch.calls))
	}
}

func TestNotify_ResetClearsThrottle(t *testing.T) {
	ch := &fakeChannel{name: "ch"}
	n := notify.New(10*time.Minute, ch)

	_ = n.Notify("key", "first", "body")
	n.Reset("key")
	_ = n.Notify("key", "second", "body")

	if len(ch.calls) != 2 {
		t.Errorf("expected 2 calls after reset, got %d", len(ch.calls))
	}
}

func TestNotify_ContinuesOnChannelError(t *testing.T) {
	ch1 := &fakeChannel{name: "ch1", failOn: "alert"}
	ch2 := &fakeChannel{name: "ch2"}
	n := notify.New(0, ch1, ch2)

	err := n.Notify("k", "alert", "body")
	if err == nil {
		t.Error("expected error from failing channel")
	}
	if len(ch2.calls) != 1 {
		t.Errorf("ch2 should still receive notification, got %d calls", len(ch2.calls))
	}
}

func TestWriterChannel_Send(t *testing.T) {
	var buf bytes.Buffer
	ch := notify.NewWriterChannel("buf", &buf)

	if err := ch.Send("test subject", "test body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "test subject") {
		t.Errorf("expected subject in output, got: %s", buf.String())
	}
}
