package metrics

import (
	"bytes"
	"strings"
	"testing"
)

func TestCounter_IncAndValue(t *testing.T) {
	c := &Counter{}
	if c.Value() != 0 {
		t.Fatalf("expected 0, got %d", c.Value())
	}
	c.Inc()
	c.Inc()
	if c.Value() != 2 {
		t.Fatalf("expected 2, got %d", c.Value())
	}
}

func TestCounter_Add(t *testing.T) {
	c := &Counter{}
	c.Add(10)
	if c.Value() != 10 {
		t.Fatalf("expected 10, got %d", c.Value())
	}
}

func TestRegistry_CounterReuse(t *testing.T) {
	r := New()
	a := r.Counter("scan.total")
	b := r.Counter("scan.total")
	if a != b {
		t.Fatal("expected same counter instance for same name")
	}
}

func TestRegistry_Snapshot(t *testing.T) {
	r := New()
	r.Counter("alerts.sent").Add(3)
	r.Counter("ports.new").Inc()

	snap := r.Snapshot()
	if snap["alerts.sent"] != 3 {
		t.Errorf("alerts.sent: expected 3, got %d", snap["alerts.sent"])
	}
	if snap["ports.new"] != 1 {
		t.Errorf("ports.new: expected 1, got %d", snap["ports.new"])
	}
}

func TestRegistry_SnapshotIsCopy(t *testing.T) {
	r := New()
	r.Counter("x").Inc()
	snap := r.Snapshot()
	snap["x"] = 999
	if r.Counter("x").Value() != 1 {
		t.Fatal("snapshot mutation affected registry")
	}
}

func TestRegistry_Print_ContainsUptime(t *testing.T) {
	r := New()
	r.Counter("scan.errors").Add(2)
	var buf bytes.Buffer
	r.Print(&buf)
	out := buf.String()
	if !strings.Contains(out, "uptime:") {
		t.Errorf("expected uptime in output, got: %s", out)
	}
	if !strings.Contains(out, "scan.errors") {
		t.Errorf("expected counter name in output, got: %s", out)
	}
}

func TestRegistry_Print_EmptyCounters(t *testing.T) {
	r := New()
	var buf bytes.Buffer
	r.Print(&buf)
	if !strings.Contains(buf.String(), "no counters") {
		t.Errorf("expected empty notice, got: %s", buf.String())
	}
}
