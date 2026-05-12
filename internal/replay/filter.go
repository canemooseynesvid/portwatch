package replay

import (
	"portwatch/internal/alerting"
	"portwatch/internal/snapshot"
	"strings"
	"time"
)

// TimeRange restricts replay to events within [From, To].
// A zero value means unbounded on that side.
type TimeRange struct {
	From time.Time
	To   time.Time
}

// Contains reports whether t falls within the range.
func (tr TimeRange) Contains(t time.Time) bool {
	if !tr.From.IsZero() && t.Before(tr.From) {
		return false
	}
	if !tr.To.IsZero() && t.After(tr.To) {
		return false
	}
	return true
}

// WithTimeRange returns a filter that passes only events within tr.
func WithTimeRange(tr TimeRange) func(alerting.Alert) bool {
	return func(a alerting.Alert) bool {
		return tr.Contains(a.Timestamp)
	}
}

// WithMinLevel returns a filter that passes only alerts at or above level.
func WithMinLevel(level alerting.Level) func(alerting.Alert) bool {
	return func(a alerting.Alert) bool {
		return a.Level >= level
	}
}

// WithKind returns a filter that passes only events matching kind.
func WithKind(kind snapshot.EventKind) func(alerting.Alert) bool {
	prefix := kind.String() + ":"
	return func(a alerting.Alert) bool {
		return strings.HasPrefix(a.Message, prefix)
	}
}

// CombineFilters returns a filter that passes only when all filters pass.
func CombineFilters(filters ...func(alerting.Alert) bool) func(alerting.Alert) bool {
	return func(a alerting.Alert) bool {
		for _, f := range filters {
			if !f(a) {
				return false
			}
		}
		return true
	}
}
