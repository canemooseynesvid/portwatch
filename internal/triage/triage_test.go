package triage_test

import (
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/triage"
)

func makeAlert(level alerting.Level, msg string) alerting.Alert {
	return alerting.Alert{
		Level:     level,
		Message:   msg,
		Timestamp: time.Now(),
	}
}

func TestTriager_RankedOrder(t *testing.T) {
	tr := triage.New(triage.DefaultWeights())

	tr.Add(makeAlert(alerting.Info, "info alert"))
	tr.Add(makeAlert(alerting.Critical, "critical alert"))
	tr.Add(makeAlert(alerting.Warning, "warning alert"))

	ranked := tr.Ranked()
	if len(ranked) != 3 {
		t.Fatalf("expected 3 scored alerts, got %d", len(ranked))
	}
	if ranked[0].Alert.Message != "critical alert" {
		t.Errorf("expected critical first, got %q", ranked[0].Alert.Message)
	}
	if ranked[2].Alert.Message != "info alert" {
		t.Errorf("expected info last, got %q", ranked[2].Alert.Message)
	}
}

func TestTriager_RankedClearsQueue(t *testing.T) {
	tr := triage.New(nil)
	tr.Add(makeAlert(alerting.Warning, "w"))

	_ = tr.Ranked()
	if tr.Len() != 0 {
		t.Errorf("expected empty queue after Ranked(), got %d", tr.Len())
	}
}

func TestTriager_CustomWeights(t *testing.T) {
	w := triage.Weights{
		"info":    10.0,
		"warning": 1.0,
	}
	tr := triage.New(w)
	tr.Add(makeAlert(alerting.Info, "info"))
	tr.Add(makeAlert(alerting.Warning, "warning"))

	ranked := tr.Ranked()
	if ranked[0].Alert.Message != "info" {
		t.Errorf("expected info first with custom weights, got %q", ranked[0].Alert.Message)
	}
}

func TestTriager_EmptyRanked(t *testing.T) {
	tr := triage.New(nil)
	if got := tr.Ranked(); len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}
