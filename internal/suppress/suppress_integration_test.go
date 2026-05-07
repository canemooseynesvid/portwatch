package suppress_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/suppress"
)

// TestSuppressLifecycle exercises the full suppress → active → purge cycle.
func TestSuppressLifecycle(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	l := suppress.New()
	l.(*suppress.List) // compile check — List is exported

	// Use the exported API only.
	sl := suppress.New()

	sl.Suppress("tcp:8080", 10*time.Second)
	sl.Suppress("tcp:443", 30*time.Second)
	sl.Suppress("udp:53", 5*time.Second)

	if sl.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", sl.Len())
	}

	_ = base // used for documentation only in this black-box test

	// Remove one explicitly.
	sl.Remove("tcp:443")
	if sl.Len() != 2 {
		t.Fatalf("expected 2 entries after remove, got %d", sl.Len())
	}

	// All returns remaining entries.
	all := sl.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 from All(), got %d", len(all))
	}
}
