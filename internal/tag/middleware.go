package tag

import (
	"fmt"

	"portwatch/internal/alerting"
)

// Middleware injects tags from a Store into every alert that passes
// through it. The entity key is derived from the alert's port metadata
// when present, falling back to the alert's message.
type Middleware struct {
	store *Store
	next  alerting.Handler
}

// NewMiddleware returns a Middleware that enriches alerts with tags
// from store before forwarding them to next.
func NewMiddleware(store *Store, next alerting.Handler) *Middleware {
	return &Middleware{store: store, next: next}
}

// Handle enriches the alert with stored tags, then calls next.
func (m *Middleware) Handle(a alerting.Alert) error {
	if m.next == nil {
		return nil
	}
	entity := entityKey(a)
	if tags, ok := m.store.Get(entity); ok {
		for k, v := range tags {
			a.Tags[k] = v
		}
	}
	return m.next.Handle(a)
}

// entityKey builds a lookup key from alert metadata. It prefers
// port+protocol when available so that tags are matched per-port.
func entityKey(a alerting.Alert) string {
	port, hasPort := a.Meta["port"]
	proto, hasProto := a.Meta["protocol"]
	if hasPort && hasProto {
		return fmt.Sprintf("%s/%s", proto, port)
	}
	return a.Message
}
