package classify_test

import (
	"testing"
	"time"

	"github.com/example/portwatch/internal/alerting"
	"github.com/example/portwatch/internal/classify"
)

type collectorHandler struct {
	alerts []alerting.Alert
}

func (c *collectorHandler) Handle(a alerting.Alert) error {
	c.alerts = append(c.alerts, a)
	return nil
}

func makeTestAlert(port any) alerting.Alert {
	return alerting.Alert{
		Level:   alerting.LevelInfo,
		Message: "test",
		Time:    time.Now(),
		Meta:    map[string]any{"port": port},
	}
}

func TestTagMiddleware_SystemPortTagged(t *testing.T) {
	col := &collectorHandler{}
	mw := classify.NewTagMiddleware(col)

	if err := mw.Handle(makeTestAlert(uint16(443))); err != nil {
		t.Fatal(err)
	}
	if len(col.alerts) != 1 {
		t.Fatal("expected one alert")
	}
	if col.alerts[0].Meta["category"] != "system" {
		t.Fatalf("expected system category, got %v", col.alerts[0].Meta["category"])
	}
}

func TestTagMiddleware_NoPortMeta(t *testing.T) {
	col := &collectorHandler{}
	mw := classify.NewTagMiddleware(col)
	a := alerting.Alert{
		Level:   alerting.LevelInfo,
		Message: "no port",
		Time:    time.Now(),
		Meta:    map[string]any{},
	}
	if err := mw.Handle(a); err != nil {
		t.Fatal(err)
	}
	if _, ok := col.alerts[0].Meta["category"]; ok {
		t.Fatal("expected no category tag")
	}
}

func TestTagMiddleware_NilNext(t *testing.T) {
	mw := classify.NewTagMiddleware(nil)
	if err := mw.Handle(makeTestAlert(uint16(80))); err != nil {
		t.Fatalf("expected no error with nil next, got %v", err)
	}
}
