// Package envelope wraps an alert with routing metadata such as
// destination tags, priority hints, and a unique dispatch ID.
package envelope

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"portwatch/internal/alerting"
)

// Envelope wraps an alert with routing and dispatch metadata.
type Envelope struct {
	ID        string
	Alert     alerting.Alert
	Tags      []string
	Priority  int
	CreatedAt time.Time
}

// New wraps alert in an Envelope, assigning a unique ID and timestamp.
func New(a alerting.Alert, tags ...string) Envelope {
	return Envelope{
		ID:        uuid.NewString(),
		Alert:     a,
		Tags:      tags,
		Priority:  priorityFromLevel(a.Level),
		CreatedAt: time.Now(),
	}
}

// HasTag reports whether the envelope carries the given tag.
func (e Envelope) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// String returns a human-readable summary of the envelope.
func (e Envelope) String() string {
	return fmt.Sprintf("[%s] pri=%d tags=%v alert=%s",
		e.ID[:8], e.Priority, e.Tags, e.Alert)
}

// priorityFromLevel maps an alert level to a numeric priority (higher = more urgent).
func priorityFromLevel(l alerting.Level) int {
	switch l {
	case alerting.LevelCritical:
		return 4
	case alerting.LevelError:
		return 3
	case alerting.LevelWarning:
		return 2
	case alerting.LevelInfo:
		return 1
	default:
		return 0
	}
}
