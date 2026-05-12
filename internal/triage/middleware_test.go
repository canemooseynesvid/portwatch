package triage_test

import (
	"errors"
	"testing"

	"portwatch/internal/alerting"
	"portwatch/internal/triage"
)

type collectHandler struct {
	got []alerting.Alert
	err error
}

func (c *collectHandler) Handle(a alerting.Alert) error {
	c.got = append(c.got, a)
	return c.err
}

func TestMiddleware_ForwardsAlert(t *testing.T) {
	tr := triage.New(nil)
	collector := &collectHandler{}
	mw := triage.NewMiddleware(tr, collector)

	a := makeAlert(alerting.Error, "test")
	if err := mw.Handle(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(collector.got) != 1 {
		t.Fatalf("expected 1 forwarded alert, got %d", len(collector.got))
	}
	if tr.Len() != 1 {
		t.Errorf("expected 1 pending alert in triager, got %d", tr.Len())
	}
}

func TestMiddleware_PropagatesError(t *testing.T) {
	tr := triage.New(nil)
	collector := &collectHandler{err: errors.New("downstream failure")}
	mw := triage.NewMiddleware(tr, collector)

	if err := mw.Handle(makeAlert(alerting.Warning, "x")); err == nil {
		t.Error("expected error from downstream handler")
	}
}

func TestMiddleware_NilNext(t *testing.T) {
	tr := triage.New(nil)
	mw := triage.NewMiddleware(tr, nil)

	if err := mw.Handle(makeAlert(alerting.Info, "no-op")); err != nil {
		t.Errorf("unexpected error with nil next: %v", err)
	}
	if tr.Len() != 1 {
		t.Errorf("expected alert recorded in triager, got %d", tr.Len())
	}
}
