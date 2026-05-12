package pipeline_test

import (
	"context"
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/pipeline"
)

func makeAlert(level alerting.AlertLevel, tags ...string) alerting.Alert {
	return alerting.Alert{
		Level:     level,
		Message:   "test alert",
		Timestamp: time.Now(),
		Tags:      tags,
	}
}

func TestPipeline_PassesAllStages(t *testing.T) {
	var received []alerting.Alert
	sink := func(a alerting.Alert) { received = append(received, a) }

	p := pipeline.New(sink, pipeline.AlwaysAllow(), pipeline.AlwaysAllow())
	p.Process(makeAlert(alerting.LevelWarn))

	if len(received) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(received))
	}
}

func TestPipeline_DropsOnFailedStage(t *testing.T) {
	var received []alerting.Alert
	sink := func(a alerting.Alert) { received = append(received, a) }

	drop := func(_ alerting.Alert) bool { return false }
	p := pipeline.New(sink, pipeline.AlwaysAllow(), drop)
	p.Process(makeAlert(alerting.LevelWarn))

	if len(received) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(received))
	}
}

func TestMinLevel_FiltersLow(t *testing.T) {
	var received []alerting.Alert
	sink := func(a alerting.Alert) { received = append(received, a) }

	p := pipeline.New(sink, pipeline.MinLevel(alerting.LevelError))
	p.Process(makeAlert(alerting.LevelInfo))
	p.Process(makeAlert(alerting.LevelWarn))
	p.Process(makeAlert(alerting.LevelError))

	if len(received) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(received))
	}
	if received[0].Level != alerting.LevelError {
		t.Errorf("expected LevelError, got %v", received[0].Level)
	}
}

func TestWithTag_FiltersUntagged(t *testing.T) {
	var received []alerting.Alert
	sink := func(a alerting.Alert) { received = append(received, a) }

	p := pipeline.New(sink, pipeline.WithTag("critical"))
	p.Process(makeAlert(alerting.LevelWarn, "info"))
	p.Process(makeAlert(alerting.LevelWarn, "critical", "extra"))

	if len(received) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(received))
	}
}

func TestProcessCtx_RespectsCancel(t *testing.T) {
	var received []alerting.Alert
	sink := func(a alerting.Alert) { received = append(received, a) }

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := pipeline.New(sink)
	p.ProcessCtx(ctx, makeAlert(alerting.LevelWarn))

	if len(received) != 0 {
		t.Fatalf("expected 0 alerts after cancel, got %d", len(received))
	}
}
