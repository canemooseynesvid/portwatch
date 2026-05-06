package watchlist_test

import (
	"testing"

	"portwatch/internal/alerting"
	"portwatch/internal/portscanner"
	"portwatch/internal/watchlist"
)

func TestWatchlist_AddAndContains(t *testing.T) {
	wl := watchlist.New()
	wl.Add(watchlist.Entry{Port: 22, Protocol: "tcp", Label: "SSH"})

	se := portscanner.Entry{Protocol: "tcp", LocalPort: 22}
	e, ok := wl.Contains(se)
	if !ok {
		t.Fatal("expected watchlist to contain port 22/tcp")
	}
	if e.Label != "SSH" {
		t.Errorf("expected label SSH, got %q", e.Label)
	}
}

func TestWatchlist_Remove(t *testing.T) {
	wl := watchlist.New()
	wl.Add(watchlist.Entry{Port: 80, Protocol: "tcp", Label: "HTTP"})
	wl.Remove("tcp", 80)

	se := portscanner.Entry{Protocol: "tcp", LocalPort: 80}
	if _, ok := wl.Contains(se); ok {
		t.Fatal("expected port 80/tcp to be removed")
	}
}

func TestWatchlist_Len(t *testing.T) {
	wl := watchlist.New()
	if wl.Len() != 0 {
		t.Fatalf("expected 0, got %d", wl.Len())
	}
	wl.Add(watchlist.Entry{Port: 443, Protocol: "tcp"})
	wl.Add(watchlist.Entry{Port: 53, Protocol: "udp"})
	if wl.Len() != 2 {
		t.Fatalf("expected 2, got %d", wl.Len())
	}
}

func TestWatchlist_All(t *testing.T) {
	wl := watchlist.New()
	wl.Add(watchlist.Entry{Port: 8080, Protocol: "tcp", Label: "dev"})
	all := wl.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(all))
	}
	if all[0].Label != "dev" {
		t.Errorf("unexpected label: %q", all[0].Label)
	}
}

func TestChecker_CheckAdded_EmitsAlert(t *testing.T) {
	wl := watchlist.New()
	wl.Add(watchlist.Entry{Port: 22, Protocol: "tcp", Label: "SSH"})

	var collected []alerting.Alert
	alerter := alerting.NewAlerter(alerting.CollectorHandler(&collected))
	checker := watchlist.NewChecker(wl, alerter)

	checker.CheckAdded([]portscanner.Entry{
		{Protocol: "tcp", LocalPort: 22, PID: 1234},
	})

	if len(collected) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(collected))
	}
	if collected[0].Level != alerting.Warning {
		t.Errorf("expected Warning level, got %v", collected[0].Level)
	}
}

func TestChecker_CheckAdded_NoAlertForUnwatched(t *testing.T) {
	wl := watchlist.New()

	var collected []alerting.Alert
	alerter := alerting.NewAlerter(alerting.CollectorHandler(&collected))
	checker := watchlist.NewChecker(wl, alerter)

	checker.CheckAdded([]portscanner.Entry{
		{Protocol: "tcp", LocalPort: 9999},
	})

	if len(collected) != 0 {
		t.Fatalf("expected no alerts, got %d", len(collected))
	}
}
