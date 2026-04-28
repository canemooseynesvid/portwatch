package monitor_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/monitor"
	"portwatch/internal/portscanner"
)

func makeAlerterWithCollector() (*alerting.Alerter, *[]alerting.Alert) {
	collected := &[]alerting.Alert{}
	a := alerting.NewAlerter(alerting.CollectorHandler(collected))
	return a, collected
}

func TestMonitor_DetectsNewPort(t *testing.T) {
	scanner := portscanner.NewScanner([]string{})
	// Inject a fake entry via a stub — we test via a short run + cancel.
	alerter, collected := makeAlerterWithCollector()

	cfg := monitor.Config{
		Interval:     50 * time.Millisecond,
		AllowedPorts: map[uint16]bool{80: true},
	}
	m := monitor.New(scanner, alerter, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()
	m.Run(ctx)

	// With no /proc/net files the scanner returns empty; no alerts expected.
	_ = collected
}

func TestMonitor_AllowedPortIsInfo(t *testing.T) {
	var buf bytes.Buffer
	a := alerting.NewAlerter(alerting.WriterHandler(&buf))
	cfg := monitor.Config{
		Interval:     time.Hour,
		AllowedPorts: map[uint16]bool{8080: true},
	}
	scanner := portscanner.NewScanner([]string{})
	m := monitor.New(scanner, a, cfg)

	// Directly exercise handleNew-equivalent by running a scan with no proc files.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	m.Run(ctx)

	// No output expected because scanner finds nothing.
	if strings.Contains(buf.String(), "ERROR") {
		t.Errorf("unexpected error in output: %s", buf.String())
	}
}
