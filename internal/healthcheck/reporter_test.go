package healthcheck_test

import (
	"strings"
	"testing"
	"time"

	"portwatch/internal/healthcheck"
)

func TestPrint_ContainsHeaders(t *testing.T) {
	var sb strings.Builder
	results := []healthcheck.Result{
		{Name: "procfs", Status: healthcheck.StatusOK, Message: "accessible", CheckedAt: time.Now()},
	}
	healthcheck.Print(&sb, results)
	out := sb.String()
	for _, want := range []string{"NAME", "STATUS", "MESSAGE", "procfs", "ok", "Overall"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}

func TestPrint_EmptyResults(t *testing.T) {
	var sb strings.Builder
	healthcheck.Print(&sb, nil)
	out := sb.String()
	if !strings.Contains(out, "Overall") {
		t.Errorf("expected Overall line in output:\n%s", out)
	}
}

func TestPrint_ShowsFailedOverall(t *testing.T) {
	var sb strings.Builder
	results := []healthcheck.Result{
		{Name: "x", Status: healthcheck.StatusFailed, Message: "boom", CheckedAt: time.Now()},
	}
	healthcheck.Print(&sb, results)
	if !strings.Contains(sb.String(), string(healthcheck.StatusFailed)) {
		t.Error("expected 'failed' in output")
	}
}
