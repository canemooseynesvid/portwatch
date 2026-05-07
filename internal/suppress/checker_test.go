package suppress

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/snapshot"
)

func makeEntry(proto string, port uint16) snapshot.Entry {
	return snapshot.Entry{Protocol: proto, Port: port}
}

func TestChecker_AllowsUnsuppressed(t *testing.T) {
	l := New()
	c := NewChecker(l, time.Minute)

	if !c.Allow(makeEntry("tcp", 8080)) {
		t.Fatal("expected unsuppressed entry to be allowed")
	}
}

func TestChecker_BlocksSuppressed(t *testing.T) {
	l := New()
	c := NewChecker(l, time.Minute)
	e := makeEntry("tcp", 9090)

	c.SuppressEntry(e)

	if c.Allow(e) {
		t.Fatal("expected suppressed entry to be blocked")
	}
}

func TestChecker_SuppressFor(t *testing.T) {
	base := time.Now()
	l := New()
	l.now = fixedClock(base)
	c := NewChecker(l, time.Minute)
	e := makeEntry("udp", 5353)

	c.SuppressFor(e, 2*time.Second)

	l.now = fixedClock(base.Add(3 * time.Second))
	if !c.Allow(e) {
		t.Fatal("expected entry to be allowed after suppression expired")
	}
}

func TestChecker_FilterAlert_Suppressed(t *testing.T) {
	l := New()
	l.Suppress("tcp:80", time.Hour)
	c := NewChecker(l, time.Minute)

	a := &alerting.Alert{
		Fields: map[string]interface{}{"port_key": "tcp:80"},
	}

	if c.FilterAlert(a) != nil {
		t.Fatal("expected suppressed alert to be filtered")
	}
}

func TestChecker_FilterAlert_NotSuppressed(t *testing.T) {
	l := New()
	c := NewChecker(l, time.Minute)

	a := &alerting.Alert{
		Fields: map[string]interface{}{"port_key": "tcp:443"},
	}

	if c.FilterAlert(a) == nil {
		t.Fatal("expected unsuppressed alert to pass through")
	}
}

func TestChecker_FilterAlert_NoPortKey(t *testing.T) {
	l := New()
	c := NewChecker(l, time.Minute)

	a := &alerting.Alert{Fields: map[string]interface{}{}}
	if c.FilterAlert(a) == nil {
		t.Fatal("alert without port_key should pass through")
	}
}
