package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/audit"
)

func TestRecord_WritesJSON(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)

	ev := audit.Event{
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Level:     audit.LevelAlert,
		Message:   "unexpected binding",
		Port:      8080,
		Protocol:  "tcp",
	}
	if err := l.Record(ev); err != nil {
		t.Fatalf("Record: %v", err)
	}

	var got audit.Event
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Port != 8080 {
		t.Errorf("port: got %d, want 8080", got.Port)
	}
	if got.Level != audit.LevelAlert {
		t.Errorf("level: got %s, want ALERT", got.Level)
	}
}

func TestRecord_SetsTimestampWhenZero(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)

	if err := l.Record(audit.Event{Level: audit.LevelInfo, Message: "hello"}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	var got audit.Event
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestConvenienceMethods(t *testing.T) {
	tests := []struct {
		name  string
		fn    func(*audit.Logger) error
		want  string
	}{
		{"Info", func(l *audit.Logger) error { return l.Info("msg") }, "INFO"},
		{"Warn", func(l *audit.Logger) error { return l.Warn("msg") }, "WARN"},
		{"Alert", func(l *audit.Logger) error { return l.Alert("msg") }, "ALERT"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := audit.NewWithWriter(&buf)
			if err := tc.fn(l); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("output %q does not contain level %s", buf.String(), tc.want)
			}
		})
	}
}

func TestNew_FallsBackToStderr(t *testing.T) {
	l, err := audit.New("")
	if err != nil {
		t.Fatalf("New with empty path: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}
