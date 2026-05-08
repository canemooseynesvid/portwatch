package metrics

import (
	"testing"
	"time"
)

func TestCollector_RecordIncrementsScanTotal(t *testing.T) {
	reg := New()
	c := NewCollector(reg)

	c.Record(ScanMetrics{})
	c.Record(ScanMetrics{})

	snap := c.Snapshot()
	if snap["scans_total"] != 2 {
		t.Errorf("scans_total: want 2, got %d", snap["scans_total"])
	}
}

func TestCollector_RecordPortCounters(t *testing.T) {
	reg := New()
	c := NewCollector(reg)

	c.Record(ScanMetrics{
		PortsObserved: 10,
		PortsAdded:    3,
		PortsRemoved:  1,
		AlertsEmitted: 2,
	})

	snap := c.Snapshot()

	cases := map[string]int64{
		"ports_observed_total": 10,
		"ports_added_total":    3,
		"ports_removed_total":  1,
		"alerts_emitted_total": 2,
	}
	for key, want := range cases {
		if snap[key] != want {
			t.Errorf("%s: want %d, got %d", key, want, snap[key])
		}
	}
}

func TestCollector_RecordAccumulatesAcrossScans(t *testing.T) {
	reg := New()
	c := NewCollector(reg)

	c.Record(ScanMetrics{PortsAdded: 2})
	c.Record(ScanMetrics{PortsAdded: 5})

	snap := c.Snapshot()
	if snap["ports_added_total"] != 7 {
		t.Errorf("ports_added_total: want 7, got %d", snap["ports_added_total"])
	}
}

func TestCollector_RecordScanDuration(t *testing.T) {
	reg := New()
	c := NewCollector(reg)

	c.Record(ScanMetrics{ScanDuration: 250 * time.Millisecond})
	c.Record(ScanMetrics{ScanDuration: 100 * time.Millisecond})

	snap := c.Snapshot()
	if snap["scan_duration_ms_total"] != 350 {
		t.Errorf("scan_duration_ms_total: want 350, got %d", snap["scan_duration_ms_total"])
	}
}

func TestCollector_SnapshotIsolation(t *testing.T) {
	reg := New()
	c := NewCollector(reg)
	c.Record(ScanMetrics{PortsAdded: 1})

	s1 := c.Snapshot()
	c.Record(ScanMetrics{PortsAdded: 1})
	s2 := c.Snapshot()

	if s1["ports_added_total"] == s2["ports_added_total"] {
		t.Error("expected snapshots to differ after additional record")
	}
}
