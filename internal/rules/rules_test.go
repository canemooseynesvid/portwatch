package rules_test

import (
	"testing"

	"portwatch/internal/config"
	"portwatch/internal/rules"
)

func makeConfig(allowed, denied []config.PortRule) *config.Config {
	cfg := config.Default()
	cfg.AllowedPorts = allowed
	cfg.DeniedPorts = denied
	return cfg
}

func TestEvaluate_AllowedPort(t *testing.T) {
	cfg := makeConfig(
		[]config.PortRule{{Port: 8080, Protocol: "tcp"}},
		nil,
	)
	e := rules.NewEvaluator(cfg)
	if got := e.Evaluate(8080, "tcp"); got != rules.VerdictAllow {
		t.Errorf("expected allow, got %s", got)
	}
}

func TestEvaluate_DeniedPort(t *testing.T) {
	cfg := makeConfig(
		nil,
		[]config.PortRule{{Port: 9999, Protocol: "tcp"}},
	)
	e := rules.NewEvaluator(cfg)
	if got := e.Evaluate(9999, "tcp"); got != rules.VerdictDeny {
		t.Errorf("expected deny, got %s", got)
	}
}

func TestEvaluate_UnknownPort(t *testing.T) {
	e := rules.NewEvaluator(config.Default())
	if got := e.Evaluate(12345, "tcp"); got != rules.VerdictUnknown {
		t.Errorf("expected unknown, got %s", got)
	}
}

func TestEvaluate_DenyTakesPrecedence(t *testing.T) {
	cfg := makeConfig(
		[]config.PortRule{{Port: 443, Protocol: "tcp"}},
		[]config.PortRule{{Port: 443, Protocol: "tcp"}},
	)
	e := rules.NewEvaluator(cfg)
	if got := e.Evaluate(443, "tcp"); got != rules.VerdictDeny {
		t.Errorf("expected deny to take precedence, got %s", got)
	}
}

func TestEvaluate_ProtocolDistinct(t *testing.T) {
	cfg := makeConfig(
		[]config.PortRule{{Port: 53, Protocol: "udp"}},
		nil,
	)
	e := rules.NewEvaluator(cfg)
	if got := e.Evaluate(53, "tcp"); got != rules.VerdictUnknown {
		t.Errorf("tcp/53 should be unknown when only udp/53 is allowed, got %s", got)
	}
	if got := e.Evaluate(53, "udp"); got != rules.VerdictAllow {
		t.Errorf("udp/53 should be allowed, got %s", got)
	}
}

func TestVerdictString(t *testing.T) {
	cases := []struct {
		v    rules.Verdict
		want string
	}{
		{rules.VerdictAllow, "allow"},
		{rules.VerdictDeny, "deny"},
		{rules.VerdictUnknown, "unknown"},
	}
	for _, tc := range cases {
		if got := tc.v.String(); got != tc.want {
			t.Errorf("Verdict(%d).String() = %q, want %q", tc.v, got, tc.want)
		}
	}
}
