// Package replay provides a mechanism to replay historical alerts
// through a handler pipeline, useful for post-incident analysis and
// testing alert pipelines against recorded events.
package replay

import (
	"context"
	"time"

	"portwatch/internal/alerting"
	"portwatch/internal/snapshot"
)

// Handler is a function that receives a replayed alert.
type Handler func(alert alerting.Alert) error

// Options configures replay behaviour.
type Options struct {
	// Speed is the playback multiplier. 1.0 = real-time, 2.0 = double speed.
	// 0 means instant (no delay).
	Speed float64

	// Filter, if set, is called before each alert. Return false to skip it.
	Filter func(alerting.Alert) bool
}

// DefaultOptions returns sensible defaults for replay.
func DefaultOptions() Options {
	return Options{
		Speed:  0,
		Filter: nil,
	}
}

// Replayer replays a sequence of snapshot events as alerts.
type Replayer struct {
	history *snapshot.History
	opts    Options
}

// New creates a Replayer backed by the given history.
func New(h *snapshot.History, opts Options) *Replayer {
	if opts.Speed < 0 {
		opts.Speed = 0
	}
	return &Replayer{history: h, opts: opts}
}

// Run replays all events in chronological order, calling handler for each.
// Respects ctx cancellation. Returns the number of alerts dispatched.
func (r *Replayer) Run(ctx context.Context, handler Handler) (int, error) {
	events := r.history.Recent(r.history.Len())
	if len(events) == 0 {
		return 0, nil
	}

	var prev time.Time
	dispatched := 0

	for _, ev := range events {
		select {
		case <-ctx.Done():
			return dispatched, ctx.Err()
		default:
		}

		alert := eventToAlert(ev)

		if r.opts.Filter != nil && !r.opts.Filter(alert) {
			continue
		}

		if r.opts.Speed > 0 && !prev.IsZero() {
			gap := ev.At.Sub(prev)
			delay := time.Duration(float64(gap) / r.opts.Speed)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return dispatched, ctx.Err()
			}
		}

		if err := handler(alert); err != nil {
			return dispatched, err
		}

		prev = ev.At
		dispatched++
	}

	return dispatched, nil
}

func eventToAlert(ev snapshot.Event) alerting.Alert {
	level := alerting.LevelInfo
	if ev.Kind == snapshot.EventRemoved {
		level = alerting.LevelWarn
	}
	return alerting.Alert{
		Level:     level,
		Message:   ev.Kind.String() + ": " + ev.Entry.String(),
		Timestamp: ev.At,
		Tags:      map[string]string{"source": "replay"},
	}
}
