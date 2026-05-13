package classify_test

import (
	"testing"

	"github.com/example/portwatch/internal/classify"
	"github.com/example/portwatch/internal/portscanner"
)

func TestCategorize_System(t *testing.T) {
	c := classify.New()
	e := portscanner.Entry{Port: 80}
	if got := c.Categorize(e); got != classify.CategorySystem {
		t.Fatalf("expected system, got %s", got)
	}
}

func TestCategorize_Registered(t *testing.T) {
	c := classify.New()
	e := portscanner.Entry{Port: 8080}
	if got := c.Categorize(e); got != classify.CategoryRegistered {
		t.Fatalf("expected registered, got %s", got)
	}
}

func TestCategorize_Ephemeral(t *testing.T) {
	c := classify.New()
	e := portscanner.Entry{Port: 55000}
	if got := c.Categorize(e); got != classify.CategoryEphemeral {
		t.Fatalf("expected ephemeral, got %s", got)
	}
}

func TestCategorize_Boundary(t *testing.T) {
	c := classify.New()
	cases := []struct {
		port uint16
		want classify.Category
	}{
		{1, classify.CategorySystem},
		{1023, classify.CategorySystem},
		{1024, classify.CategoryRegistered},
		{49151, classify.CategoryRegistered},
		{49152, classify.CategoryEphemeral},
		{65535, classify.CategoryEphemeral},
	}
	for _, tc := range cases {
		e := portscanner.Entry{Port: tc.port}
		if got := c.Categorize(e); got != tc.want {
			t.Errorf("port %d: expected %s, got %s", tc.port, tc.want, got)
		}
	}
}

func TestAlertLevel_System(t *testing.T) {
	lvl := classify.AlertLevel(classify.CategorySystem)
	if lvl.String() != "warning" {
		t.Fatalf("expected warning, got %s", lvl)
	}
}

func TestAlertLevel_Registered(t *testing.T) {
	lvl := classify.AlertLevel(classify.CategoryRegistered)
	if lvl.String() != "info" {
		t.Fatalf("expected info, got %s", lvl)
	}
}

func TestCategoryString(t *testing.T) {
	if classify.CategoryUnknown.String() != "unknown" {
		t.Fatal("unexpected string for unknown")
	}
}
