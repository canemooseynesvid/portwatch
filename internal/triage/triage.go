// Package triage scores and prioritises alerts based on configurable
// severity weights, producing a ranked list for operator review.
package triage

import (
	"sort"
	"sync"

	"portwatch/internal/alerting"
)

// Score holds a scored alert ready for triage output.
type Score struct {
	Alert alerting.Alert
	Value float64
}

// Weights maps alert level strings to numeric multipliers.
type Weights map[string]float64

// DefaultWeights returns sensible default level weights.
func DefaultWeights() Weights {
	return Weights{
		"info":     1.0,
		"warning":  2.0,
		"error":    4.0,
		"critical": 8.0,
	}
}

// Triager collects alerts and ranks them by score.
type Triager struct {
	mu      sync.Mutex
	weights Weights
	pending []Score
}

// New creates a Triager with the given weights.
func New(w Weights) *Triager {
	if w == nil {
		w = DefaultWeights()
	}
	return &Triager{weights: w}
}

// Add scores an alert and appends it to the pending queue.
func (t *Triager) Add(a alerting.Alert) {
	v := t.score(a)
	t.mu.Lock()
	t.pending = append(t.pending, Score{Alert: a, Value: v})
	t.mu.Unlock()
}

// Ranked returns a copy of pending alerts sorted highest score first,
// then clears the internal queue.
func (t *Triager) Ranked() []Score {
	t.mu.Lock()
	out := make([]Score, len(t.pending))
	copy(out, t.pending)
	t.pending = t.pending[:0]
	t.mu.Unlock()

	sort.Slice(out, func(i, j int) bool {
		return out[i].Value > out[j].Value
	})
	return out
}

// Len returns the number of pending alerts.
func (t *Triager) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.pending)
}

func (t *Triager) score(a alerting.Alert) float64 {
	level := a.Level.String()
	if w, ok := t.weights[level]; ok {
		return w
	}
	return 1.0
}
