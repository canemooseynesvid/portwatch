package trend_test

import (
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/trend"
)

type collectorHandler struct {
	alerts []alerting.Alert
}

func (c *collectorHandler) Handle(a alerting.Alert) {
	c.alerts = append(c.alerts, a)
}

func newCollectorAlerter() (*alerting.Alerter, *collectorHandler) {
	ch := &collectorHandler{}
	a := alerting.NewAlerter(ch)
	return a, ch
}

func TestChecker_NoSpikeWhenBelowThreshold(t *testing.T) {
	alerter, col := newCollectorAlerter()
	checker := trend.NewChecker(alerter, 5, time.Minute)

	for i := 0; i < 4; i++ {
		checker.Record("tcp", uint16(8080+i))
	}

	if len(col.alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(col.alerts))
	}
}

func TestChecker_EmitsSpikeAlertWhenThresholdExceeded(t *testing.T) {
	alerter, col := newCollectorAlerter()
	checker := trend.NewChecker(alerter, 3, time.Minute)

	for i := 0; i < 5; i++ {
		checker.Record("tcp", uint16(9000+i))
	}

	if len(col.alerts) == 0 {
		t.Fatal("expected at least one spike alert")
	}

	got := col.alerts[0]
	if got.Level != alerting.LevelWarning {
		t.Errorf("expected Warning level, got %s", got.Level)
	}
}

func TestChecker_SpikeAlertContainsProtocol(t *testing.T) {
	alerter, col := newCollectorAlerter()
	checker := trend.NewChecker(alerter, 2, time.Minute)

	for i := 0; i < 4; i++ {
		checker.Record("udp", uint16(5000+i))
	}

	if len(col.alerts) == 0 {
		t.Fatal("expected spike alert")
	}

	msg := col.alerts[0].Message
	if msg == "" {
		t.Error("expected non-empty alert message")
	}
}

func TestChecker_ResetClearsTrend(t *testing.T) {
	alerter, col := newCollectorAlerter()
	checker := trend.NewChecker(alerter, 3, time.Minute)

	for i := 0; i < 5; i++ {
		checker.Record("tcp", uint16(7000+i))
	}

	initialAlerts := len(col.alerts)
	checker.Reset()

	for i := 0; i < 2; i++ {
		checker.Record("tcp", uint16(7100+i))
	}

	if len(col.alerts) != initialAlerts {
		t.Errorf("expected no new alerts after reset, got %d new", len(col.alerts)-initialAlerts)
	}
}
