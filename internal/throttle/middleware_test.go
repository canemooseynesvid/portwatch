package throttle_test

import (
	"sync"
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/throttle"
)

// collector captures alerts for inspection.
type collector struct {
	mu     sync.Mutex
	alerts []alerting.Alert
}

func (c *collector) Handle(a alerting.Alert) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.alerts = append(c.alerts, a)
}

func (c *collector) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.alerts)
}

func makeAlert(port, proto string) alerting.Alert {
	return alerting.Alert{
		Message: "test alert",
		Level:   alerting.LevelWarn,
		Meta:    map[string]string{"port": port, "protocol": proto},
	}
}

func TestAlertThrottler_PassesWithinBurst(t *testing.T) {
	c := &collector{}
	th := throttle.NewAlertThrottler(c, 3, time.Minute)
	a := makeAlert("8080", "tcp")

	for i := 0; i < 3; i++ {
		th.Handle(a)
	}
	if c.Len() != 3 {
		t.Fatalf("expected 3 alerts forwarded, got %d", c.Len())
	}
}

func TestAlertThrottler_DropsBeyondBurst(t *testing.T) {
	c := &collector{}
	th := throttle.NewAlertThrottler(c, 2, time.Minute)
	a := makeAlert("443", "tcp")

	th.Handle(a)
	th.Handle(a)
	th.Handle(a) // should be dropped

	if c.Len() != 2 {
		t.Fatalf("expected 2 alerts, got %d", c.Len())
	}
}

func TestAlertThrottler_SeparateKeysIndependent(t *testing.T) {
	c := &collector{}
	th := throttle.NewAlertThrottler(c, 1, time.Minute)

	th.Handle(makeAlert("80", "tcp"))
	th.Handle(makeAlert("443", "tcp"))

	if c.Len() != 2 {
		t.Fatalf("expected 2 alerts for distinct keys, got %d", c.Len())
	}
}

func TestAlertThrottler_ResetKeyAllowsNext(t *testing.T) {
	c := &collector{}
	th := throttle.NewAlertThrottler(c, 1, time.Minute)
	a := makeAlert("9090", "udp")

	th.Handle(a) // consumes token
	th.Handle(a) // dropped
	th.ResetKey("udp/9090")
	th.Handle(a) // should pass

	if c.Len() != 2 {
		t.Fatalf("expected 2 alerts after reset, got %d", c.Len())
	}
}
