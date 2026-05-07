package watchlist_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/example/portwatch/internal/alerting"
	"github.com/example/portwatch/internal/portscanner"
	"github.com/example/portwatch/internal/watchlist"
)

func TestWatchlist_AddAndContains(t *testing.T) {
	wl := watchlist.New()
	wl.Add(8080, "tcp")
	if !wl.Contains(8080, "tcp") {
		t.Fatal("expected watchlist to contain 8080/tcp")
	}
	if wl.Contains(9090, "tcp") {
		t.Fatal("did not expect watchlist to contain 9090/tcp")
	}
}

func TestWatchlist_Remove(t *testing.T) {
	wl := watchlist.New()
	wl.Add(443, "tcp")
	wl.Remove(443, "tcp")
	if wl.Contains(443, "tcp") {
		t.Fatal("expected 443/tcp to be removed")
	}
}

func TestWatchlist_Len(t *testing.T) {
	wl := watchlist.New()
	wl.Add(80, "tcp")
	wl.Add(443, "tcp")
	wl.Add(53, "udp")
	if wl.Len() != 3 {
		t.Fatalf("expected len 3, got %d", wl.Len())
	}
}

func TestWatchlist_All(t *testing.T) {
	wl := watchlist.New()
	wl.Add(22, "tcp")
	wl.Add(53, "udp")
	entries := wl.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestChecker_CheckAdded_EmitsAlert(t *testing.T) {
	wl := watchlist.New()
	wl.Add(8080, "tcp")

	var buf bytes.Buffer
	alerter := alerting.NewAlerter(alerting.WriterHandler(&buf))
	checker := watchlist.NewChecker(wl, alerter)

	e := portscanner.Entry{Port: 8080, Protocol: "tcp", PID: 42}
	checker.CheckAdded(e)

	if !strings.Contains(buf.String(), "8080") {
		t.Fatalf("expected alert output to mention port 8080, got: %s", buf.String())
	}
}

func TestChecker_CheckRemoved_EmitsAlert(t *testing.T) {
	wl := watchlist.New()
	wl.Add(9090, "tcp")

	var buf bytes.Buffer
	alerter := alerting.NewAlerter(alerting.WriterHandler(&buf))
	checker := watchlist.NewChecker(wl, alerter)

	e := portscanner.Entry{Port: 9090, Protocol: "tcp", PID: 7}
	checker.CheckRemoved(e)

	if !strings.Contains(buf.String(), "closed") {
		t.Fatalf("expected alert output to mention 'closed', got: %s", buf.String())
	}
}

func TestChecker_CheckAdded_NoAlertForUnwatched(t *testing.T) {
	wl := watchlist.New()

	var buf bytes.Buffer
	alerter := alerting.NewAlerter(alerting.WriterHandler(&buf))
	checker := watchlist.NewChecker(wl, alerter)

	e := portscanner.Entry{Port: 3000, Protocol: "tcp", PID: 99}
	checker.CheckAdded(e)

	if buf.Len() != 0 {
		t.Fatalf("expected no output for unwatched port, got: %s", buf.String())
	}
}
