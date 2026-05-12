package triage

import "portwatch/internal/alerting"

// Middleware wraps an alerting.Handler, feeding every alert into the
// Triager before forwarding it downstream.
type Middleware struct {
	triager *Triager
	next    alerting.Handler
}

// NewMiddleware returns a Middleware that scores alerts via t and
// passes them on to next.
func NewMiddleware(t *Triager, next alerting.Handler) *Middleware {
	return &Middleware{triager: t, next: next}
}

// Handle scores the alert and forwards it.
func (m *Middleware) Handle(a alerting.Alert) error {
	m.triager.Add(a)
	if m.next != nil {
		return m.next.Handle(a)
	}
	return nil
}
