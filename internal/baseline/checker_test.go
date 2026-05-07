package baseline_test

import (
	"strings"
	"testing"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/baseline"
	"github.com/user/portwatch/internal/portscanner"
)

func makeCollector() (*alerting.Alerter, *[]alerting.Alert) {
	var collected []alerting.Alert
	h := alerting.CollectorHandler(&collected)
	a := alerting.NewAlerter(h)
	return a, &collected
}

func TestChecker_UnknownPortEmitsWarning(t *testing.T) {
	b, _ := baseline.New(tempPath(t))
	a, collected := makeCollector()
	c := baseline.NewChecker(b, a)

	e := portscanner.Entry{Protocol: "tcp", LocalAddr: "0.0.0.0", LocalPort: 9000}
	c.Check(e)

	if len(*collected) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(*collected))
	}
	if (*collected)[0].Level != alerting.Warning {
		t.Errorf("expected Warning level, got %v", (*collected)[0].Level)
	}
	if !strings.Contains((*collected)[0].Message, "9000") {
		t.Errorf("alert message missing port: %s", (*collected)[0].Message)
	}
}

func TestChecker_KnownPortNoAlert(t *testing.T) {
	b, _ := baseline.New(tempPath(t))
	b.Add(baseline.Entry{Protocol: "tcp", Address: "0.0.0.0", Port: 443})
	a, collected := makeCollector()
	c := baseline.NewChecker(b, a)

	e := portscanner.Entry{Protocol: "tcp", LocalAddr: "0.0.0.0", LocalPort: 443}
	c.Check(e)

	if len(*collected) != 0 {
		t.Fatalf("expected no alerts, got %d", len(*collected))
	}
}

func TestChecker_LearnSuppressesFutureAlert(t *testing.T) {
	b, _ := baseline.New(tempPath(t))
	a, collected := makeCollector()
	c := baseline.NewChecker(b, a)

	e := portscanner.Entry{Protocol: "udp", LocalAddr: "127.0.0.1", LocalPort: 5353}
	c.Check(e) // unknown — should alert
	if len(*collected) != 1 {
		t.Fatalf("expected 1 alert before learn, got %d", len(*collected))
	}

	if err := c.Learn(e); err != nil {
		t.Fatalf("Learn failed: %v", err)
	}
	*collected = nil
	c.Check(e) // now known — no alert
	if len(*collected) != 0 {
		t.Fatalf("expected 0 alerts after learn, got %d", len(*collected))
	}
}

func TestChecker_ForgetReEnablesAlert(t *testing.T) {
	b, _ := baseline.New(tempPath(t))
	b.Add(baseline.Entry{Protocol: "tcp", Address: "0.0.0.0", Port: 8080})
	a, collected := makeCollector()
	c := baseline.NewChecker(b, a)

	e := portscanner.Entry{Protocol: "tcp", LocalAddr: "0.0.0.0", LocalPort: 8080}
	c.Check(e) // known — no alert
	if len(*collected) != 0 {
		t.Fatalf("expected no alert before forget, got %d", len(*collected))
	}

	c.Forget("tcp", "0.0.0.0", 8080)
	c.Check(e) // forgotten — should alert again
	if len(*collected) != 1 {
		t.Fatalf("expected 1 alert after forget, got %d", len(*collected))
	}
}
