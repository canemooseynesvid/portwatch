// Package pipeline provides a composable alert processing pipeline
// that chains multiple alert handlers in sequence.
package pipeline

import (
	"context"

	"portwatch/internal/alerting"
)

// Stage is a function that processes an alert and optionally forwards it.
// Returning false drops the alert from further processing.
type Stage func(alert alerting.Alert) bool

// Handler is a function that receives a final processed alert.
type Handler func(alert alerting.Alert)

// Pipeline chains stages and delivers passing alerts to a handler.
type Pipeline struct {
	stages  []Stage
	sink    Handler
}

// New creates a Pipeline that delivers alerts to sink after passing all stages.
func New(sink Handler, stages ...Stage) *Pipeline {
	return &Pipeline{
		stages: stages,
		sink:   sink,
	}
}

// Process runs the alert through all stages. If any stage returns false
// the alert is dropped. Otherwise it is delivered to the sink.
func (p *Pipeline) Process(alert alerting.Alert) {
	for _, stage := range p.stages {
		if !stage(alert) {
			return
		}
	}
	if p.sink != nil {
		p.sink(alert)
	}
}

// ProcessCtx is like Process but respects context cancellation.
func (p *Pipeline) ProcessCtx(ctx context.Context, alert alerting.Alert) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	p.Process(alert)
}

// MinLevel returns a Stage that drops alerts below the given level.
func MinLevel(level alerting.AlertLevel) Stage {
	return func(alert alerting.Alert) bool {
		return alert.Level >= level
	}
}

// WithTag returns a Stage that drops alerts not containing the given tag.
func WithTag(tag string) Stage {
	return func(alert alerting.Alert) bool {
		for _, t := range alert.Tags {
			if t == tag {
				return true
			}
		}
		return false
	}
}

// AlwaysAllow is a no-op Stage that passes every alert.
func AlwaysAllow() Stage {
	return func(_ alerting.Alert) bool { return true }
}
