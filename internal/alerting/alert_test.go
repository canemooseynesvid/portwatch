package alerting

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestAlertLevelString(t *testing.T) {
	cases := []struct {
		level    AlertLevel
		expected string
	}{
		{AlertInfo, "INFO"},
		{AlertWarn, "WARN"},
		{AlertCritical, "CRITICAL"},
		{AlertLevel(99), "UNKNOWN"},
	}
	for _, tc := range cases {
		if tc.level.String() != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, tc.level.String())
		}
	}
}

func TestAlertString(t *testing.T) {
	a := Alert{
		Level:     AlertWarn,
		Message:   "test message",
		Port:      8080,
		Protocol:  "tcp",
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	result := a.String()
	if !strings.Contains(result, "WARN") || !strings.Contains(result, "8080") {
		t.Errorf("unexpected alert string: %s", result)
	}
}

func TestAlerterSend_TimestampSet(t *testing.T) {
	h, collected := CollectorHandler()
	alerter := NewAlerter(h)
	alerter.Send(Alert{Level: AlertInfo, Message: "hello", Port: 80, Protocol: "tcp"})
	if len(*collected) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(*collected))
	}
	if (*collected)[0].Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestFilterHandler(t *testing.T) {
	h, collected := CollectorHandler()
	filtered := FilterHandler(AlertWarn, h)
	alerter := NewAlerter(filtered)

	alerter.Send(Alert{Level: AlertInfo, Port: 1, Protocol: "tcp"})
	alerter.Send(Alert{Level: AlertWarn, Port: 2, Protocol: "tcp"})
	alerter.Send(Alert{Level: AlertCritical, Port: 3, Protocol: "tcp"})

	if len(*collected) != 2 {
		t.Errorf("expected 2 alerts after filter, got %d", len(*collected))
	}
}

func TestWriterHandler(t *testing.T) {
	var buf bytes.Buffer
	h := WriterHandler(&buf)
	h(Alert{Level: AlertCritical, Message: "conflict", Port: 443, Protocol: "tcp", Timestamp: time.Now()})
	if !strings.Contains(buf.String(), "CRITICAL") {
		t.Errorf("expected CRITICAL in output, got: %s", buf.String())
	}
}
