package replay_test

import (
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/replay"
	"portwatch/internal/snapshot"
)

func makeAlert(level alerting.Level, msg string, at time.Time) alerting.Alert {
	return alerting.Alert{Level: level, Message: msg, Timestamp: at}
}

func TestWithTimeRange_Inside(t *testing.T) {
	now := time.Now()
	tr := replay.TimeRange{From: now.Add(-time.Hour), To: now.Add(time.Hour)}
	f := replay.WithTimeRange(tr)
	if !f(makeAlert(alerting.LevelInfo, "x", now)) {
		t.Error("expected alert within range to pass")
	}
}

func TestWithTimeRange_Outside(t *testing.T) {
	now := time.Now()
	tr := replay.TimeRange{From: now.Add(time.Hour), To: now.Add(2 * time.Hour)}
	f := replay.WithTimeRange(tr)
	if f(makeAlert(alerting.LevelInfo, "x", now)) {
		t.Error("expected alert outside range to be blocked")
	}
}

func TestWithMinLevel_Passes(t *testing.T) {
	f := replay.WithMinLevel(alerting.LevelWarn)
	if !f(makeAlert(alerting.LevelWarn, "x", time.Now())) {
		t.Error("expected warn to pass min-level warn filter")
	}
	if !f(makeAlert(alerting.LevelError, "x", time.Now())) {
		t.Error("expected error to pass min-level warn filter")
	}
}

func TestWithMinLevel_Blocks(t *testing.T) {
	f := replay.WithMinLevel(alerting.LevelWarn)
	if f(makeAlert(alerting.LevelInfo, "x", time.Now())) {
		t.Error("expected info to be blocked by min-level warn filter")
	}
}

func TestWithKind_MatchesPrefix(t *testing.T) {
	f := replay.WithKind(snapshot.EventAdded)
	prefix := snapshot.EventAdded.String() + ": tcp:8080"
	if !f(makeAlert(alerting.LevelInfo, prefix, time.Now())) {
		t.Error("expected matching prefix to pass")
	}
}

func TestCombineFilters_AllMustPass(t *testing.T) {
	now := time.Now()
	f := replay.CombineFilters(
		replay.WithMinLevel(alerting.LevelWarn),
		replay.WithTimeRange(replay.TimeRange{From: now.Add(-time.Minute)}),
	)
	if f(makeAlert(alerting.LevelInfo, "x", now)) {
		t.Error("info should be blocked by combined filter")
	}
	if !f(makeAlert(alerting.LevelWarn, "x", now)) {
		t.Error("warn within range should pass combined filter")
	}
}
