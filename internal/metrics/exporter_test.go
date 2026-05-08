package metrics

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestExportJSON_ContainsCounters(t *testing.T) {
	reg := New()
	reg.Counter("alerts.sent").Add(5)
	reg.Counter("ports.scanned").Add(42)

	var buf bytes.Buffer
	ex := NewExporter(reg, &buf)
	if err := ex.ExportJSON(); err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}

	var snap ExportSnapshot
	if err := json.Unmarshal(buf.Bytes(), &snap); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if snap.Counters["alerts.sent"] != 5 {
		t.Errorf("expected alerts.sent=5, got %d", snap.Counters["alerts.sent"])
	}
	if snap.Counters["ports.scanned"] != 42 {
		t.Errorf("expected ports.scanned=42, got %d", snap.Counters["ports.scanned"])
	}
}

func TestExportJSON_TimestampPresent(t *testing.T) {
	reg := New()
	var buf bytes.Buffer
	ex := NewExporter(reg, &buf)
	_ = ex.ExportJSON()

	if !strings.Contains(buf.String(), "timestamp") {
		t.Error("expected 'timestamp' field in JSON output")
	}
}

func TestExportText_ContainsHeaders(t *testing.T) {
	reg := New()
	reg.Counter("scan.cycles").Inc()

	var buf bytes.Buffer
	ex := NewExporter(reg, &buf)
	if err := ex.ExportText(); err != nil {
		t.Fatalf("ExportText error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"COUNTER", "VALUE", "scan.cycles", "1"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in text output:\n%s", want, out)
		}
	}
}

func TestExportText_SortedKeys(t *testing.T) {
	reg := New()
	reg.Counter("z.last").Inc()
	reg.Counter("a.first").Inc()
	reg.Counter("m.middle").Inc()

	var buf bytes.Buffer
	ex := NewExporter(reg, &buf)
	_ = ex.ExportText()

	out := buf.String()
	posA := strings.Index(out, "a.first")
	posM := strings.Index(out, "m.middle")
	posZ := strings.Index(out, "z.last")

	if posA > posM || posM > posZ {
		t.Errorf("expected sorted output: a < m < z, got positions %d %d %d", posA, posM, posZ)
	}
}

func TestNewExporter_DefaultsToStdout(t *testing.T) {
	reg := New()
	ex := NewExporter(reg, nil)
	if ex.out == nil {
		t.Error("expected non-nil writer when nil passed")
	}
}
