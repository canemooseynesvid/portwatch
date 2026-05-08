package metrics

import (
	"sync"
	"time"
)

// ScanMetrics holds counters captured during a single scan cycle.
type ScanMetrics struct {
	PortsObserved  int
	PortsAdded     int
	PortsRemoved   int
	AlertsEmitted  int
	ScanDuration   time.Duration
}

// Collector accumulates scan-level metrics into a Registry.
type Collector struct {
	mu       sync.Mutex
	reg      *Registry
	scanTotal    *Counter
	portsAdded   *Counter
	portsRemoved *Counter
	portsObs     *Counter
	alertsEmit   *Counter
	scanDurMs    *Counter
}

// NewCollector creates a Collector backed by the given Registry.
func NewCollector(reg *Registry) *Collector {
	return &Collector{
		reg:          reg,
		scanTotal:    reg.Counter("scans_total"),
		portsAdded:   reg.Counter("ports_added_total"),
		portsRemoved: reg.Counter("ports_removed_total"),
		portsObs:     reg.Counter("ports_observed_total"),
		alertsEmit:   reg.Counter("alerts_emitted_total"),
		scanDurMs:    reg.Counter("scan_duration_ms_total"),
	}
}

// Record ingests the metrics from one scan cycle.
func (c *Collector) Record(m ScanMetrics) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.scanTotal.Inc()
	c.portsAdded.Add(int64(m.PortsAdded))
	c.portsRemoved.Add(int64(m.PortsRemoved))
	c.portsObs.Add(int64(m.PortsObserved))
	c.alertsEmit.Add(int64(m.AlertsEmitted))
	c.scanDurMs.Add(int64(m.ScanDuration.Milliseconds()))
}

// Snapshot returns a point-in-time copy of the underlying registry.
func (c *Collector) Snapshot() map[string]int64 {
	return c.reg.Snapshot()
}
