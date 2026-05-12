package envelope_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/envelope"
)

func makeAlert(level alerting.Level) alerting.Alert {
	return alerting.Alert{
		Level:     level,
		Message:   "test alert",
		Timestamp: time.Now(),
	}
}

func TestNew_SetsFields(t *testing.T) {
	a := makeAlert(alerting.LevelWarning)
	env := envelope.New(a, "slack", "pagerduty")

	if env.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if env.Priority != 2 {
		t.Errorf("expected priority 2 for Warning, got %d", env.Priority)
	}
	if !env.HasTag("slack") {
		t.Error("expected HasTag(slack) to be true")
	}
	if env.HasTag("email") {
		t.Error("expected HasTag(email) to be false")
	}
}

func TestNew_PriorityMapping(t *testing.T) {
	cases := []struct {
		level    alerting.Level
		wantPri  int
	}{
		{alerting.LevelInfo, 1},
		{alerting.LevelWarning, 2},
		{alerting.LevelError, 3},
		{alerting.LevelCritical, 4},
	}
	for _, tc := range cases {
		env := envelope.New(makeAlert(tc.level))
		if env.Priority != tc.wantPri {
			t.Errorf("level %v: want priority %d, got %d", tc.level, tc.wantPri, env.Priority)
		}
	}
}

func TestEnvelope_String_ContainsID(t *testing.T) {
	env := envelope.New(makeAlert(alerting.LevelInfo), "ops")
	s := env.String()
	if !strings.Contains(s, env.ID[:8]) {
		t.Errorf("String() missing short ID, got: %s", s)
	}
}

func TestRouter_DispatchesToMatchingTag(t *testing.T) {
	router := envelope.NewRouter()
	var got string
	router.Register("slack", func(e envelope.Envelope) error {
		got = "slack"
		return nil
	})
	env := envelope.New(makeAlert(alerting.LevelInfo), "slack")
	if err := router.Dispatch(env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "slack" {
		t.Errorf("expected slack handler to be called")
	}
}

func TestRouter_FallsBackToDefault(t *testing.T) {
	router := envelope.NewRouter()
	var called bool
	router.SetDefault(func(e envelope.Envelope) error {
		called = true
		return nil
	})
	env := envelope.New(makeAlert(alerting.LevelInfo), "unknown")
	if err := router.Dispatch(env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected default handler to be called")
	}
}

func TestRouter_ErrorWhenNoMatch(t *testing.T) {
	router := envelope.NewRouter()
	env := envelope.New(makeAlert(alerting.LevelInfo), "nowhere")
	err := router.Dispatch(env)
	if err == nil {
		t.Fatal("expected error when no handler matches")
	}
	if !strings.Contains(err.Error(), "no matching route") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRouter_HandlerErrorPropagates(t *testing.T) {
	router := envelope.NewRouter()
	sentinel := errors.New("handler failed")
	router.Register("email", func(e envelope.Envelope) error {
		return sentinel
	})
	env := envelope.New(makeAlert(alerting.LevelError), "email")
	if err := router.Dispatch(env); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
