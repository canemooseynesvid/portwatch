package healthcheck_test

import (
	"strings"
	"testing"
	"time"

	"portwatch/internal/healthcheck"
)

func TestRegistry_RunAll_ReturnsResults(t *testing.T) {
	reg := healthcheck.New()
	reg.Register(healthcheck.NewFuncChecker("always-ok", func() healthcheck.Result {
		return healthcheck.Result{Status: healthcheck.StatusOK, Message: "fine"}
	}))

	results := reg.RunAll()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != healthcheck.StatusOK {
		t.Errorf("expected ok, got %s", results[0].Status)
	}
}

func TestRegistry_RunAll_SetsTimestamp(t *testing.T) {
	reg := healthcheck.New()
	reg.Register(healthcheck.NewFuncChecker("ts", func() healthcheck.Result {
		return healthcheck.Result{Status: healthcheck.StatusOK}
	}))
	results := reg.RunAll()
	if results[0].CheckedAt.IsZero() {
		t.Error("expected CheckedAt to be set")
	}
}

func TestOverall_WorstWins(t *testing.T) {
	tests := []struct {
		statuses []healthcheck.Status
		want     healthcheck.Status
	}{
		{[]healthcheck.Status{healthcheck.StatusOK, healthcheck.StatusOK}, healthcheck.StatusOK},
		{[]healthcheck.Status{healthcheck.StatusOK, healthcheck.StatusDegraded}, healthcheck.StatusDegraded},
		{[]healthcheck.Status{healthcheck.StatusDegraded, healthcheck.StatusFailed}, healthcheck.StatusFailed},
	}
	for _, tt := range tests {
		var results []healthcheck.Result
		for _, s := range tt.statuses {
			results = append(results, healthcheck.Result{Status: s})
		}
		if got := healthcheck.Overall(results); got != tt.want {
			t.Errorf("Overall(%v) = %s, want %s", tt.statuses, got, tt.want)
		}
	}
}

func TestUptimeChecker(t *testing.T) {
	start := time.Now().Add(-5 * time.Second)
	c := healthcheck.NewUptimeChecker(start)
	res := c.Check()
	if res.Status != healthcheck.StatusOK {
		t.Errorf("expected ok, got %s", res.Status)
	}
	if !strings.Contains(res.Message, "running for") {
		t.Errorf("unexpected message: %s", res.Message)
	}
}

func TestResultString(t *testing.T) {
	r := healthcheck.Result{Name: "foo", Status: healthcheck.StatusDegraded, Message: "slow"}
	got := r.String()
	if !strings.Contains(got, "foo") || !strings.Contains(got, "degraded") {
		t.Errorf("unexpected string: %s", got)
	}
}
