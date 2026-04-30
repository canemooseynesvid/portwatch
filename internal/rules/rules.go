package rules

import (
	"fmt"

	"portwatch/internal/config"
)

// Verdict represents the result of evaluating a port against the rule set.
type Verdict int

const (
	VerdictAllow   Verdict = iota // port is explicitly allowed
	VerdictDeny                   // port is explicitly denied
	VerdictUnknown                // port not covered by any rule
)

func (v Verdict) String() string {
	switch v {
	case VerdictAllow:
		return "allow"
	case VerdictDeny:
		return "deny"
	default:
		return "unknown"
	}
}

// Evaluator checks port/protocol pairs against configured allow and deny lists.
type Evaluator struct {
	allowed map[string]struct{}
	denied  map[string]struct{}
}

// NewEvaluator builds an Evaluator from the provided config.
func NewEvaluator(cfg *config.Config) *Evaluator {
	e := &Evaluator{
		allowed: make(map[string]struct{}, len(cfg.AllowedPorts)),
		denied:  make(map[string]struct{}, len(cfg.DeniedPorts)),
	}
	for _, p := range cfg.AllowedPorts {
		e.allowed[portKey(p.Port, p.Protocol)] = struct{}{}
	}
	for _, p := range cfg.DeniedPorts {
		e.denied[portKey(p.Port, p.Protocol)] = struct{}{}
	}
	return e
}

// Evaluate returns the Verdict for the given port and protocol.
// Deny rules take precedence over allow rules.
func (e *Evaluator) Evaluate(port uint16, protocol string) Verdict {
	key := portKey(int(port), protocol)
	if _, ok := e.denied[key]; ok {
		return VerdictDeny
	}
	if _, ok := e.allowed[key]; ok {
		return VerdictAllow
	}
	return VerdictUnknown
}

func portKey(port int, protocol string) string {
	return fmt.Sprintf("%s:%d", protocol, port)
}
