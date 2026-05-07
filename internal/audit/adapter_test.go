package audit_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/audit"
)

func makeAlert(level alerting.Level, port uint16) alerting.Alert {
	return alerting.Alert{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   "test alert",
		Port:      port,
		Protocol:  "tcp",
	}
}

func TestAlertAdapter_RecordsAlert(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)
	adapter := audit.NewAlertAdapter(l)

	al := makeAlert(alerting.LevelWarn, 9090)
	adapter.Handle(al)

	var ev audit.Event
	if err := json.Unmarshal(buf.Bytes(), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Port != 9090 {
		t.Errorf("port: got %d, want 9090", ev.Port)
	}
	if ev.Level != audit.LevelWarn {
		t.Errorf("level: got %s, want WARN", ev.Level)
	}
}

func TestAlertAdapter_LevelMapping(t *testing.T) {
	tests := []struct {
		alertLevel alerting.Level
		wantLevel  audit.Level
	}{
		{alerting.LevelInfo, audit.LevelInfo},
		{alerting.LevelWarn, audit.LevelWarn},
		{alerting.LevelCritical, audit.LevelAlert},
	}
	for _, tc := range tests {
		t.Run(string(tc.alertLevel), func(t *testing.T) {
			var buf bytes.Buffer
			adapter := audit.NewAlertAdapter(audit.NewWithWriter(&buf))
			adapter.Handle(makeAlert(tc.alertLevel, 1234))

			var ev audit.Event
			if err := json.Unmarshal(buf.Bytes(), &ev); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if ev.Level != tc.wantLevel {
				t.Errorf("level: got %s, want %s", ev.Level, tc.wantLevel)
			}
		})
	}
}
