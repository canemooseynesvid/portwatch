package schema_test

import (
	"strings"
	"testing"

	"portwatch/internal/alerting"
	"portwatch/internal/portscanner"
	"portwatch/internal/schema"
)

func goodEntry() portscanner.Entry {
	return portscanner.Entry{Port: 8080, Protocol: "tcp", Addr: "0.0.0.0"}
}

func TestValidate_ValidEntry_NoAlerts(t *testing.T) {
	v := schema.New(schema.DefaultRules())
	alerts := v.Validate(goodEntry())
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestValidate_ZeroPort_EmitsWarning(t *testing.T) {
	v := schema.New(schema.DefaultRules())
	e := goodEntry()
	e.Port = 0
	alerts := v.Validate(e)
	if len(alerts) == 0 {
		t.Fatal("expected at least one alert for zero port")
	}
	if alerts[0].Level != alerting.LevelWarning {
		t.Errorf("expected Warning, got %s", alerts[0].Level)
	}
	if !strings.Contains(alerts[0].Message, "non-zero-port") {
		t.Errorf("message should mention rule name, got: %s", alerts[0].Message)
	}
}

func TestValidate_UnknownProtocol_EmitsWarning(t *testing.T) {
	v := schema.New(schema.DefaultRules())
	e := goodEntry()
	e.Protocol = "sctp"
	alerts := v.Validate(e)
	if len(alerts) == 0 {
		t.Fatal("expected alert for unknown protocol")
	}
	if !strings.Contains(alerts[0].Message, "known-protocol") {
		t.Errorf("message should mention rule name, got: %s", alerts[0].Message)
	}
}

func TestValidate_EmptyAddr_EmitsInfo(t *testing.T) {
	v := schema.New(schema.DefaultRules())
	e := goodEntry()
	e.Addr = ""
	alerts := v.Validate(e)
	if len(alerts) == 0 {
		t.Fatal("expected alert for empty addr")
	}
	if alerts[0].Level != alerting.LevelInfo {
		t.Errorf("expected Info level, got %s", alerts[0].Level)
	}
}

func TestRuleCount(t *testing.T) {
	v := schema.New(schema.DefaultRules())
	if v.RuleCount() != len(schema.DefaultRules()) {
		t.Errorf("rule count mismatch")
	}
}

func TestChecker_Check_CountsViolations(t *testing.T) {
	collected := []alerting.Alert{}
	h := alerting.CollectorHandler(&collected)
	a := alerting.NewAlerter(h)
	v := schema.New(schema.DefaultRules())
	ch := schema.NewChecker(v, a)

	entries := []portscanner.Entry{
		goodEntry(),
		{Port: 0, Protocol: "tcp", Addr: "127.0.0.1"},  // violates non-zero-port
		{Port: 80, Protocol: "xyz", Addr: "127.0.0.1"}, // violates known-protocol
	}

	n := ch.Check(entries)
	if n != 2 {
		t.Errorf("expected 2 violations, got %d", n)
	}
	if len(collected) != 2 {
		t.Errorf("expected 2 collected alerts, got %d", len(collected))
	}
}
