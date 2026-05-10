package correlate_test

import (
	"testing"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/correlate"
)

func makeAlert(port, proto, msg string) alerting.Alert {
	return alerting.Alert{
		Level:   alerting.LevelWarning,
		Message: msg,
		Meta:    map[string]string{"port": port, "protocol": proto},
	}
}

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRecord_GroupsByPortAndProtocol(t *testing.T) {
	now := time.Now()
	c := correlate.NewWithClock(time.Minute, fixedClock(now))

	c.Record(makeAlert("8080", "tcp", "first"))
	c.Record(makeAlert("8080", "tcp", "second"))
	c.Record(makeAlert("443", "tcp", "other"))

	groups := c.Groups()
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestRecord_SameKeyAccumulates(t *testing.T) {
	now := time.Now()
	c := correlate.NewWithClock(time.Minute, fixedClock(now))

	c.Record(makeAlert("9090", "udp", "a"))
	c.Record(makeAlert("9090", "udp", "b"))
	c.Record(makeAlert("9090", "udp", "c"))

	groups := c.Groups()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Alerts) != 3 {
		t.Errorf("expected 3 alerts in group, got %d", len(groups[0].Alerts))
	}
}

func TestRecord_ExpiredGroupsArePurged(t *testing.T) {
	base := time.Now()
	current := base
	clock := func() time.Time { return current }

	c := correlate.NewWithClock(30*time.Second, clock)
	c.Record(makeAlert("8080", "tcp", "old"))

	// advance past the window
	current = base.Add(60 * time.Second)
	c.Record(makeAlert("9090", "tcp", "new"))

	groups := c.Groups()
	if len(groups) != 1 {
		t.Fatalf("expected expired group to be purged, got %d groups", len(groups))
	}
	if groups[0].Key != "tcp/9090" {
		t.Errorf("expected surviving group key tcp/9090, got %s", groups[0].Key)
	}
}

func TestGroups_ReturnsCopy(t *testing.T) {
	now := time.Now()
	c := correlate.NewWithClock(time.Minute, fixedClock(now))
	c.Record(makeAlert("80", "tcp", "x"))

	g1 := c.Groups()
	g1[0].Alerts = nil // mutate the copy

	g2 := c.Groups()
	if len(g2[0].Alerts) != 1 {
		t.Error("Groups() should return independent copies")
	}
}

func TestRecord_SetsFirstAndLastTimestamps(t *testing.T) {
	base := time.Now()
	current := base
	clock := func() time.Time { return current }

	c := correlate.NewWithClock(time.Minute, clock)
	c.Record(makeAlert("22", "tcp", "first"))

	current = base.Add(5 * time.Second)
	c.Record(makeAlert("22", "tcp", "second"))

	groups := c.Groups()
	if len(groups) != 1 {
		t.Fatal("expected 1 group")
	}
	g := groups[0]
	if !g.First.Equal(base) {
		t.Errorf("First mismatch: got %v want %v", g.First, base)
	}
	if !g.Last.Equal(current) {
		t.Errorf("Last mismatch: got %v want %v", g.Last, current)
	}
}
