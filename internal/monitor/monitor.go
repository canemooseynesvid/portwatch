package monitor

import (
	"context"
	"log"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/portscanner"
)

// Config holds configuration for the Monitor.
type Config struct {
	Interval    time.Duration
	AllowedPorts map[uint16]bool
}

// Monitor periodically scans ports and emits alerts on unexpected bindings.
type Monitor struct {
	scanner  *portscanner.Scanner
	alerter  *alerting.Alerter
	config   Config
	previous map[string]portscanner.PortEntry
}

// New creates a new Monitor.
func New(s *portscanner.Scanner, a *alerting.Alerter, cfg Config) *Monitor {
	return &Monitor{
		scanner:  s,
		alerter:  a,
		config:   cfg,
		previous: make(map[string]portscanner.PortEntry),
	}
}

// Run starts the monitoring loop until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("monitor: shutting down")
			return
		case <-ticker.C:
			if err := m.scan(); err != nil {
				log.Printf("monitor: scan error: %v", err)
			}
		}
	}
}

func (m *Monitor) scan() error {
	entries, err := m.scanner.Scan()
	if err != nil {
		return err
	}

	current := make(map[string]portscanner.PortEntry, len(entries))
	for _, e := range entries {
		key := entryKey(e)
		current[key] = e

		if _, seen := m.previous[key]; !seen {
			m.handleNew(e)
		}
	}

	for key, e := range m.previous {
		if _, still := current[key]; !still {
			m.handleClosed(e)
		}
	}

	m.previous = current
	return nil
}

func (m *Monitor) handleNew(e portscanner.PortEntry) {
	level := alerting.Info
	if !m.config.AllowedPorts[e.LocalPort] {
		level = alerting.Warning
	}
	alert := alerting.NewPortBindAlert(e.LocalPort, e.Protocol, level)
	m.alerter.Send(alert)
}

func (m *Monitor) handleClosed(e portscanner.PortEntry) {
	alert := alerting.NewPortClosedAlert(e.LocalPort, e.Protocol)
	m.alerter.Send(alert)
}

func entryKey(e portscanner.PortEntry) string {
	return e.Protocol + ":" + string(rune(e.LocalPort))
}
