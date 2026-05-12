// Package partition groups a stream of alerts into named buckets
// based on a user-supplied classifier function. Useful for routing
// alerts to per-protocol or per-severity processing lanes.
package partition

import (
	"sync"

	"portwatch/internal/alerting"
)

// Classifier returns the bucket key for a given alert.
// Returning an empty string causes the alert to be discarded.
type Classifier func(a alerting.Alert) string

// Handler is called with every alert routed to a bucket.
type Handler func(bucket string, a alerting.Alert) error

// Partitioner distributes alerts into named buckets and forwards
// each one to a registered Handler.
type Partitioner struct {
	mu         sync.RWMutex
	classify   Classifier
	handlers   map[string]Handler
	fallback   Handler
}

// New creates a Partitioner using the supplied Classifier.
// A nil fallback silently drops alerts whose bucket has no handler.
func New(classify Classifier, fallback Handler) *Partitioner {
	return &Partitioner{
		classify: classify,
		handlers: make(map[string]Handler),
		fallback: fallback,
	}
}

// Register associates a Handler with a bucket key.
// Calling Register with the same key replaces the previous handler.
func (p *Partitioner) Register(bucket string, h Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[bucket] = h
}

// Send classifies the alert and dispatches it to the matching handler.
// If no handler is found and a fallback was provided, the fallback is
// called instead. An empty bucket key causes the alert to be dropped.
func (p *Partitioner) Send(a alerting.Alert) error {
	bucket := p.classify(a)
	if bucket == "" {
		return nil
	}

	p.mu.RLock()
	h, ok := p.handlers[bucket]
	p.mu.RUnlock()

	if ok {
		return h(bucket, a)
	}
	if p.fallback != nil {
		return p.fallback(bucket, a)
	}
	return nil
}

// Buckets returns the currently registered bucket keys.
func (p *Partitioner) Buckets() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	keys := make([]string, 0, len(p.handlers))
	for k := range p.handlers {
		keys = append(keys, k)
	}
	return keys
}
