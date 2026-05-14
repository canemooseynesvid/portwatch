package redact_test

import (
	"testing"

	"portwatch/internal/alerting"
	"portwatch/internal/redact"
)

func makeAlert(meta map[string]string) alerting.Alert {
	a := alerting.Alert{
		Level:   alerting.LevelInfo,
		Message: "test alert",
		Meta:    meta,
	}
	return a
}

func TestApply_RedactsDefaultSensitiveKeys(t *testing.T) {
	r := redact.New()
	a := makeAlert(map[string]string{
		"token":    "abc123",
		"username": "alice",
		"password": "hunter2",
	})
	out := r.Apply(a)
	if out.Meta["token"] != "[REDACTED]" {
		t.Errorf("expected token to be redacted, got %q", out.Meta["token"])
	}
	if out.Meta["password"] != "[REDACTED]" {
		t.Errorf("expected password to be redacted, got %q", out.Meta["password"])
	}
	if out.Meta["username"] != "alice" {
		t.Errorf("expected username to be preserved, got %q", out.Meta["username"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	r := redact.New()
	orig := map[string]string{"api_key": "secret-value", "port": "8080"}
	a := makeAlert(orig)
	r.Apply(a)
	if orig["api_key"] == "[REDACTED]" {
		t.Error("original meta map was mutated")
	}
}

func TestApply_EmptyMetaIsNoop(t *testing.T) {
	r := redact.New()
	a := makeAlert(nil)
	out := r.Apply(a)
	if out.Meta != nil {
		t.Errorf("expected nil meta, got %v", out.Meta)
	}
}

func TestNewWithKeys_RedactsCustomKey(t *testing.T) {
	r := redact.NewWithKeys("session_id")
	a := makeAlert(map[string]string{
		"session_id": "xyz",
		"host":       "localhost",
	})
	out := r.Apply(a)
	if out.Meta["session_id"] != "[REDACTED]" {
		t.Errorf("expected session_id to be redacted, got %q", out.Meta["session_id"])
	}
	if out.Meta["host"] != "localhost" {
		t.Errorf("expected host to be preserved, got %q", out.Meta["host"])
	}
}

func TestMiddleware_RedactsBeforeForwarding(t *testing.T) {
	r := redact.New()
	var received alerting.Alert
	next := alerting.HandlerFunc(func(a alerting.Alert) error {
		received = a
		return nil
	})
	h := r.Middleware(next)
	a := makeAlert(map[string]string{"auth": "bearer xyz", "port": "443"})
	if err := h.Handle(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Meta["auth"] != "[REDACTED]" {
		t.Errorf("expected auth redacted in forwarded alert, got %q", received.Meta["auth"])
	}
	if received.Meta["port"] != "443" {
		t.Errorf("expected port preserved, got %q", received.Meta["port"])
	}
}
